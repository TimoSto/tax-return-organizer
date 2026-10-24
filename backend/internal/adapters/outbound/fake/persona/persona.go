// Package persona contains the in-memory persona repository for tests.
package persona

import (
	"bytes"
	"cmp"
	"context"
	"slices"
	"sync"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

var _ port.PersonaRepository = (*Repository)(nil)

// Repository is an in-memory port.PersonaRepository. A persona is stored as one object including its items.
type Repository struct {
	mu       sync.RWMutex
	personas map[domain.PersonaID]domain.Persona
}

// NewRepository returns an empty repository.
func NewRepository() *Repository {
	return &Repository{personas: make(map[domain.PersonaID]domain.Persona)}
}

// GetPersona implements port.PersonaRepository.
func (r *Repository) GetPersona(_ context.Context, collectorID domain.CollectorID, id domain.PersonaID) (*domain.Persona, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.personas[id]
	if !ok || p.CollectorID() != collectorID {
		return nil, domain.ErrPersonaNotFound
	}

	return clone(&p), nil
}

// ListPersonas implements port.PersonaRepository.
func (r *Repository) ListPersonas(_ context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	personas := make([]*domain.Persona, 0)

	for _, p := range r.personas {
		if p.CollectorID() == collectorID {
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
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.personas[persona.ID()]; ok && existing.CollectorID() != persona.CollectorID() {
		return domain.ErrPersonaNotFound
	}

	for id, other := range r.personas {
		if id == persona.ID() {
			continue
		}

		for _, item := range persona.Items() {
			if slices.ContainsFunc(other.Items(), func(o domain.ItemDefinition) bool { return o.ID() == item.ID() }) {
				return domain.ErrItemIDTaken
			}
		}
	}

	r.personas[persona.ID()] = *clone(persona)

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
