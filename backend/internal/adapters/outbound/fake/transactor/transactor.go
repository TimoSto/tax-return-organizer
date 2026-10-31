// Package transactor contains the in-memory transactor for tests.
package transactor

import (
	"context"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

var _ outbound.Transactor[struct{}] = (*Transactor[struct{}])(nil)

// Transactor runs the callback on fixed repositories. It does not roll back: the in-memory repositories have no
// transactions, rollback behaviour is covered by the Postgres tests.
type Transactor[R any] struct {
	repos R
	// Calls counts the InTx calls.
	Calls int
}

// New returns a transactor that hands repos to every callback.
func New[R any](repos R) *Transactor[R] {
	return &Transactor[R]{repos: repos}
}

// InTx implements outbound.Transactor.
func (t *Transactor[R]) InTx(ctx context.Context, fn func(ctx context.Context, repos R) error) error {
	t.Calls++

	return fn(ctx, t.repos)
}
