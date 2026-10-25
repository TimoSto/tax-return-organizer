// Package getcollector implements the use case to retrieve a collector by its id.
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

// Execute retrieves the collector with the given id.
func (u *Usecase) Execute(ctx context.Context, id domain.CollectorID) (*domain.Collector, error) {
	c, err := u.repo.GetCollector(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}
	return c, nil
}
