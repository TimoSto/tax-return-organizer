// Package transactor contains the in-memory transactor for tests.
package transactor

import (
	"context"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

var _ port.Transactor = (*Transactor)(nil)

// Transactor runs the callback on fixed repositories. It does not roll back: the in-memory repositories have no
// transactions, rollback behaviour is covered by the Postgres tests.
type Transactor struct {
	repos port.Repositories
	// Calls counts the InTx calls.
	Calls int
}

// New returns a transactor that hands repos to every callback.
func New(repos port.Repositories) *Transactor {
	return &Transactor{repos: repos}
}

// InTx implements port.Transactor.
func (t *Transactor) InTx(ctx context.Context, fn func(ctx context.Context, repos port.Repositories) error) error {
	t.Calls++

	return fn(ctx, t.repos)
}
