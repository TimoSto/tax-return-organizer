// Package tenant carries the collector a request acts for in the context. The inbound adapter sets it (today from
// the URL, later from the authenticated principal); use cases read it instead of taking a collector id argument.
package tenant

import (
	"context"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type ctxKey struct{}

// WithCollector returns a context that acts for the given collector.
func WithCollector(ctx context.Context, id domain.CollectorID) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// CollectorFrom returns the collector ctx acts for.
func CollectorFrom(ctx context.Context) (domain.CollectorID, bool) {
	id, ok := ctx.Value(ctxKey{}).(domain.CollectorID)

	return id, ok
}

// MustCollectorFrom is CollectorFrom for use cases: a missing tenant is a wiring bug, not a client error.
func MustCollectorFrom(ctx context.Context) (domain.CollectorID, error) {
	id, ok := CollectorFrom(ctx)
	if !ok {
		return "", ErrNoTenant
	}

	return id, nil
}
