// Package createcollector implements the use case of creating a new collector.
package createcollector

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

var _ ports.CreateCollectorPort = (*Usecase)(nil).Execute

var ErrNameRequired = errors.New("collector name is required")

type Usecase struct {
	repo ports.CollectorRepository
}

func New(repo ports.CollectorRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute creates and persists a collector and returns its id.
func (u *Usecase) Execute(ctx context.Context, name, email string) (string, error) {
	name, email = strings.TrimSpace(name), strings.TrimSpace(email)

	if name == "" {
		return "", ErrNameRequired
	}

	if err := domain.ValidateEMail(email); err != nil {
		return "", err
	}

	c := &domain.Collector{ID: uuid.New(), Name: name, Email: email}
	if err := u.repo.SaveCollector(ctx, c); err != nil {
		return "", fmt.Errorf("save collector: %w", err)
	}

	return c.ID.String(), nil
}
