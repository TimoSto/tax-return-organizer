package collector_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/porttest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/postgrestest"
)

func TestCollectorRepository(t *testing.T) {
	pool := postgrestest.RequirePool(t)

	porttest.TestCollectorRepository(t, func(t *testing.T) outbound.CollectorRepository {
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
	if err := collector.NewRepository(pool).SaveCollector(ctx, domaintest.Collector("bbbb", "B", "b@example.com")); err != nil {
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

	if err := persona.NewStore(pool).For("bbbb").SavePersona(ctx, existing); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	clashing := domaintest.Persona(domaintest.PersonaID(), "aaaa", "Employed", "", item)

	tx := pgdb.NewTransactor(pool, func(db pgdb.DBTX) outbound.Repositories {
		return outbound.Repositories{Collectors: collector.NewRepository(db), Personas: persona.NewStore(db)}
	})

	err = tx.InTx(ctx, func(ctx context.Context, repos outbound.Repositories) error {
		if err := repos.Collectors.SaveCollector(ctx, domaintest.Collector("aaaa", "A", "a@example.com")); err != nil {
			return err
		}

		return repos.Personas.For("aaaa").SavePersona(ctx, clashing)
	})
	if !errors.Is(err, domain.ErrItemIDTaken) {
		t.Fatalf("err = %v, want %v", err, domain.ErrItemIDTaken)
	}

	if _, err := collector.NewRepository(pool).GetCollector(ctx, "aaaa"); !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("GetCollector(aaaa) err = %v, want %v: the collector was not rolled back", err, domain.ErrCollectorNotFound)
	}

	// With the collector rolled back, the store reports it missing before looking for the persona.
	if _, err := persona.NewStore(pool).For("aaaa").GetPersona(ctx, clashing.ID()); !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("GetPersona err = %v, want %v", err, domain.ErrCollectorNotFound)
	}

	// The seeded data of the other collector is untouched.
	got, err := persona.NewStore(pool).For("bbbb").GetPersona(ctx, existing.ID())
	if err != nil {
		t.Fatalf("seeded persona lost: %v", err)
	}

	if items := got.Items(); len(items) != 1 || items[0].ID() != item.ID() {
		t.Errorf("seeded persona items = %+v, want the original item", items)
	}
}
