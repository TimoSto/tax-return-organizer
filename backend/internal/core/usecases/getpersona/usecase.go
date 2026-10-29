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

// New returns a use case that gets personas from personas.
func New(personas outbound.PersonaStore) *Usecase {
	return &Usecase{personas: personas}
}

// Execute retrieves a persona of the actor's collector by its id. Only staff may: clients never see personas.
// It returns domain.ErrForbidden for any other actor, before the collector is looked at,
// domain.ErrCollectorNotFound if the collector does not exist and domain.ErrPersonaNotFound if it has no persona
// with the id, which includes personas of other collectors.
func (u *Usecase) Execute(ctx context.Context, actor domain.Actor, id domain.PersonaID) (*domain.Persona, error) {
	if err := actor.RequireStaff(); err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	p, err := u.personas.For(actor.CollectorID()).GetPersona(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	return p, nil
}
