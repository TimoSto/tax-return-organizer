// Package getcollector implements the use case to retrieve a collector by its id.
package getcollector

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

// Usecase is the getcollector use case.
type Usecase struct {
	repo port.CollectorRepository
}

// New returns a use case that gets collectors from repo.
func New(repo port.CollectorRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute retrieves a collector by its id.
func (u *Usecase) Execute(ctx context.Context, id domain.CollectorID) (*domain.Collector, error) {
	c, err := u.repo.GetCollector(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}
	return c, nil
}
