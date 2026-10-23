// Package persona contains the PostgreSQL implementation of the persona repository.
package persona

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

const (
	foreignKeyViolation = "23503"
	collectorFKName     = "personas_collector_id_fkey"
)

var _ ports.PersonaRepository = (*Repository)(nil)

// Repository is the Postgres implementation of ports.PersonaRepository.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a repository backed by pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetPersona implements ports.PersonaRepository.
// Persona and items are read in one statement, so a concurrent SavePersona can never show up half-applied.
func (r *Repository) GetPersona(ctx context.Context, collectorID domain.CollectorID, id uuid.UUID) (*domain.Persona, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.collector_id, p.name, p.description, p.version,
		       i.id, i.title, i.kind, i.recurrence, i.document_spec, i.value_spec
		FROM personas p
		LEFT JOIN item_definitions i ON i.persona_id = p.id
		WHERE p.id = $1 AND p.collector_id = $2
		ORDER BY i.position`, id, string(collectorID))
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}
	defer rows.Close()

	var persona *domain.Persona

	for rows.Next() {
		var (
			p    personaRow
			item nullableItemRow
		)

		if err := rows.Scan(&p.ID, &p.CollectorID, &p.Name, &p.Description, &p.Version,
			&item.ID, &item.Title, &item.Kind, &item.Recurrence, &item.DocumentSpec, &item.ValueSpec); err != nil {
			return nil, fmt.Errorf("get persona: scan: %w", err)
		}

		if persona == nil {
			persona = p.toDomain()
		}

		// A persona without items yields one row with NULL item columns.
		if item.ID == nil {
			continue
		}

		domainItem, err := item.toRow().toDomain(persona.ID)
		if err != nil {
			return nil, fmt.Errorf("get persona: %w", err)
		}

		persona.Items = append(persona.Items, domainItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get persona: read rows: %w", err)
	}

	if persona == nil {
		return nil, domain.ErrPersonaNotFound
	}

	return persona, nil
}

// SavePersona implements ports.PersonaRepository. Persona and items are written in one transaction.
func (r *Repository) SavePersona(ctx context.Context, persona *domain.Persona) error {
	items := make([]itemRow, 0, len(persona.Items))

	for _, item := range persona.Items {
		row, err := newItemRow(item)
		if err != nil {
			return fmt.Errorf("save persona: %w", err)
		}

		items = append(items, row)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("save persona: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	version, err := savePersonaRow(ctx, tx, newPersonaRow(persona))
	if err != nil {
		return err
	}

	if err := saveItems(ctx, tx, persona.ID, items); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("save persona: commit: %w", err)
	}

	persona.Version = version

	return nil
}

// savePersonaRow creates the persona (version 0) or updates it if the stored version still matches.
// The collector of an existing persona never changes. It returns the new stored version.
func savePersonaRow(ctx context.Context, tx pgx.Tx, row personaRow) (int, error) {
	if row.Version == 0 {
		tag, err := tx.Exec(ctx, `
			INSERT INTO personas (id, collector_id, name, description) VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING`,
			row.ID, row.CollectorID, row.Name, row.Description)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation && pgErr.ConstraintName == collectorFKName {
			return 0, fmt.Errorf("save persona: %w", domain.ErrUnknownCollector)
		}

		if err != nil {
			return 0, fmt.Errorf("save persona: %w", err)
		}

		// No row means the id exists already: either a conflict or someone else's persona.
		if tag.RowsAffected() == 0 {
			return 0, existingPersonaError(ctx, tx, row)
		}

		return 1, nil
	}

	var version int

	err := tx.QueryRow(ctx, `
		UPDATE personas SET name = $3, description = $4, version = version + 1
		WHERE id = $1 AND collector_id = $2 AND version = $5
		RETURNING version`,
		row.ID, row.CollectorID, row.Name, row.Description, row.Version).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, existingPersonaError(ctx, tx, row)
	}

	if err != nil {
		return 0, fmt.Errorf("save persona: %w", err)
	}

	return version, nil
}

// existingPersonaError explains why a save did not touch the persona row.
func existingPersonaError(ctx context.Context, tx pgx.Tx, row personaRow) error {
	var collectorID string

	err := tx.QueryRow(ctx, `SELECT collector_id FROM personas WHERE id = $1`, row.ID).Scan(&collectorID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && collectorID != row.CollectorID) {
		return domain.ErrPersonaNotFound
	}

	if err != nil {
		return fmt.Errorf("save persona: %w", err)
	}

	return domain.ErrVersionConflict
}

// saveItems upserts the items by id, so they keep their identity, and deletes the ones that are gone.
func saveItems(ctx context.Context, tx pgx.Tx, personaID uuid.UUID, items []itemRow) error {
	ids := make([]string, 0, len(items))

	for position, item := range items {
		tag, err := tx.Exec(ctx, `
			INSERT INTO item_definitions (id, persona_id, position, title, kind, recurrence, document_spec, value_spec)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				position = EXCLUDED.position, title = EXCLUDED.title, kind = EXCLUDED.kind,
				recurrence = EXCLUDED.recurrence, document_spec = EXCLUDED.document_spec, value_spec = EXCLUDED.value_spec
			WHERE item_definitions.persona_id = EXCLUDED.persona_id`,
			item.ID, personaID, position, item.Title, item.Kind, item.Recurrence, item.DocumentSpec, item.ValueSpec)
		if err != nil {
			return fmt.Errorf("save item %s: %w", item.ID, err)
		}

		// The statement inserts a new id or updates an item of this persona, both affect one row. Zero rows
		// means the id exists under another persona and the WHERE above skipped the update. Keep the
		// statement and this check in sync: any other way to affect no rows would be misreported here.
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("save item %s: %w", item.ID, domain.ErrItemIDTaken)
		}

		ids = append(ids, item.ID.String())
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM item_definitions WHERE persona_id = $1 AND id::text <> ALL($2)`, personaID, ids); err != nil {
		return fmt.Errorf("delete removed items: %w", err)
	}

	return nil
}

// personaRow mirrors the personas table and keeps SQL concerns out of the domain type.
type personaRow struct {
	ID          uuid.UUID
	CollectorID string
	Name        string
	Description string
	Version     int
}

func newPersonaRow(p *domain.Persona) personaRow {
	return personaRow{ID: p.ID, CollectorID: string(p.CollectorID), Name: p.Name, Description: p.Description, Version: p.Version}
}

func (r personaRow) toDomain() *domain.Persona {
	return &domain.Persona{
		ID:          r.ID,
		CollectorID: domain.CollectorID(r.CollectorID),
		Name:        r.Name,
		Description: r.Description,
		Version:     r.Version,
	}
}
