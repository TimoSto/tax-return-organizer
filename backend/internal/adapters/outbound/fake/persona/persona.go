// Package persona contains the in-memory persona repository for tests.
package persona

import (
	"context"
	"slices"
	"sync"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

var _ ports.PersonaRepository = (*Repository)(nil)

// Repository is an in-memory ports.PersonaRepository. A persona is stored as one object including its items.
type Repository struct {
	mu       sync.RWMutex
	personas map[uuid.UUID]domain.Persona
}

// NewRepository returns an empty repository.
func NewRepository() *Repository {
	return &Repository{personas: make(map[uuid.UUID]domain.Persona)}
}

// GetPersona implements ports.PersonaRepository.
func (r *Repository) GetPersona(_ context.Context, collectorID domain.CollectorID, id uuid.UUID) (*domain.Persona, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.personas[id]
	if !ok || p.CollectorID != collectorID {
		return nil, domain.ErrPersonaNotFound
	}

	return clone(&p), nil
}

// SavePersona implements ports.PersonaRepository.
func (r *Repository) SavePersona(_ context.Context, persona *domain.Persona) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.personas[persona.ID]; ok && existing.CollectorID != persona.CollectorID {
		return domain.ErrPersonaNotFound
	}

	for id, other := range r.personas {
		if id == persona.ID {
			continue
		}

		for _, item := range persona.Items {
			if slices.ContainsFunc(other.Items, func(o domain.ItemDefinition) bool { return o.ID == item.ID }) {
				return domain.ErrItemIDTaken
			}
		}
	}

	r.personas[persona.ID] = *clone(persona)

	return nil
}

// clone copies p deeply, so callers and the repository never share items or specs.
// Items are normalised to belong to the persona, as they do in a relational store.
func clone(p *domain.Persona) *domain.Persona {
	c := *p
	c.Items = slices.Clone(p.Items)

	for i := range c.Items {
		item := &c.Items[i]
		item.PersonaID = p.ID

		if item.DocumentSpec != nil {
			spec := *item.DocumentSpec
			spec.MetadataFields = slices.Clone(spec.MetadataFields)
			item.DocumentSpec = &spec
		}

		if item.ValueSpec != nil {
			spec := *item.ValueSpec
			item.ValueSpec = &spec
		}
	}

	return &c
}
