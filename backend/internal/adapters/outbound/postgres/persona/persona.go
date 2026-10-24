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

	personas, _, err := scanPersonas(rows)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	if len(personas) == 0 {
		return nil, domain.ErrPersonaNotFound
	}

	return personas[0], nil
}

// ListPersonas implements ports.PersonaRepository.
// Names sort bytewise (collation C), like the in-memory repository. The query starts at the collector, so an unknown collector (no rows) is told apart from one without personas.
func (r *Repository) ListPersonas(ctx context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.collector_id, p.name, p.description,
		       i.id, i.title, i.kind, i.recurrence, i.document_spec, i.value_spec
		FROM collectors c
		LEFT JOIN personas p ON p.collector_id = c.id
		LEFT JOIN item_definitions i ON i.persona_id = p.id
		WHERE c.id = $1
		ORDER BY p.name COLLATE "C", p.id, i.position`, string(collectorID))
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}
	defer rows.Close()

	personas, found, err := scanPersonas(rows)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	if !found {
		return nil, domain.ErrCollectorNotFound
	}

	return personas, nil
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
// All statements go out as one batch, so it costs a single round trip however many items there are.
func saveItems(ctx context.Context, tx pgx.Tx, personaID uuid.UUID, items []itemRow) error {
	// Never nil: a nil slice is sent as NULL and "<> ALL(NULL)" would delete nothing.
	ids := make([]uuid.UUID, 0, len(items))
	batch := &pgx.Batch{}

	for position, item := range items {
		batch.Queue(`
			INSERT INTO item_definitions (id, persona_id, position, title, kind, recurrence, document_spec, value_spec)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				position = EXCLUDED.position, title = EXCLUDED.title, kind = EXCLUDED.kind,
				recurrence = EXCLUDED.recurrence, document_spec = EXCLUDED.document_spec, value_spec = EXCLUDED.value_spec
			WHERE item_definitions.persona_id = EXCLUDED.persona_id`,
			item.ID, personaID, position, item.Title, item.Kind, item.Recurrence, item.DocumentSpec, item.ValueSpec)

		ids = append(ids, item.ID)
	}

	batch.Queue(`DELETE FROM item_definitions WHERE persona_id = $1 AND id <> ALL($2::uuid[])`, personaID, ids)

	results := tx.SendBatch(ctx, batch)

	for _, item := range items {
		tag, err := results.Exec()
		if err != nil {
			_ = results.Close()

			return fmt.Errorf("save item %s: %w", item.ID, err)
		}

		// The statement inserts a new id or updates an item of this persona, both affect one row. Zero rows
		// means the id exists under another persona and the WHERE above skipped the update. Keep the
		// statement and this check in sync: any other way to affect no rows would be misreported here.
		if tag.RowsAffected() == 0 {
			_ = results.Close()

			return fmt.Errorf("save item %s: %w", item.ID, domain.ErrItemIDTaken)
		}
	}

	if _, err := results.Exec(); err != nil {
		_ = results.Close()

		return fmt.Errorf("delete removed items: %w", err)
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("save items: %w", err)
	}

	return nil
}

// scanPersonas reads rows of persona and item columns ordered by persona. It reports whether there was any row
// at all: a LEFT JOIN from collectors yields one row with NULL persona columns for a collector without personas.
func scanPersonas(rows pgx.Rows) ([]*domain.Persona, bool, error) {
	personas := make([]*domain.Persona, 0)
	sawRow := false

	var current *domain.Persona

	for rows.Next() {
		sawRow = true

		var (
			p    nullablePersonaRow
			item nullableItemRow
		)

		if err := rows.Scan(&p.ID, &p.CollectorID, &p.Name, &p.Description,
			&item.ID, &item.Title, &item.Kind, &item.Recurrence, &item.DocumentSpec, &item.ValueSpec); err != nil {
			return nil, false, fmt.Errorf("scan: %w", err)
		}

		if p.ID == nil {
			continue
		}

		if current == nil || current.ID != *p.ID {
			current = p.toRow().toDomain()
			personas = append(personas, current)
		}

		// A persona without items yields one row with NULL item columns.
		if item.ID == nil {
			continue
		}

		domainItem, err := item.toRow().toDomain(current.ID)
		if err != nil {
			return nil, false, err
		}

		current.Items = append(current.Items, domainItem)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("read rows: %w", err)
	}

	return personas, sawRow, nil
}

// nullablePersonaRow holds the persona columns of an outer join, which are NULL if there is no persona.
// ID is nil in that case, the other fields are only valid if it is not.
type nullablePersonaRow struct {
	ID          *uuid.UUID
	CollectorID *string
	Name        *string
	Description *string
}

func (r nullablePersonaRow) toRow() personaRow {
	return personaRow{ID: *r.ID, CollectorID: *r.CollectorID, Name: *r.Name, Description: *r.Description}
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
