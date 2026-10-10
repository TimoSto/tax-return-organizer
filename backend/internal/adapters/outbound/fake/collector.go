// Package fake contains in-memory implementations of the outbound ports for tests.
package fake

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

var _ ports.CollectorRepository = (*CollectorRepository)(nil)

type CollectorRepository struct {
	mu         sync.RWMutex
	collectors map[uuid.UUID]domain.Collector
}

func NewCollectorRepository() *CollectorRepository {
	return &CollectorRepository{collectors: make(map[uuid.UUID]domain.Collector)}
}

func (r *CollectorRepository) GetCollector(_ context.Context, id uuid.UUID) (*domain.Collector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.collectors[id]
	if !ok {
		return nil, domain.ErrCollectorNotFound
	}

	return &c, nil
}

func (r *CollectorRepository) SaveCollector(_ context.Context, collector *domain.Collector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.collectors[collector.ID] = *collector

	return nil
}
