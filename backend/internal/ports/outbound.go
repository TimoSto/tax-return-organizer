package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type CollectorRepository interface {
	// GetCollector returns domain.ErrCollectorNotFound if no collector with the given id exists.
	GetCollector(ctx context.Context, id uuid.UUID) (*domain.Collector, error)
	// SaveCollector inserts the collector or updates it if its id already exists.
	SaveCollector(ctx context.Context, collector *domain.Collector) error
}
