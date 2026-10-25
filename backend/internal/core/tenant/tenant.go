// Package tenant carries the collector a request acts for in the context. The inbound adapter sets it (today from
// the URL, later from the authenticated principal); use cases read it instead of taking a collector id argument.
package tenant

import (
	"context"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type ctxKey struct{}

// With returns a context that acts for the given collector.
func With(ctx context.Context, id domain.CollectorID) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// From returns the collector ctx acts for.
func From(ctx context.Context) (domain.CollectorID, bool) {
	id, ok := ctx.Value(ctxKey{}).(domain.CollectorID)

	return id, ok
}

// Require is From for use cases, returning ErrNoTenant when there is none: a missing tenant is a wiring bug, not a client error.
func Require(ctx context.Context) (domain.CollectorID, error) {
	id, ok := From(ctx)
	if !ok {
		return "", ErrNoTenant
	}

	return id, nil
}
