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
	collectors port.CollectorRepository
	personas   port.PersonaRepository
}

// New returns a use case that lists personas from personas, for collectors found in collectors.
func New(collectors port.CollectorRepository, personas port.PersonaRepository) *Usecase {
	return &Usecase{collectors: collectors, personas: personas}
}

// Execute retrieves all personas of the collector the context acts for.
// It returns domain.ErrCollectorNotFound if that collector does not exist.
func (u *Usecase) Execute(ctx context.Context) ([]*domain.Persona, error) {
	collectorID, err := tenant.MustCollectorFrom(ctx)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	if _, err := u.collectors.GetCollector(ctx, collectorID); err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	personas, err := u.personas.ListPersonas(ctx, collectorID)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	return personas, nil
}
