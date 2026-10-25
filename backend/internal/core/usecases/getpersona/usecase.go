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
	collectors outbound.CollectorRepository
	personas   outbound.PersonaStore
}

var _ inbound.GetPersona = (*Usecase)(nil)

// New returns a use case that gets personas from personas, for collectors found in collectors.
func New(collectors outbound.CollectorRepository, personas outbound.PersonaStore) *Usecase {
	return &Usecase{collectors: collectors, personas: personas}
}

// Execute retrieves a persona of the collector by its id.
// It returns domain.ErrCollectorNotFound if the collector does not exist and domain.ErrPersonaNotFound if it has
// no persona with the id, which includes personas of other collectors.
func (u *Usecase) Execute(ctx context.Context, collectorID domain.CollectorID, id domain.PersonaID) (*domain.Persona, error) {
	if _, err := u.collectors.GetCollector(ctx, collectorID); err != nil {
		return nil, fmt.Errorf("check collector: %w", err)
	}

	p, err := u.personas.For(collectorID).GetPersona(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get persona: %w", err)
	}

	return p, nil
}
