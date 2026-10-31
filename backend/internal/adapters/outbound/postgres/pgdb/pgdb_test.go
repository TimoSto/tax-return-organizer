package pgdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/postgrestest"
)

// newTransactor returns a transactor that hands its callbacks the transaction itself, so the tests write into the
// table intx_probe instead of going through repositories and do not depend on any resource package.
func newTransactor(t *testing.T) (*pgdb.Transactor[pgdb.DBTX], *pgxpool.Pool) {
	t.Helper()

	pool := postgrestest.RequirePool(t)

	if _, err := pool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS intx_probe (n int PRIMARY KEY)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	if _, err := pool.Exec(context.Background(), `TRUNCATE intx_probe`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	return pgdb.NewTransactor(pool, func(db pgdb.DBTX) pgdb.DBTX { return db }), pool
}

// insertAll returns an InTx callback that inserts ns and then returns err.
func insertAll(err error, ns ...int) func(context.Context, pgdb.DBTX) error {
	return func(ctx context.Context, db pgdb.DBTX) error {
		for _, n := range ns {
			if _, err := db.Exec(ctx, `INSERT INTO intx_probe (n) VALUES ($1)`, n); err != nil {
				return err
			}
		}

		return err
	}
}

// rows counts the committed rows.
func rows(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()

	var n int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM intx_probe`).Scan(&n); err != nil {
		t.Fatalf("count rows: %v", err)
	}

	return n
}

func TestInTx_CommitsOnSuccess(t *testing.T) {
	tx, pool := newTransactor(t)

	if err := tx.InTx(context.Background(), insertAll(nil, 1, 2)); err != nil {
		t.Fatalf("InTx: %v", err)
	}

	if n := rows(t, pool); n != 2 {
		t.Errorf("rows = %d, want 2", n)
	}
}

func TestInTx_RollsBackOnError(t *testing.T) {
	tx, pool := newTransactor(t)
	boom := errors.New("boom")

	if err := tx.InTx(context.Background(), insertAll(boom, 1, 2)); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}

	if n := rows(t, pool); n != 0 {
		t.Errorf("rows = %d, want 0 after rollback", n)
	}
}

func TestInTx_RollsBackOnPanic(t *testing.T) {
	tx, pool := newTransactor(t)

	func() {
		defer func() { _ = recover() }()

		_ = tx.InTx(context.Background(), func(ctx context.Context, db pgdb.DBTX) error {
			if err := insertAll(nil, 1)(ctx, db); err != nil {
				return err
			}

			panic("boom")
		})
	}()

	if n := rows(t, pool); n != 0 {
		t.Errorf("rows = %d, want 0 after panic", n)
	}
}

// A failing statement aborts the Postgres transaction. InTx returns the database error unchanged, so a repository
// can map it, and the rollback removes the earlier writes.
func TestInTx_RollsBackAfterDatabaseError(t *testing.T) {
	tx, pool := newTransactor(t)

	err := tx.InTx(context.Background(), insertAll(nil, 1, 2, 1))

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("err = %v, want a unique violation (23505)", err)
	}

	if n := rows(t, pool); n != 0 {
		t.Errorf("rows = %d, want 0 after rollback", n)
	}
}
