// Package persona contains the in-memory persona repository for tests.
package persona

import (
	"bytes"
	"cmp"
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
	mu              sync.RWMutex
	collectorExists func(context.Context, domain.CollectorID) bool
	personas        map[uuid.UUID]domain.Persona
}

// NewRepository returns an empty repository. Personas can only be saved for the given collectors,
// as a foreign key does in a relational store.
func NewRepository(collectors ...domain.CollectorID) *Repository {
	return &Repository{
		collectorExists: func(_ context.Context, id domain.CollectorID) bool { return slices.Contains(collectors, id) },
		personas:        make(map[uuid.UUID]domain.Persona),
	}
}

// NewRepositoryFor returns an empty repository that accepts personas of the collectors stored in collectors.
func NewRepositoryFor(collectors ports.CollectorRepository) *Repository {
	return &Repository{
		collectorExists: func(ctx context.Context, id domain.CollectorID) bool {
			_, err := collectors.GetCollector(ctx, id)

			return err == nil
		},
		personas: make(map[uuid.UUID]domain.Persona),
	}
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

// ListPersonas implements ports.PersonaRepository.
func (r *Repository) ListPersonas(ctx context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.collectorExists(ctx, collectorID) {
		return nil, domain.ErrCollectorNotFound
	}

	personas := make([]*domain.Persona, 0)

	for _, p := range r.personas {
		if p.CollectorID == collectorID {
			personas = append(personas, clone(&p))
		}
	}

	slices.SortFunc(personas, func(a, b *domain.Persona) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), bytes.Compare(a.ID[:], b.ID[:]))
	})

	return personas, nil
}

// SavePersona implements ports.PersonaRepository.
func (r *Repository) SavePersona(ctx context.Context, persona *domain.Persona) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.collectorExists(ctx, persona.CollectorID) {
		return domain.ErrUnknownCollector
	}

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
			// Like the domain and a relational store, report no fields as nil, never as an empty slice.
			if len(spec.MetadataFields) == 0 {
				spec.MetadataFields = nil
			}

			item.DocumentSpec = &spec
		}

		if item.ValueSpec != nil {
			spec := *item.ValueSpec
			item.ValueSpec = &spec
		}
	}

	return &c
}
