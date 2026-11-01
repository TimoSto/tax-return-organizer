// Package listpersonas implements the use case to list all personas of a collector.
package listpersonas

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/locale"
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

// Execute retrieves all personas of the actor's collector, sorted by name for a reader and then by id, so equal names
// keep a stable order. Only staff may: clients never see personas.
// It returns domain.ErrForbidden for any other actor, before the collector is looked at, and
// domain.ErrCollectorNotFound if the collector does not exist.
func (u *Usecase) Execute(ctx context.Context, actor domain.Actor) ([]*domain.Persona, error) {
	if err := actor.RequireStaff(); err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	personas, err := u.personas.For(actor.CollectorID()).ListPersonas(ctx)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	// TODO: sort for the collector's or the user's locale once one is stored.
	compareNames := locale.DefaultLocale.Comparer()
	slices.SortFunc(personas, func(a, b *domain.Persona) int {
		return cmp.Or(compareNames(a.Name(), b.Name()), compareIDs(a.ID(), b.ID()))
	})

	return personas, nil
}

func compareIDs(a, b domain.PersonaID) int { return bytes.Compare(a[:], b[:]) }
