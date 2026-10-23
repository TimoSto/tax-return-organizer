// Package persona contains the PostgreSQL implementation of the persona repository.
package persona

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
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
func (r *Repository) GetPersona(ctx context.Context, collectorID domain.CollectorID, id uuid.UUID) (*domain.Persona, error) {
	var row personaRow

	err := r.pool.QueryRow(ctx,
		`SELECT id, collector_id, name, description FROM personas WHERE id = $1 AND collector_id = $2`,
		id, string(collectorID)).
		Scan(&row.ID, &row.CollectorID, &row.Name, &row.Description)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPersonaNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	persona := row.toDomain()

	persona.Items, err = r.getItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	return persona, nil
}

func (r *Repository) getItems(ctx context.Context, personaID uuid.UUID) ([]domain.ItemDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, kind, recurrence, document_spec, value_spec
		FROM item_definitions WHERE persona_id = $1 ORDER BY position`, personaID)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	var items []domain.ItemDefinition

	for rows.Next() {
		var row itemRow
		if err := rows.Scan(&row.ID, &row.Title, &row.Kind, &row.Recurrence, &row.DocumentSpec, &row.ValueSpec); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		item, err := row.toDomain(personaID)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read items: %w", err)
	}

	return items, nil
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

		if tag.RowsAffected() == 0 {
			return fmt.Errorf("save item %s: id belongs to another persona", item.ID)
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
