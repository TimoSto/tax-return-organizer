package collector_test

import (
	"context"
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/porttest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/postgrestest"
)

var (
	// testCollectorID is the collector the tests act for.
	testCollectorID = domaintest.CollectorID()
	// otherCollectorID is another collector, which owns data the tests must not see.
	otherCollectorID = domaintest.CollectorID()
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
// The failure is a real one: the persona id already belongs to another collector, so Postgres updates no row.
func TestSaveCollector_RolledBackWhenPersonaSaveFails(t *testing.T) {
	ctx := context.Background()
	pool := postgrestest.RequirePool(t)

	if _, err := pool.Exec(ctx, `TRUNCATE personas, collectors CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	// otherCollectorID owns a persona, whose id testCollectorID's persona then tries to reuse.
	if err := collector.NewRepository(pool).SaveCollector(ctx, domaintest.Collector(otherCollectorID, "B", "b@example.com")); err != nil {
		t.Fatalf("seed collector: %v", err)
	}

	existing := domaintest.Persona(domaintest.PersonaID(), otherCollectorID, "Employed", "")

	if err := persona.NewStore(pool).For(otherCollectorID).SavePersona(ctx, existing); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	clashing := domaintest.Persona(existing.ID(), testCollectorID, "Employed", "")

	tx := pgdb.NewTransactor(pool, func(db pgdb.DBTX) pgdb.DBTX { return db })

	err := tx.InTx(ctx, func(ctx context.Context, db pgdb.DBTX) error {
		if err := collector.NewRepository(db).SaveCollector(ctx, domaintest.Collector(testCollectorID, "A", "a@example.com")); err != nil {
			return err
		}

		return persona.NewStore(db).For(testCollectorID).SavePersona(ctx, clashing)
	})
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Fatalf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}

	if _, err := collector.NewRepository(pool).GetCollector(ctx, testCollectorID); !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("GetCollector(testCollectorID) err = %v, want %v: the collector was not rolled back", err, domain.ErrCollectorNotFound)
	}

	// With the collector rolled back, the store reports it missing before looking for the persona.
	if _, err := persona.NewStore(pool).For(testCollectorID).GetPersona(ctx, clashing.ID()); !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("GetPersona err = %v, want %v", err, domain.ErrCollectorNotFound)
	}

	// The seeded data of the other collector is untouched.
	got, err := persona.NewStore(pool).For(otherCollectorID).GetPersona(ctx, existing.ID())
	if err != nil {
		t.Fatalf("seeded persona lost: %v", err)
	}

	if got.CollectorID() != otherCollectorID || got.Name() != existing.Name() {
		t.Errorf("seeded persona = %q of %v, want %q of %v", got.Name(), got.CollectorID(), existing.Name(), otherCollectorID)
	}
}
