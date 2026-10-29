package pgdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/idgen"
	pgcollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/collector"
	pgpersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
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

func TestInTx_CommitsOnSuccess(t *testing.T) {
	tx, pool := newTransactor(t, repositories)

	err := tx.InTx(context.Background(), func(ctx context.Context, r outbound.Repositories) error {
		return r.Collectors.SaveCollector(ctx, domaintest.Collector(testCollectorID, "A", "a@example.com"))
	})
	if err != nil {
		t.Fatalf("InTx: %v", err)
	}

	if n := countCollectors(t, pool); n != 1 {
		t.Errorf("collectors = %d, want 1", n)
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
		if err := r.Collectors.SaveCollector(ctx, domaintest.Collector(testCollectorID, "A", "a@example.com")); err != nil {
			return err
		}

		p, err := domain.NewPersona(domain.PersonaID(uuid.New()), testCollectorID, "Employed", "")
		if err != nil {
			return err
		}

		if err := r.Personas.For(testCollectorID).SavePersona(ctx, p); err != nil {
			return err
		}

		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}

	var personas int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM personas`).Scan(&personas); err != nil {
		t.Fatalf("count personas: %v", err)
	}

	if personas != 0 || countCollectors(t, pool) != 0 {
		t.Errorf("personas = %d, collectors = %d, want 0 and 0", personas, countCollectors(t, pool))
	}
}

// failingPersonas fails the n-th SavePersona, whichever collector's repository it goes through, like a database
// error in the middle of creating a collector.
type failingPersonas struct {
	outbound.PersonaStore
	failAt, calls int
}

func (s *failingPersonas) For(id domain.CollectorID) outbound.PersonaRepository {
	return &failingPersonaRepo{PersonaRepository: s.PersonaStore.For(id), store: s}
}

type failingPersonaRepo struct {
	outbound.PersonaRepository
	store *failingPersonas
}

func (r *failingPersonaRepo) SavePersona(ctx context.Context, p *domain.Persona) error {
	r.store.calls++
	if r.store.calls == r.store.failAt {
		return errors.New("simulated persona failure")
	}

	return r.PersonaRepository.SavePersona(ctx, p)
}

// Regression test: a persona failing after the collector was saved must not leave the collector behind.
func TestCreateCollector_PersonaFailureLeavesNothing(t *testing.T) {
	failing := &failingPersonas{failAt: 2}
	tx, pool := newTransactor(t, func(db pgdb.DBTX) outbound.Repositories {
		r := repositories(db)
		failing.PersonaStore = r.Personas
		r.Personas = failing

		return r
	})

	_, err := createcollector.New(tx, personatemplate.Predefined(), idgen.Random{}).Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if err == nil {
		t.Fatal("expected an error")
	}

	if n := countCollectors(t, pool); n != 0 {
		t.Errorf("collectors = %d, want 0", n)
	}

	var personas int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM personas`).Scan(&personas); err != nil {
		t.Fatalf("count personas: %v", err)
	}

	if personas != 0 {
		t.Errorf("personas = %d, want 0", personas)
	}
}

func TestCreateCollector_StoresCollectorWithPersonas(t *testing.T) {
	tx, pool := newTransactor(t, repositories)

	id, err := createcollector.New(tx, personatemplate.Predefined(), idgen.Random{}).Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	personas, err := pgpersona.NewStore(pool).For(id).ListPersonas(context.Background())
	if err != nil {
		t.Fatalf("ListPersonas: %v", err)
	}

	if len(personas) != len(personatemplate.Predefined()) {
		t.Errorf("personas = %d, want %d", len(personas), len(personatemplate.Predefined()))
	}
}

// Collector ids are random UUIDs, so a taken id is a bug: it fails without a retry and stores nothing.
func TestCreateCollector_TakenIDFails(t *testing.T) {
	tx, pool := newTransactor(t, repositories)

	if _, err := pool.Exec(context.Background(), `INSERT INTO collectors (id, name, email) VALUES ($1, 'A', 'a@example.com')`, uuid.UUID(testCollectorID)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	uc := createcollector.New(tx, personatemplate.Predefined(), fixedIDs{id: testCollectorID})

	if _, err := uc.Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"}); !errors.Is(err, domain.ErrCollectorIDTaken) {
		t.Fatalf("err = %v, want it to wrap %v", err, domain.ErrCollectorIDTaken)
	}

	if n := countCollectors(t, pool); n != 1 {
		t.Errorf("collectors = %d, want 1", n)
	}
}

// fixedIDs draws the collector id id and random ids otherwise.
type fixedIDs struct {
	idgen.Random
	id domain.CollectorID
}

func (f fixedIDs) NewCollectorID() (domain.CollectorID, error) { return f.id, nil }
