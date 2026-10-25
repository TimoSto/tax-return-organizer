// Package getpersona implements the use case to retrieve a persona of a collector by its id.
package getpersona

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// Usecase is the getpersona use case.
type Usecase struct {
	personas outbound.PersonaStore
}

var _ inbound.GetPersona = (*Usecase)(nil)

// New returns a use case that gets personas from the persona store.
func New(personas outbound.PersonaStore) *Usecase {
	return &Usecase{personas: personas}
}

// Execute retrieves a persona of the collector by its id.
func (u *Usecase) Execute(ctx context.Context, collectorID domain.CollectorID, id domain.PersonaID) (*domain.Persona, error) {
	p, err := u.personas.For(collectorID).GetPersona(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	return p, nil
}
