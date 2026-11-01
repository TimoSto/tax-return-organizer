// Package listpersonas implements the use case to list all personas of a tax practice.
package listpersonas

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// Usecase is the listpersonas use case.
type Usecase struct {
	personas outbound.PersonaStore
}

var _ inbound.ListPersonas = (*Usecase)(nil)

// New returns a use case that lists personas from personas.
func New(personas outbound.PersonaStore) *Usecase {
	return &Usecase{personas: personas}
}

// Execute retrieves all personas of the actor's tax practice. Only staff may: clients never see personas.
// It returns domain.ErrForbidden for any other actor, before the tax practice is looked at, and
// domain.ErrTaxPracticeNotFound if the tax practice does not exist.
func (u *Usecase) Execute(ctx context.Context, actor domain.Actor) ([]*domain.Persona, error) {
	if err := actor.RequireStaff(); err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	personas, err := u.personas.For(actor.TaxPracticeID()).ListPersonas(ctx)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	return personas, nil
}
