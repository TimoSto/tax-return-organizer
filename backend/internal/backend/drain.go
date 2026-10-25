package backend

import (
	"context"
	"sync/atomic"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/health"
)

// drainingPinger delegates to inner until startDraining is called, then reports not ready.
type drainingPinger struct {
	inner    health.Pinger
	draining atomic.Bool
}

func (p *drainingPinger) Ping(ctx context.Context) error {
	if p.draining.Load() {
		return health.ErrDraining
	}

	return p.inner.Ping(ctx)
}

func (p *drainingPinger) startDraining() { p.draining.Store(true) }
