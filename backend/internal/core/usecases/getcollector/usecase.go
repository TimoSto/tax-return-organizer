package getcollector

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

var _ ports.GetCollectorPort = (*Usecase)(nil).Execute

// Usecase implements ports.GetCollector.
type Usecase struct {
	repo ports.CollectorRepository
}

// New returns a use case that gets collectors from repo.
func New(repo ports.CollectorRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute retrieves a collector by its id.
func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (*domain.Collector, error) {
	c, err := u.repo.GetCollector(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}
	return c, nil
}

