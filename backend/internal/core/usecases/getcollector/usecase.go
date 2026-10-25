// Package getcollector implements the use case to retrieve a collector by its id.
package getcollector

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/tenant"
)

// Usecase is the getcollector use case.
type Usecase struct {
	repo port.CollectorRepository
}

// New returns a use case that gets collectors from repo.
func New(repo port.CollectorRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute retrieves the collector the context acts for.
func (u *Usecase) Execute(ctx context.Context) (*domain.Collector, error) {
	id, err := tenant.Require(ctx)
	if err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}

	c, err := u.repo.GetCollector(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}
	return c, nil
}
