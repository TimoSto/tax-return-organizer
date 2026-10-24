// Package listpersonas implements the use case to list all personas of a collector.
package listpersonas

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/tenant"
)

// Usecase is the listpersonas use case.
type Usecase struct {
	repo port.PersonaRepository
}

// New returns a use case that lists personas from repo.
func New(repo port.PersonaRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute retrieves all personas of the collector the context acts for.
func (u *Usecase) Execute(ctx context.Context) ([]*domain.Persona, error) {
	collectorID, err := tenant.MustCollectorFrom(ctx)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	personas, err := u.repo.ListPersonas(ctx, collectorID)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	return personas, nil
}
