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
	// SavePersona creates the persona if persona.Version is 0, otherwise it replaces name, description and
	// items of the stored persona, which must still have that version. Items missing from persona.Items are
	// deleted, the others keep their id. On success persona.Version is the new stored version.
	// It returns domain.ErrPersonaNotFound if the id belongs to another collector,
	// domain.ErrItemIDTaken if an item id is already used by another persona and
	// domain.ErrUnknownCollector if the persona's collector does not exist, and domain.ErrVersionConflict
	// if the stored version differs or a persona with the id exists already although Version is 0.
	// Nothing is changed then.
	SavePersona(ctx context.Context, persona *domain.Persona) error
}
