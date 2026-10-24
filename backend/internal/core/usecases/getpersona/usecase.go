// Package getpersona implements the use case to retrieve a persona of a collector by its id.
package getpersona

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/tenant"
)

// Usecase is the getpersona use case.
type Usecase struct {
	repo port.PersonaRepository
}

// New returns a use case that gets personas from repo.
func New(repo port.PersonaRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute retrieves a persona of the collector the context acts for by its id.
func (u *Usecase) Execute(ctx context.Context, id domain.PersonaID) (*domain.Persona, error) {
	collectorID, err := tenant.MustCollectorFrom(ctx)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	p, err := u.repo.GetPersona(ctx, collectorID, id)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	return p, nil
}
