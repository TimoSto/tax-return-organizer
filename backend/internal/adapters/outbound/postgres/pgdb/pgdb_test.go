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
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/postgrestest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
)

func repositories(db pgdb.DBTX) port.Repositories {
	return port.Repositories{Collectors: pgcollector.NewRepository(db), Personas: pgpersona.NewRepository(db)}
}

// newTransactor returns a transactor on an emptied database.
func newTransactor(t *testing.T, build func(pgdb.DBTX) port.Repositories) (*pgdb.Transactor, *pgxpool.Pool) {
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

	err := tx.InTx(context.Background(), func(ctx context.Context, r port.Repositories) error {
		return r.Collectors.SaveCollector(ctx, &domain.Collector{ID: "aaaa", Name: "A", Email: "a@example.com"})
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

	err := tx.InTx(context.Background(), func(ctx context.Context, r port.Repositories) error {
		if err := r.Collectors.SaveCollector(ctx, &domain.Collector{ID: "aaaa", Name: "A", Email: "a@example.com"}); err != nil {
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

		_ = tx.InTx(context.Background(), func(ctx context.Context, r port.Repositories) error {
			if err := r.Collectors.SaveCollector(ctx, &domain.Collector{ID: "aaaa", Name: "A", Email: "a@example.com"}); err != nil {
				return err
			}

			panic("boom")
		})
	}()

	if n := countCollectors(t, pool); n != 0 {
		t.Errorf("collectors = %d, want 0 after panic", n)
	}
}

// SavePersona opens its own transaction, which inside InTx is a savepoint of the outer one.
func TestInTx_SavePersonaInsideTransaction(t *testing.T) {
	tx, pool := newTransactor(t, repositories)
	boom := errors.New("boom")

	err := tx.InTx(context.Background(), func(ctx context.Context, r port.Repositories) error {
		if err := r.Collectors.SaveCollector(ctx, &domain.Collector{ID: "aaaa", Name: "A", Email: "a@example.com"}); err != nil {
			return err
		}

		p, err := domain.NewPersona(domain.PersonaID(uuid.New()), "aaaa", "Employed", "")
		if err != nil {
			return err
		}

		if err := r.Personas.SavePersona(ctx, p); err != nil {
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

// failingPersonas fails the n-th SavePersona, like a database error in the middle of creating a collector.
type failingPersonas struct {
	port.PersonaRepository
	failAt, calls int
}

func (r *failingPersonas) SavePersona(ctx context.Context, p *domain.Persona) error {
	r.calls++
	if r.calls == r.failAt {
		return errors.New("simulated persona failure")
	}

	return r.PersonaRepository.SavePersona(ctx, p)
}

// Regression test: a persona failing after the collector was saved must not leave the collector behind.
func TestCreateCollector_PersonaFailureLeavesNothing(t *testing.T) {
	failing := &failingPersonas{failAt: 2}
	tx, pool := newTransactor(t, func(db pgdb.DBTX) port.Repositories {
		r := repositories(db)
		failing.PersonaRepository = r.Personas
		r.Personas = failing

		return r
	})

	_, err := createcollector.New(tx).Execute(context.Background(), createcollector.Command{Name: "Jane", Email: "jane@example.com"})
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

	id, err := createcollector.New(tx).Execute(context.Background(), createcollector.Command{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	personas, err := pgpersona.NewRepository(pool).ListPersonas(context.Background(), id)
	if err != nil {
		t.Fatalf("ListPersonas: %v", err)
	}

	if len(personas) != len(domain.PersonaTemplates()) {
		t.Errorf("personas = %d, want %d", len(personas), len(domain.PersonaTemplates()))
	}
}

// A taken id aborts the Postgres transaction, so the retry must run in a new one and still succeed.
func TestCreateCollector_RetriesAfterTakenID(t *testing.T) {
	tx, pool := newTransactor(t, repositories)

	if _, err := pool.Exec(context.Background(), `INSERT INTO collectors (id, name, email) VALUES ('aaaa', 'A', 'a@example.com')`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	ids := []domain.CollectorID{"aaaa", "bbbb"}
	uc := createcollector.New(tx, createcollector.WithIDGenerator(&sequenceIDs{ids: ids}))

	id, err := uc.Execute(context.Background(), createcollector.Command{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if id != "bbbb" {
		t.Errorf("id = %q, want bbbb", id)
	}

	if n := countCollectors(t, pool); n != 2 {
		t.Errorf("collectors = %d, want 2", n)
	}
}

// sequenceIDs hands out the given collector ids in order and random ids otherwise.
type sequenceIDs struct {
	domain.RandomIDs

	ids []domain.CollectorID
}

func (s *sequenceIDs) NewCollectorID() (domain.CollectorID, error) {
	id := s.ids[0]
	s.ids = s.ids[1:]

	return id, nil
}
