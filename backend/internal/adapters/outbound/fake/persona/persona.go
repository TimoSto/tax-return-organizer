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
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

var (
	_ port.PersonaStore      = (*Store)(nil)
	_ port.PersonaRepository = (*Repository)(nil)
)

// Store is an in-memory port.PersonaStore. A persona is stored as one object including its items.
type Store struct {
	mu       sync.RWMutex
	personas map[domain.PersonaID]domain.Persona
}

// NewStore returns an empty store.
func NewStore() *Store {
	return &Store{personas: make(map[domain.PersonaID]domain.Persona)}
}

// For implements port.PersonaStore. The repositories of all collectors share the storage of the store.
func (s *Store) For(collectorID domain.CollectorID) port.PersonaRepository {
	return &Repository{store: s, collectorID: collectorID}
}

// Repository is the view of a Store on the personas of one collector.
type Repository struct {
	store       *Store
	collectorID domain.CollectorID
}

// GetPersona implements port.PersonaRepository.
func (r *Repository) GetPersona(_ context.Context, id domain.PersonaID) (*domain.Persona, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	p, ok := r.store.personas[id]
	if !ok || p.CollectorID() != r.collectorID {
		return nil, domain.ErrPersonaNotFound
	}

	return clone(&p), nil
}

// ListPersonas implements port.PersonaRepository.
func (r *Repository) ListPersonas(_ context.Context) ([]*domain.Persona, error) {
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

// SavePersona implements port.PersonaRepository.
func (r *Repository) SavePersona(_ context.Context, persona *domain.Persona) error {
	if persona.CollectorID() != r.collectorID {
		return fmt.Errorf("save persona %s: %w", persona.ID(), port.ErrTenantMismatch)
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if existing, ok := r.store.personas[persona.ID()]; ok && existing.CollectorID() != persona.CollectorID() {
		return domain.ErrPersonaNotFound
	}

	for id, other := range r.store.personas {
		if id == persona.ID() {
			continue
		}

		for _, item := range persona.Items() {
			if slices.ContainsFunc(other.Items(), func(o domain.ItemDefinition) bool { return o.ID() == item.ID() }) {
				return domain.ErrItemIDTaken
			}
		}
	}

	r.store.personas[persona.ID()] = *clone(persona)

	return nil
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
