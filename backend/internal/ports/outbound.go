package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// CollectorRepository persists collectors.
type CollectorRepository interface {
	// GetCollector returns domain.ErrCollectorNotFound if no collector with the given id exists.
	GetCollector(ctx context.Context, id domain.CollectorID) (*domain.Collector, error)
	// SaveCollector inserts a new collector. It returns domain.ErrCollectorIDTaken if the id is already in use.
	SaveCollector(ctx context.Context, collector *domain.Collector) error
}

// PersonaRepository persists personas together with their item definitions, which are part of the persona.
type PersonaRepository interface {
	// GetPersona returns the persona with all its items in their stored order.
	// It returns domain.ErrPersonaNotFound if the collector has no persona with the given id,
	// which includes personas that belong to another collector.
	GetPersona(ctx context.Context, collectorID domain.CollectorID, id uuid.UUID) (*domain.Persona, error)
	// ListPersonas returns all personas of the collector with their items, ordered by name and then id.
	// A collector without personas yields an empty list.
	// It returns domain.ErrCollectorNotFound if the collector does not exist.
	ListPersonas(ctx context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error)
	// SavePersona inserts the persona or, if its id exists already, replaces name, description and items.
	// Items missing from persona.Items are deleted, the others keep their id.
	// It returns domain.ErrPersonaNotFound if the id belongs to another collector,
	// domain.ErrItemIDTaken if an item id is already used by another persona and
	// domain.ErrUnknownCollector if the persona's collector does not exist. Nothing is changed then.
	//
	// There is no concurrency control: SavePersona replaces the whole persona, so the last save wins. Two
	// people editing the same persona from stale copies silently overwrite each other, and a stale save
	// also deletes the items the other person added meanwhile. That is acceptable while personas are edited
	// rarely by one person per collector. Before that changes, or before client data refers to item
	// definitions (deleting one would then lose that data), add optimistic locking: a persona version that
	// the client sends back, saved with "UPDATE ... WHERE version = $n" and answered with 409 on a mismatch.
	SavePersona(ctx context.Context, persona *domain.Persona) error
}
