// Package listpersonas implements the use case to list all personas of a collector.
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

// Execute retrieves all personas of the collector.
// It returns domain.ErrCollectorNotFound if that collector does not exist.
func (u *Usecase) Execute(ctx context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error) {
	personas, err := u.personas.For(collectorID).ListPersonas(ctx)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	return personas, nil
}
