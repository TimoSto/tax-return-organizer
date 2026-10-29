// Package persona contains the PostgreSQL implementation of the persona repository.
package persona

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

const (
	// uniqueViolation is the SQLSTATE of a unique or primary key violation.
	uniqueViolation = "23505"
	// foreignKeyViolation is the SQLSTATE of a foreign key violation.
	foreignKeyViolation = "23503"
)

var (
	_ outbound.PersonaStore      = (*Store)(nil)
	_ outbound.PersonaRepository = (*Repository)(nil)
)

// Store is the Postgres implementation of outbound.PersonaStore.
type Store struct {
	db pgdb.DBTX
}

// NewStore returns a store backed by db, the pool or a transaction.
func NewStore(db pgdb.DBTX) *Store {
	return &Store{db: db}
}

// For implements outbound.PersonaStore. The repository is a view on the same db, nothing is allocated per collector.
func (s *Store) For(collectorID domain.CollectorID) outbound.PersonaRepository {
	return &Repository{db: s.db, collectorID: collectorID}
}

// Repository is the Postgres implementation of outbound.PersonaRepository for one collector.
type Repository struct {
	db          pgdb.DBTX
	collectorID domain.CollectorID
}

// GetPersona implements outbound.PersonaRepository.
// Collector, persona and items are read in one statement, so a concurrent SavePersona can never show up
// half-applied. The query starts from the collector, so an unknown collector yields no row at all.
func (r *Repository) GetPersona(ctx context.Context, id domain.PersonaID) (*domain.Persona, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.collector_id, p.name, p.description,
		       i.id, i.title, i.kind, i.recurrence, i.document_spec, i.value_spec
		FROM collectors c
		LEFT JOIN personas p ON p.collector_id = c.id AND p.id = $1
		LEFT JOIN item_definitions i ON i.persona_id = p.id
		WHERE c.id = $2
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

// ListPersonas implements outbound.PersonaRepository.
// Names sort bytewise (collation C), like the in-memory repository. The query starts from the collector, so a
// collector without personas yields one row with NULL persona columns and an unknown collector no row.
func (r *Repository) ListPersonas(ctx context.Context) ([]*domain.Persona, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.collector_id, p.name, p.description,
		       i.id, i.title, i.kind, i.recurrence, i.document_spec, i.value_spec
		FROM collectors c
		LEFT JOIN personas p ON p.collector_id = c.id
		LEFT JOIN item_definitions i ON i.persona_id = p.id
		WHERE c.id = $1
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

// SavePersona implements outbound.PersonaRepository. The persona and its items go out as one batch, which Postgres
// runs as a single implicit transaction: either everything is written or nothing is. The repository opens no
// transaction itself, callers that need more than this one save to be atomic wrap it in outbound.Transactor.
func (r *Repository) SavePersona(ctx context.Context, persona *domain.Persona) error {
	if persona.CollectorID() != r.collectorID {
		return fmt.Errorf("save persona %s: %w", persona.ID(), outbound.ErrTenantMismatch)
	}

	items := make([]itemRow, 0, len(persona.Items()))

	for _, item := range persona.Items() {
		row, err := newItemRow(item)
		if err != nil {
			return err
		}

		items = append(items, row)
	}

	row := newPersonaRow(persona)

	// Never nil: a nil slice is sent as NULL and "<> ALL(NULL)" would delete nothing.
	ids := make([]uuid.UUID, 0, len(items))
	batch := &pgx.Batch{}

	// The persona is upserted first. The collector of an existing persona never changes. If the id belongs to
	// another collector this affects no row, and the guards below keep the item statements from touching the
	// foreign persona.
	batch.Queue(`
		INSERT INTO personas (id, collector_id, name, description) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description
		WHERE personas.collector_id = EXCLUDED.collector_id`,
		row.ID, row.CollectorID, row.Name, row.Description)

	// An unknown collector makes the upsert fail on the foreign key, unless the persona id exists already: then
	// the insert turns into the guarded update, which affects no row and checks no foreign key. This statement
	// tells the two causes of "no row" apart, so the collector takes precedence like in the reads.
	batch.Queue(`SELECT EXISTS (SELECT 1 FROM collectors WHERE id = $1)`, row.CollectorID)

	// Items keep their identity: an item of this persona is updated, any other id is inserted. The insert is
	// plain on purpose. An id that exists under another persona makes it fail with a unique violation, which
	// aborts the whole batch, so nothing stays written (a statement that merely affects no row could not undo
	// the persona upsert before it). The EXISTS guard turns the statement into a no-op if the persona is not
	// this collector's, the persona upsert has reported that already.
	for position, item := range items {
		batch.Queue(`
			WITH updated AS (
				UPDATE item_definitions SET position = $3, title = $4, kind = $5, recurrence = $6,
					document_spec = $7, value_spec = $8
				WHERE id = $1 AND persona_id = $2
				RETURNING 1)
			INSERT INTO item_definitions (id, persona_id, position, title, kind, recurrence, document_spec, value_spec)
			SELECT $1, $2, $3, $4, $5, $6, $7, $8
			WHERE NOT EXISTS (SELECT 1 FROM updated)
			  AND EXISTS (SELECT 1 FROM personas WHERE id = $2 AND collector_id = $9)`,
			item.ID, row.ID, position, item.Title, item.Kind, item.Recurrence, item.DocumentSpec, item.ValueSpec,
			row.CollectorID)

		ids = append(ids, item.ID)
	}

	batch.Queue(`
		DELETE FROM item_definitions
		WHERE persona_id = $1 AND id <> ALL($2::uuid[])
		  AND EXISTS (SELECT 1 FROM personas WHERE id = $1 AND collector_id = $3)`,
		row.ID, ids, row.CollectorID)

	results := r.db.SendBatch(ctx, batch)

	if err := readSaveResults(results, items); err != nil {
		_ = results.Close()

		return err
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("save persona: %w", err)
	}

	return nil
}

// readSaveResults checks the result of every statement queued by SavePersona, in order. An error aborts the
// implicit transaction, so returning early leaves nothing written. Returning early because the upsert affected
// no row is safe too: the guards of the item statements turn them into no-ops.
func readSaveResults(results pgx.BatchResults, items []itemRow) error {
	tag, err := results.Exec()
	if err != nil {
		if isCollectorMissing(err) {
			return domain.ErrCollectorNotFound
		}

		return fmt.Errorf("upsert persona: %w", err)
	}

	var collectorExists bool
	if err := results.QueryRow().Scan(&collectorExists); err != nil {
		return fmt.Errorf("check collector: %w", err)
	}

	if tag.RowsAffected() == 0 {
		if !collectorExists {
			return domain.ErrCollectorNotFound
		}

		return domain.ErrPersonaNotFound
	}

	for _, item := range items {
		if _, err := results.Exec(); err != nil {
			if isItemIDTaken(err) {
				return fmt.Errorf("save item %s: %w", item.ID, domain.ErrItemIDTaken)
			}

			return fmt.Errorf("save item %s: %w", item.ID, err)
		}
	}

	if _, err := results.Exec(); err != nil {
		return fmt.Errorf("delete removed items: %w", err)
	}

	return nil
}

// isItemIDTaken reports whether err is the primary key violation of an item id that exists under another persona.
func isItemIDTaken(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "item_definitions_pkey"
}

// isCollectorMissing reports whether err is the foreign key violation of a persona whose collector does not exist.
func isCollectorMissing(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation && pgErr.ConstraintName == "personas_collector_id_fkey"
}

// scanPersonas reads rows of persona and item columns ordered by persona, as selected from collectors LEFT JOIN
// personas. It returns domain.ErrCollectorNotFound if there is no row, so the collector does not exist. A row with
// NULL persona columns only shows that the collector exists.
func scanPersonas(rows pgx.Rows) ([]*domain.Persona, error) {
	personas := make([]*domain.Persona, 0)
	collectorFound := false

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
			np   nullablePersonaRow
			item nullableItemRow
		)

		if err := rows.Scan(&np.ID, &np.CollectorID, &np.Name, &np.Description,
			&item.ID, &item.Title, &item.Kind, &item.Recurrence, &item.DocumentSpec, &item.ValueSpec); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		collectorFound = true

		// The collector has no (matching) persona.
		if np.ID == nil {
			continue
		}

		p := np.toRow()

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

	if !collectorFound {
		return nil, domain.ErrCollectorNotFound
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

// nullablePersonaRow scans the persona columns of a LEFT JOIN, which are all NULL for a collector without
// (matching) personas. Only ID is checked for NULL, the other columns are NOT NULL once a persona exists.
type nullablePersonaRow struct {
	ID          *uuid.UUID
	CollectorID *string
	Name        *string
	Description *string
}

// toRow must only be called when ID is not nil.
func (r nullablePersonaRow) toRow() personaRow {
	return personaRow{ID: *r.ID, CollectorID: *r.CollectorID, Name: *r.Name, Description: *r.Description}
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
		return nil, outbound.CorruptData("persona "+r.ID.String(), err)
	}

	return p, nil
}
