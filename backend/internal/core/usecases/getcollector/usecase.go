// Package getcollector implements the use case to retrieve the collector of the caller.
package getcollector

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// Usecase is the getcollector use case.
type Usecase struct {
	repo outbound.CollectorRepository
}

var _ inbound.GetCollector = (*Usecase)(nil)

// New returns a use case that gets collectors from repo.
func New(repo outbound.CollectorRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute retrieves the collector the actor belongs to. Staff and clients may, the zero actor has no collector
// and gets domain.ErrCollectorNotFound.
func (u *Usecase) Execute(ctx context.Context, actor domain.Actor) (*domain.Collector, error) {
	c, err := u.repo.GetCollector(ctx, actor.CollectorID())
	if err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}
	return c, nil
}
