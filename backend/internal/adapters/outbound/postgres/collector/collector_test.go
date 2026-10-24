package collector_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/postgrestest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/porttest"
)

func TestCollectorRepository(t *testing.T) {
	pool := postgrestest.RequirePool(t)

	porttest.TestCollectorRepository(t, func(t *testing.T) port.CollectorRepository {
		if _, err := pool.Exec(context.Background(), `TRUNCATE collectors CASCADE`); err != nil {
			t.Fatalf("truncate collectors: %v", err)
		}

		return collector.NewRepository(pool)
	})
}

// If saving a persona fails inside the transaction, the collector saved before it must not stay behind.
// The failure is a real one: Postgres rejects an item id that another persona already uses.
func TestSaveCollector_RolledBackWhenPersonaSaveFails(t *testing.T) {
	ctx := context.Background()
	pool := postgrestest.RequirePool(t)

	if _, err := pool.Exec(ctx, `TRUNCATE personas, collectors CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	// bbbb owns a persona with an item, whose id aaaa's persona then tries to reuse.
	if err := collector.NewRepository(pool).SaveCollector(ctx, &domain.Collector{ID: "bbbb", Name: "B", Email: "b@example.com"}); err != nil {
		t.Fatalf("seed collector: %v", err)
	}

	existing, err := domain.NewPersona(domain.PersonaID(uuid.New()), "bbbb", "Employed", "")
	if err != nil {
		t.Fatalf("new persona: %v", err)
	}

	item, err := existing.AddItem(domain.ItemID(uuid.New()), "Days", domain.RecurrenceAnnually, domain.ValueSpec{Type: domain.ValueTypeNumber})
	if err != nil {
		t.Fatalf("add item: %v", err)
	}

	if err := persona.NewRepository(pool).SavePersona(ctx, existing); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	clashing, err := domain.NewPersona(domain.PersonaID(uuid.New()), "aaaa", "Employed", "")
	if err != nil {
		t.Fatalf("new persona: %v", err)
	}

	clashing.Items = []domain.ItemDefinition{item}

	tx := pgdb.NewTransactor(pool, func(db pgdb.DBTX) port.Repositories {
		return port.Repositories{Collectors: collector.NewRepository(db), Personas: persona.NewRepository(db)}
	})

	err = tx.InTx(ctx, func(ctx context.Context, repos port.Repositories) error {
		if err := repos.Collectors.SaveCollector(ctx, &domain.Collector{ID: "aaaa", Name: "A", Email: "a@example.com"}); err != nil {
			return err
		}

		return repos.Personas.SavePersona(ctx, clashing)
	})
	if !errors.Is(err, domain.ErrItemIDTaken) {
		t.Fatalf("err = %v, want %v", err, domain.ErrItemIDTaken)
	}

	if _, err := collector.NewRepository(pool).GetCollector(ctx, "aaaa"); !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("GetCollector(aaaa) err = %v, want %v: the collector was not rolled back", err, domain.ErrCollectorNotFound)
	}

	if _, err := persona.NewRepository(pool).GetPersona(ctx, "aaaa", clashing.ID); !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Errorf("GetPersona err = %v, want %v", err, domain.ErrPersonaNotFound)
	}

	// The seeded data of the other collector is untouched.
	got, err := persona.NewRepository(pool).GetPersona(ctx, "bbbb", existing.ID)
	if err != nil {
		t.Fatalf("seeded persona lost: %v", err)
	}

	if len(got.Items) != 1 || got.Items[0].ID != item.ID {
		t.Errorf("seeded persona items = %+v, want the original item", got.Items)
	}
}
