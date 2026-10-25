// Package collector contains the in-memory collector repository for tests.
package collector

import (
	"context"
	"sync"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

var _ outbound.CollectorRepository = (*Repository)(nil)

// Repository is an in-memory outbound.CollectorRepository.
type Repository struct {
	mu         sync.RWMutex
	collectors map[domain.CollectorID]domain.Collector
}

// NewRepository returns an empty repository.
func NewRepository() *Repository {
	return &Repository{collectors: make(map[domain.CollectorID]domain.Collector)}
}

// GetCollector implements outbound.CollectorRepository.
func (r *Repository) GetCollector(_ context.Context, id domain.CollectorID) (*domain.Collector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.collectors[id]
	if !ok {
		return nil, domain.ErrCollectorNotFound
	}

	return &c, nil
}

// SaveCollector implements outbound.CollectorRepository.
func (r *Repository) SaveCollector(_ context.Context, collector *domain.Collector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, taken := r.collectors[collector.ID()]; taken {
		return domain.ErrCollectorIDTaken
	}

	r.collectors[collector.ID()] = *collector

	return nil
}
