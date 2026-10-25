// Package inbound defines the ports through which the outside world drives the core: one interface per use case,
// and the input types they need. Inbound adapters (HTTP) depend on these interfaces only, never on a use case
// package; the use cases implement them and the composition root wires the two together.
//
// Errors are the ones documented on the domain (see domain.Error): invalid input, not found and conflict are
// meant for the client, anything else is an internal failure.
package inbound

import (
	"context"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// CreateCollectorCommand is the input of CreateCollector.
type CreateCollectorCommand struct {
	Name  string
	Email string
	// Locale is the language of the predefined personas, empty for the default.
	Locale string
}

// CreateCollector creates a collector together with the predefined personas and returns its id.
// Nothing is stored if it fails.
type CreateCollector interface {
	Execute(ctx context.Context, cmd CreateCollectorCommand) (domain.CollectorID, error)
}

// GetCollector retrieves the collector with the given id.
// It returns domain.ErrCollectorNotFound if there is none.
type GetCollector interface {
	Execute(ctx context.Context, id domain.CollectorID) (*domain.Collector, error)
}

// GetPersona retrieves a persona of the collector by its id.
// It returns domain.ErrCollectorNotFound if the collector does not exist, and domain.ErrPersonaNotFound if the
// collector has no such persona, which includes personas of other collectors.
type GetPersona interface {
	Execute(ctx context.Context, collectorID domain.CollectorID, id domain.PersonaID) (*domain.Persona, error)
}

// ListPersonas retrieves all personas of the collector.
// It returns domain.ErrCollectorNotFound if that collector does not exist.
type ListPersonas interface {
	Execute(ctx context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error)
}
