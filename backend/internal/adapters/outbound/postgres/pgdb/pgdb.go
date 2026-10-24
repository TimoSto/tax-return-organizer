// Package pgdb contains what the Postgres repositories share to run on either the pool or a transaction.
package pgdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

// DBTX is the part of the pgx API the repositories use. Both *pgxpool.Pool and pgx.Tx implement it, so a
// repository works the same on the pool and inside a transaction. Begin on a pgx.Tx starts a savepoint.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
}

var (
	_ DBTX = (*pgxpool.Pool)(nil)
	_ DBTX = pgx.Tx(nil)
)

// Transactor implements port.Transactor on a pgx pool.
type Transactor struct {
	pool  *pgxpool.Pool
	repos func(DBTX) port.Repositories
}

var _ port.Transactor = (*Transactor)(nil)

// NewTransactor returns a transactor on pool. repos builds the repositories that run on the given transaction;
// the composition root passes it so this package does not depend on the resource packages.
func NewTransactor(pool *pgxpool.Pool, repos func(DBTX) port.Repositories) *Transactor {
	return &Transactor{pool: pool, repos: repos}
}

// InTx implements port.Transactor. It rolls back if fn returns an error or panics.
func (t *Transactor) InTx(ctx context.Context, fn func(ctx context.Context, repos port.Repositories) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// A no-op after Commit. The rollback must not depend on ctx, which may be the reason fn failed.
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if err := fn(ctx, t.repos(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
