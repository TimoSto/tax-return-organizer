package ports

import (
	"context"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// CollectorRepository persists collectors.
type CollectorRepository interface {
	// GetCollector returns domain.ErrCollectorNotFound if no collector with the given id exists.
	GetCollector(ctx context.Context, id domain.CollectorID) (*domain.Collector, error)
	// SaveCollector inserts a new collector. It returns domain.ErrCollectorIDTaken if the id is already in use.
	SaveCollector(ctx context.Context, collector *domain.Collector) error
}
