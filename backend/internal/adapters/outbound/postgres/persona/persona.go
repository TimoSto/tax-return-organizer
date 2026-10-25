// Package persona contains the PostgreSQL implementation of the persona repository.
package persona

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

var (
	_ port.PersonaStore      = (*Store)(nil)
	_ port.PersonaRepository = (*Repository)(nil)
)

// Store is the Postgres implementation of port.PersonaStore.
type Store struct {
	db pgdb.DBTX
}

// NewStore returns a store backed by db, the pool or a transaction.
func NewStore(db pgdb.DBTX) *Store {
	return &Store{db: db}
}

// For implements port.PersonaStore. The repository is a view on the same db, nothing is allocated per collector.
func (s *Store) For(collectorID domain.CollectorID) port.PersonaRepository {
	return &Repository{db: s.db, collectorID: collectorID}
}

// Repository is the Postgres implementation of port.PersonaRepository for one collector.
type Repository struct {
	db          pgdb.DBTX
	collectorID domain.CollectorID
}

// GetPersona implements port.PersonaRepository.
// Persona and items are read in one statement, so a concurrent SavePersona can never show up half-applied.
func (r *Repository) GetPersona(ctx context.Context, id domain.PersonaID) (*domain.Persona, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.collector_id, p.name, p.description,
		       i.id, i.title, i.kind, i.recurrence, i.document_spec, i.value_spec
		FROM personas p
		LEFT JOIN item_definitions i ON i.persona_id = p.id
		WHERE p.id = $1 AND p.collector_id = $2
		ORDER BY i.position`, uuid.UUID(id), string(r.collectorID))
	if err != nil {
		return nil, fmt.Errorf("query persona: %w", err)
	}
	defer rows.Close()

	personas, err := scanPersonas(rows)
	if err != nil {
		return nil, err
	}

	if len(personas) == 0 {
		return nil, domain.ErrPersonaNotFound
	}

	return personas[0], nil
}

// ListPersonas implements port.PersonaRepository.
// Names sort bytewise (collation C), like the in-memory repository. An unknown collector yields an empty list, callers check the collector.
func (r *Repository) ListPersonas(ctx context.Context) ([]*domain.Persona, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.collector_id, p.name, p.description,
		       i.id, i.title, i.kind, i.recurrence, i.document_spec, i.value_spec
		FROM personas p
		LEFT JOIN item_definitions i ON i.persona_id = p.id
		WHERE p.collector_id = $1
		ORDER BY p.name COLLATE "C", p.id, i.position`, string(r.collectorID))
	if err != nil {
		return nil, fmt.Errorf("query personas: %w", err)
	}
	defer rows.Close()

	personas, err := scanPersonas(rows)
	if err != nil {
		return nil, err
	}

	return personas, nil
}

// SavePersona implements port.PersonaRepository. Persona and items are written in one transaction.
func (r *Repository) SavePersona(ctx context.Context, persona *domain.Persona) error {
	if persona.CollectorID() != r.collectorID {
		return fmt.Errorf("save persona %s: %w", persona.ID(), port.ErrTenantMismatch)
	}

	items := make([]itemRow, 0, len(persona.Items()))

	for _, item := range persona.Items() {
		row, err := newItemRow(item)
		if err != nil {
			return err
		}

		items = append(items, row)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := savePersonaRow(ctx, tx, newPersonaRow(persona)); err != nil {
		return err
	}

	if err := saveItems(ctx, tx, uuid.UUID(persona.ID()), items); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
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
		return fmt.Errorf("upsert persona: %w", err)
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

// scanPersonas reads rows of persona and item columns ordered by persona.
func scanPersonas(rows pgx.Rows) ([]*domain.Persona, error) {
	personas := make([]*domain.Persona, 0)

	var (
		current *personaRow
		items   []domain.ItemDefinition
	)

	// flush builds the persona collected so far. A persona is only complete once all its rows have been read.
	flush := func() error {
		if current == nil {
			return nil
		}

		persona, err := current.toDomain(items)
		if err != nil {
			return err
		}

		personas = append(personas, persona)

		return nil
	}

	for rows.Next() {
		var (
			p    personaRow
			item nullableItemRow
		)

		if err := rows.Scan(&p.ID, &p.CollectorID, &p.Name, &p.Description,
			&item.ID, &item.Title, &item.Kind, &item.Recurrence, &item.DocumentSpec, &item.ValueSpec); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		if current == nil || current.ID != p.ID {
			if err := flush(); err != nil {
				return nil, err
			}

			row := p
			current, items = &row, nil
		}

		// A persona without items yields one row with NULL item columns.
		if item.ID == nil {
			continue
		}

		domainItem, err := item.toRow().toDomain()
		if err != nil {
			return nil, err
		}

		items = append(items, domainItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}

	if err := flush(); err != nil {
		return nil, err
	}

	return personas, nil
}

// personaRow mirrors the personas table and keeps SQL concerns out of the domain type.
type personaRow struct {
	ID          uuid.UUID
	CollectorID string
	Name        string
	Description string
}

func newPersonaRow(p *domain.Persona) personaRow {
	return personaRow{
		ID: uuid.UUID(p.ID()), CollectorID: string(p.CollectorID()), Name: p.Name(), Description: p.Description(),
	}
}

func (r personaRow) toDomain(items []domain.ItemDefinition) (*domain.Persona, error) {
	p, err := domain.RehydratePersona(
		domain.PersonaID(r.ID), domain.CollectorID(r.CollectorID), r.Name, r.Description, items)
	if err != nil {
		return nil, port.CorruptData("persona "+r.ID.String(), err)
	}

	return p, nil
}
