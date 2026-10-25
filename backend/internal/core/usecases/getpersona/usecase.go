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
	personas port.PersonaStore
}

// New returns a use case that gets personas from the persona store.
func New(personas port.PersonaStore) *Usecase {
	return &Usecase{personas: personas}
}

// Execute retrieves a persona of the collector the context acts for by its id.
func (u *Usecase) Execute(ctx context.Context, id domain.PersonaID) (*domain.Persona, error) {
	collectorID, err := tenant.RequireCollectorFrom(ctx)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	p, err := u.personas.For(collectorID).GetPersona(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	return p, nil
}
