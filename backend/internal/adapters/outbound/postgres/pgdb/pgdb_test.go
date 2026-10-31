package pgdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	pgcollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/collector"
	pgpersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/postgrestest"
)

var (
	// testCollectorID is the collector the tests act for.
	testCollectorID = domaintest.CollectorID()
)

func repositories(db pgdb.DBTX) outbound.Repositories {
	return outbound.Repositories{Collectors: pgcollector.NewRepository(db), Personas: pgpersona.NewStore(db)}
}

// newTransactor returns a transactor on an emptied database.
func newTransactor(t *testing.T, build func(pgdb.DBTX) outbound.Repositories) (*pgdb.Transactor, *pgxpool.Pool) {
	t.Helper()

	pool := postgrestest.RequirePool(t)

	if _, err := pool.Exec(context.Background(), `TRUNCATE personas, collectors CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	return pgdb.NewTransactor(pool, build), pool
}

func countCollectors(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()

	var n int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM collectors`).Scan(&n); err != nil {
		t.Fatalf("count collectors: %v", err)
	}

	return n
}

func countPersonas(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()

	var n int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM personas`).Scan(&n); err != nil {
		t.Fatalf("count personas: %v", err)
	}

	return n
}

// saveCollectorAndPersona writes through both repositories of r, so the tests cover a transaction across them.
func saveCollectorAndPersona(ctx context.Context, r outbound.Repositories) error {
	if err := r.Collectors.SaveCollector(ctx, domaintest.Collector(testCollectorID, "A", "a@example.com")); err != nil {
		return err
	}

	p, err := domain.NewPersona(domain.PersonaID(uuid.New()), testCollectorID, "Employed", "")
	if err != nil {
		return err
	}

	return r.Personas.For(testCollectorID).SavePersona(ctx, p)
}

func TestInTx_CommitsOnSuccess(t *testing.T) {
	tx, pool := newTransactor(t, repositories)

	if err := tx.InTx(context.Background(), saveCollectorAndPersona); err != nil {
		t.Fatalf("InTx: %v", err)
	}

	if c, p := countCollectors(t, pool), countPersonas(t, pool); c != 1 || p != 1 {
		t.Errorf("collectors = %d, personas = %d, want 1 and 1", c, p)
	}
}

func TestInTx_RollsBackOnError(t *testing.T) {
	tx, pool := newTransactor(t, repositories)
	boom := errors.New("boom")

	err := tx.InTx(context.Background(), func(ctx context.Context, r outbound.Repositories) error {
		if err := r.Collectors.SaveCollector(ctx, domaintest.Collector(testCollectorID, "A", "a@example.com")); err != nil {
			return err
		}

		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}

	if n := countCollectors(t, pool); n != 0 {
		t.Errorf("collectors = %d, want 0 after rollback", n)
	}
}

func TestInTx_RollsBackOnPanic(t *testing.T) {
	tx, pool := newTransactor(t, repositories)

	func() {
		defer func() { _ = recover() }()

		_ = tx.InTx(context.Background(), func(ctx context.Context, r outbound.Repositories) error {
			if err := r.Collectors.SaveCollector(ctx, domaintest.Collector(testCollectorID, "A", "a@example.com")); err != nil {
				return err
			}

			panic("boom")
		})
	}()

	if n := countCollectors(t, pool); n != 0 {
		t.Errorf("collectors = %d, want 0 after panic", n)
	}
}

// SavePersona joins the transaction of InTx: rolling back the outer transaction removes the persona too.
func TestInTx_SavePersonaInsideTransaction(t *testing.T) {
	tx, pool := newTransactor(t, repositories)
	boom := errors.New("boom")

	err := tx.InTx(context.Background(), func(ctx context.Context, r outbound.Repositories) error {
		if err := saveCollectorAndPersona(ctx, r); err != nil {
			return err
		}

		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}

	if c, p := countCollectors(t, pool), countPersonas(t, pool); c != 0 || p != 0 {
		t.Errorf("collectors = %d, personas = %d, want 0 and 0", c, p)
	}
}

// A failing statement aborts the Postgres transaction. The repository still maps the error, and the rollback removes
// the earlier writes.
func TestInTx_RollsBackAfterDatabaseError(t *testing.T) {
	tx, pool := newTransactor(t, repositories)

	err := tx.InTx(context.Background(), func(ctx context.Context, r outbound.Repositories) error {
		if err := saveCollectorAndPersona(ctx, r); err != nil {
			return err
		}

		return r.Collectors.SaveCollector(ctx, domaintest.Collector(testCollectorID, "B", "b@example.com"))
	})
	if !errors.Is(err, domain.ErrCollectorIDTaken) {
		t.Fatalf("err = %v, want it to wrap %v", err, domain.ErrCollectorIDTaken)
	}

	if c, p := countCollectors(t, pool), countPersonas(t, pool); c != 0 || p != 0 {
		t.Errorf("collectors = %d, personas = %d, want 0 and 0", c, p)
	}
}
