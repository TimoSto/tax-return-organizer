package persona_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/porttest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/postgrestest"
)

var (
	// testCollectorID is the collector the tests act for.
	testCollectorID = domaintest.CollectorID()
)

func newStore(pool *pgxpool.Pool) porttest.PersonaStoreFactory {
	return func(t *testing.T, collectors ...domain.CollectorID) outbound.PersonaStore {
		ctx := context.Background()

		// personas reference collectors, so both tables are reset and the collectors re-seeded.
		if _, err := pool.Exec(ctx, `TRUNCATE personas, collectors CASCADE`); err != nil {
			t.Fatalf("truncate: %v", err)
		}

		for _, id := range collectors {
			if _, err := pool.Exec(ctx,
				`INSERT INTO collectors (id, name, email) VALUES ($1, $2, $3)`,
				uuid.UUID(id), "Collector "+id.String(), id.String()+"@example.com"); err != nil {
				t.Fatalf("seed collector %s: %v", id, err)
			}
		}

		return persona.NewStore(pool)
	}
}

func TestPersonaRepository(t *testing.T) {
	porttest.TestPersonaRepository(t, newStore(postgrestest.RequirePool(t)))
}

func TestPersonaRepositoryList(t *testing.T) {
	porttest.TestPersonaRepositoryList(t, newStore(postgrestest.RequirePool(t)))
}

func TestPersonaRepositoryIsolation(t *testing.T) {
	porttest.TestPersonaRepositoryIsolation(t, newStore(postgrestest.RequirePool(t)))
}

func TestPersonaRepositoryUnknownCollector(t *testing.T) {
	porttest.TestPersonaRepositoryUnknownCollector(t, newStore(postgrestest.RequirePool(t)))
}

// A JSON null literal passes the "spec matches kind" check, so only the object check can reject it.
func TestItemDefinitions_SpecsMustBeJSONObjects(t *testing.T) {
	pool := postgrestest.RequirePool(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `TRUNCATE personas, collectors CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	personaID := uuid.New()

	if _, err := pool.Exec(ctx, `INSERT INTO collectors (id, name, email) VALUES ($1, 'A', 'a@example.com')`, uuid.UUID(testCollectorID)); err != nil {
		t.Fatalf("seed collector: %v", err)
	}

	if _, err := pool.Exec(ctx, `INSERT INTO personas (id, collector_id, name) VALUES ($1, $2, 'Employed')`, personaID, uuid.UUID(testCollectorID)); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	for _, tc := range []struct {
		name, kind, column, value string
		wantErr                   bool
	}{
		{"document spec object", "document", "document_spec", `{}`, false},
		{"document spec json null", "document", "document_spec", `null`, true},
		{"document spec array", "document", "document_spec", `[]`, true},
		{"value spec object", "data", "value_spec", `{}`, false},
		{"value spec json null", "data", "value_spec", `null`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pool.Exec(ctx,
				`INSERT INTO item_definitions (id, persona_id, collector_id, position, title, kind, recurrence, `+tc.column+`)
				 VALUES ($1, $2, $3, 0, 'Item', $4, 'monthly', $5::jsonb)`,
				uuid.New(), personaID, uuid.UUID(testCollectorID), tc.kind, tc.value)

			if (err != nil) != tc.wantErr {
				t.Errorf("insert error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// The composite foreign key ties an item to the collector of its persona, independent of the repository's SQL.
func TestItemDefinitions_CollectorMustMatchPersona(t *testing.T) {
	pool := postgrestest.RequirePool(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `TRUNCATE personas, collectors CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	alice, bob, personaID := uuid.New(), uuid.New(), uuid.New()

	for _, id := range []uuid.UUID{alice, bob} {
		if _, err := pool.Exec(ctx, `INSERT INTO collectors (id, name, email) VALUES ($1, 'C', 'c@example.com')`, id); err != nil {
			t.Fatalf("seed collector: %v", err)
		}
	}

	if _, err := pool.Exec(ctx, `INSERT INTO personas (id, collector_id, name) VALUES ($1, $2, 'Employed')`, personaID, alice); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	for _, tc := range []struct {
		name      string
		collector uuid.UUID
		wantErr   bool
	}{
		{"collector of the persona", alice, false},
		{"another collector", bob, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pool.Exec(ctx,
				`INSERT INTO item_definitions (id, persona_id, collector_id, position, title, kind, recurrence, document_spec)
				 VALUES ($1, $2, $3, 0, 'Item', 'document', 'monthly', '{}')`,
				uuid.New(), personaID, tc.collector)

			if !tc.wantErr {
				if err != nil {
					t.Errorf("insert error = %v, want nil", err)
				}

				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "23503" || pgErr.ConstraintName != "item_definitions_persona_fkey" {
				t.Errorf("insert error = %v, want a violation of item_definitions_persona_fkey", err)
			}
		})
	}
}
