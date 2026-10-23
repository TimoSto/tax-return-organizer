// Package ports defines the interfaces between the core and its adapters.
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// CreateCollectorCmd is the input of the CreateCollector use case.
type CreateCollectorCmd struct {
	Name  string
	Email string
}

// CreateCollectorPort creates a collector and returns its id.
type CreateCollectorPort func(ctx context.Context, cmd CreateCollectorCmd) (domain.CollectorID, error)

// GetCollectorPort retrieves a collector by its id.
type GetCollectorPort func(ctx context.Context, id domain.CollectorID) (*domain.Collector, error)

// GetPersonaPort retrieves a persona of a collector by its id.
type GetPersonaPort func(ctx context.Context, collectorID domain.CollectorID, id uuid.UUID) (*domain.Persona, error)
