// Package collector contains the in-memory collector repository for tests.
package collector

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

var _ ports.CollectorRepository = (*Repository)(nil)

// Repository stores collectors in memory.
type Repository struct {
	mu         sync.RWMutex
	collectors map[uuid.UUID]domain.Collector
}

func NewRepository() *Repository {
	return &Repository{collectors: make(map[uuid.UUID]domain.Collector)}
}

func (r *Repository) GetCollector(_ context.Context, id uuid.UUID) (*domain.Collector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.collectors[id]
	if !ok {
		return nil, domain.ErrCollectorNotFound
	}

	return &c, nil
}

func (r *Repository) SaveCollector(_ context.Context, collector *domain.Collector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.collectors[collector.ID] = *collector

	return nil
}
