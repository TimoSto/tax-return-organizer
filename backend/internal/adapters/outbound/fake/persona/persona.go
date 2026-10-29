// Package persona contains the in-memory persona repository for tests.
package persona

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

var (
	_ outbound.PersonaStore      = (*Store)(nil)
	_ outbound.PersonaRepository = (*Repository)(nil)
)

// Store is an in-memory outbound.PersonaStore. A persona is stored as one object including its items.
// It emulates the foreign key of the Postgres store by looking the collector up in collectors.
type Store struct {
	mu         sync.RWMutex
	personas   map[domain.PersonaID]domain.Persona
	collectors outbound.CollectorRepository
}

// NewStore returns an empty store whose personas belong to the collectors in collectors.
func NewStore(collectors outbound.CollectorRepository) *Store {
	return &Store{personas: make(map[domain.PersonaID]domain.Persona), collectors: collectors}
}

// For implements outbound.PersonaStore. The repositories of all collectors share the storage of the store.
func (s *Store) For(collectorID domain.CollectorID) outbound.PersonaRepository {
	return &Repository{store: s, collectorID: collectorID}
}

// Repository is the view of a Store on the personas of one collector.
type Repository struct {
	store       *Store
	collectorID domain.CollectorID
}

// GetPersona implements outbound.PersonaRepository.
func (r *Repository) GetPersona(ctx context.Context, id domain.PersonaID) (*domain.Persona, error) {
	if err := r.checkCollector(ctx); err != nil {
		return nil, err
	}

	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	p, ok := r.store.personas[id]
	if !ok || p.CollectorID() != r.collectorID {
		return nil, domain.ErrPersonaNotFound
	}

	return clone(&p), nil
}

// ListPersonas implements outbound.PersonaRepository.
func (r *Repository) ListPersonas(ctx context.Context) ([]*domain.Persona, error) {
	if err := r.checkCollector(ctx); err != nil {
		return nil, err
	}

	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	personas := make([]*domain.Persona, 0)

	for _, p := range r.store.personas {
		if p.CollectorID() == r.collectorID {
			personas = append(personas, clone(&p))
		}
	}

	slices.SortFunc(personas, func(a, b *domain.Persona) int {
		return cmp.Or(cmp.Compare(a.Name(), b.Name()), compareIDs(a.ID(), b.ID()))
	})

	return personas, nil
}

// SavePersona implements outbound.PersonaRepository.
func (r *Repository) SavePersona(ctx context.Context, persona *domain.Persona) error {
	if persona.CollectorID() != r.collectorID {
		return fmt.Errorf("save persona %s: %w", persona.ID(), outbound.ErrTenantMismatch)
	}

	if err := r.checkCollector(ctx); err != nil {
		return err
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if existing, ok := r.store.personas[persona.ID()]; ok && existing.CollectorID() != persona.CollectorID() {
		return domain.ErrPersonaNotFound
	}

	r.store.personas[persona.ID()] = *clone(persona)

	return nil
}

// checkCollector returns domain.ErrCollectorNotFound if the bound collector does not exist.
func (r *Repository) checkCollector(ctx context.Context) error {
	_, err := r.store.collectors.GetCollector(ctx, r.collectorID)

	return err
}

// clone copies p deeply, so callers and the repository never share items or specs. Persona.Items and
// RehydratePersona copy the items, so a round trip through them is a deep copy.
func clone(p *domain.Persona) *domain.Persona {
	c, err := domain.RehydratePersona(p.ID(), p.CollectorID(), p.Name(), p.Description(), p.Items())
	if err != nil {
		// p was valid, so its copy is valid too.
		panic(err)
	}

	return c
}

func compareIDs(a, b domain.PersonaID) int { return bytes.Compare(a[:], b[:]) }
