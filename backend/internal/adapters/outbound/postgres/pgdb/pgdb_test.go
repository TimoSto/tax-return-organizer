package pgdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/postgrestest"
)

// probe is a transactor whose callbacks write into the table intx_probe instead of going through repositories, so
// the tests do not depend on any resource package.
type probe struct {
	*pgdb.Transactor
	pool *pgxpool.Pool
	// tx is the transaction of the running InTx call, captured when the transactor builds its repositories.
	tx pgdb.DBTX
}

// newProbe returns a probe on an empty intx_probe table.
func newProbe(t *testing.T) *probe {
	t.Helper()

	pool := postgrestest.RequirePool(t)

	if _, err := pool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS intx_probe (n int PRIMARY KEY)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	if _, err := pool.Exec(context.Background(), `TRUNCATE intx_probe`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	p := &probe{pool: pool}
	p.Transactor = pgdb.NewTransactor(pool, func(db pgdb.DBTX) outbound.Repositories {
		p.tx = db
		return outbound.Repositories{}
	})

	return p
}

// insert writes n inside the running transaction.
func (p *probe) insert(ctx context.Context, n int) error {
	_, err := p.tx.Exec(ctx, `INSERT INTO intx_probe (n) VALUES ($1)`, n)
	return err
}

// insertAll returns an InTx callback that inserts ns and then returns err.
func (p *probe) insertAll(err error, ns ...int) func(context.Context, outbound.Repositories) error {
	return func(ctx context.Context, _ outbound.Repositories) error {
		for _, n := range ns {
			if err := p.insert(ctx, n); err != nil {
				return err
			}
		}

		return err
	}
}

// rows counts the committed rows.
func (p *probe) rows(t *testing.T) int {
	t.Helper()

	var n int
	if err := p.pool.QueryRow(context.Background(), `SELECT count(*) FROM intx_probe`).Scan(&n); err != nil {
		t.Fatalf("count rows: %v", err)
	}

	return n
}

func TestInTx_CommitsOnSuccess(t *testing.T) {
	p := newProbe(t)

	if err := p.InTx(context.Background(), p.insertAll(nil, 1, 2)); err != nil {
		t.Fatalf("InTx: %v", err)
	}

	if n := p.rows(t); n != 2 {
		t.Errorf("rows = %d, want 2", n)
	}
}

func TestInTx_RollsBackOnError(t *testing.T) {
	p := newProbe(t)
	boom := errors.New("boom")

	if err := p.InTx(context.Background(), p.insertAll(boom, 1, 2)); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}

	if n := p.rows(t); n != 0 {
		t.Errorf("rows = %d, want 0 after rollback", n)
	}
}

func TestInTx_RollsBackOnPanic(t *testing.T) {
	p := newProbe(t)

	func() {
		defer func() { _ = recover() }()

		_ = p.InTx(context.Background(), func(ctx context.Context, r outbound.Repositories) error {
			if err := p.insertAll(nil, 1)(ctx, r); err != nil {
				return err
			}

			panic("boom")
		})
	}()

	if n := p.rows(t); n != 0 {
		t.Errorf("rows = %d, want 0 after panic", n)
	}
}

// A failing statement aborts the Postgres transaction. InTx returns the database error unchanged, so a repository
// can map it, and the rollback removes the earlier writes.
func TestInTx_RollsBackAfterDatabaseError(t *testing.T) {
	p := newProbe(t)

	err := p.InTx(context.Background(), p.insertAll(nil, 1, 2, 1))

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("err = %v, want a unique violation (23505)", err)
	}

	if n := p.rows(t); n != 0 {
		t.Errorf("rows = %d, want 0 after rollback", n)
	}
}
