// Package ports defines the interfaces between the core and its adapters.
package ports

import (
	"context"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/google/uuid"
)

// CreateCollectorCmd is the input of the CreateCollector use case.
type CreateCollectorCmd struct {
	Name  string
	Email string
}

// CreateCollector creates a collector and returns its id.
type CreateCollectorPort func(ctx context.Context, cmd CreateCollectorCmd) (uuid.UUID, error)

// GetCollector retrieves a collector by its id.
type GetCollectorPort func(ctx context.Context, id uuid.UUID) (*domain.Collector, error)
