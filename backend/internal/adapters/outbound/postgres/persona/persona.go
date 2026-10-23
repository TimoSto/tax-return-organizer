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
		SELECT p.id, p.collector_id, p.name, p.description,
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

		if err := rows.Scan(&p.ID, &p.CollectorID, &p.Name, &p.Description,
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

	if err := savePersonaRow(ctx, tx, newPersonaRow(persona)); err != nil {
		return err
	}

	if err := saveItems(ctx, tx, persona.ID, items); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("save persona: commit: %w", err)
	}

	return nil
}

// savePersonaRow upserts the persona. The collector of an existing persona never changes.
func savePersonaRow(ctx context.Context, tx pgx.Tx, row personaRow) error {
	tag, err := tx.Exec(ctx, `
		INSERT INTO personas (id, collector_id, name, description) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description
		WHERE personas.collector_id = EXCLUDED.collector_id`,
		row.ID, row.CollectorID, row.Name, row.Description)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation && pgErr.ConstraintName == collectorFKName {
		return fmt.Errorf("save persona: %w", domain.ErrUnknownCollector)
	}

	if err != nil {
		return fmt.Errorf("save persona: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrPersonaNotFound
	}

	return nil
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
}

func newPersonaRow(p *domain.Persona) personaRow {
	return personaRow{ID: p.ID, CollectorID: string(p.CollectorID), Name: p.Name, Description: p.Description}
}

func (r personaRow) toDomain() *domain.Persona {
	return &domain.Persona{
		ID:          r.ID,
		CollectorID: domain.CollectorID(r.CollectorID),
		Name:        r.Name,
		Description: r.Description,
	}
}
