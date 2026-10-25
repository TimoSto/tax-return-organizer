// Package listpersonas implements the use case to list all personas of a collector.
package listpersonas

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

// Usecase is the listpersonas use case.
type Usecase struct {
	collectors port.CollectorRepository
	personas   port.PersonaStore
}

// New returns a use case that lists personas from personas, for collectors found in collectors.
func New(collectors port.CollectorRepository, personas port.PersonaStore) *Usecase {
	return &Usecase{collectors: collectors, personas: personas}
}

// Execute retrieves all personas of the collector.
// It returns domain.ErrCollectorNotFound if that collector does not exist.
func (u *Usecase) Execute(ctx context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error) {
	if _, err := u.collectors.GetCollector(ctx, collectorID); err != nil {
		return nil, fmt.Errorf("check collector: %w", err)
	}

	personas, err := u.personas.For(collectorID).ListPersonas(ctx)
	if err != nil {
		return nil, fmt.Errorf("list personas: %w", err)
	}

	return personas, nil
}
