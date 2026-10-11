// Package createcollector implements the use case of creating a new collector.
package createcollector

import (
	"context"
	"errors"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

var _ ports.CreateCollectorPort = (*Usecase)(nil).Execute

// maxIDAttempts bounds the retries when a randomly drawn collector id is already taken.
const maxIDAttempts = 5

// Usecase implements ports.CreateCollector.
type Usecase struct {
	repo ports.CollectorRepository
}

// New returns a use case that stores collectors in repo.
func New(repo ports.CollectorRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute creates and persists a collector and returns its id.
func (u *Usecase) Execute(ctx context.Context, cmd ports.CreateCollectorCmd) (domain.CollectorID, error) {
	for range maxIDAttempts {
		id, err := domain.NewCollectorID()
		if err != nil {
			return "", err
		}

		c, err := domain.NewCollector(id, cmd.Name, cmd.Email)
		if err != nil {
			return "", err
		}

		err = u.repo.SaveCollector(ctx, c)
		if errors.Is(err, domain.ErrCollectorIDTaken) {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("save collector: %w", err)
		}

		return c.ID, nil
	}

	return "", fmt.Errorf("save collector: no free id after %d attempts", maxIDAttempts)
}
