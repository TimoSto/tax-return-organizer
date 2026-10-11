// Package createcollector implements the use case of creating a new collector.
package createcollector

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

var _ ports.CreateCollector = (*Usecase)(nil)

// Usecase implements ports.CreateCollector.
type Usecase struct {
	repo ports.CollectorRepository
}

// New returns a use case that stores collectors in repo.
func New(repo ports.CollectorRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute creates and persists a collector and returns its id.
func (u *Usecase) Execute(ctx context.Context, cmd ports.CreateCollectorCmd) (uuid.UUID, error) {
	c, err := domain.NewCollector(uuid.New(), cmd.Name, cmd.Email)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.repo.SaveCollector(ctx, c); err != nil {
		return uuid.Nil, fmt.Errorf("save collector: %w", err)
	}

	return c.ID, nil
}
