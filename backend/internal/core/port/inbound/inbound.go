// Package inbound defines the ports through which the outside world drives the core: one interface per use case,
// and the input types they need. Inbound adapters (HTTP) depend on these interfaces only, never on a use case
// package; the use cases implement them and the composition root wires the two together.
//
// Use cases that act within a collector take the caller as a domain.Actor and act on its collector
// (Actor.CollectorID); they decide what the actor may do. The inbound adapter only supplies the actor.
//
// Errors are the ones documented on the domain (see domain.Error): invalid input, not found, conflict and
// forbidden are meant for the client, anything else is an internal failure.
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

// GetCollector retrieves the collector the actor belongs to. Staff and clients may.
// It returns domain.ErrCollectorNotFound if there is none.
type GetCollector interface {
	Execute(ctx context.Context, actor domain.Actor) (*domain.Collector, error)
}

// GetPersona retrieves a persona of the actor's collector by its id. Only staff may.
// It returns domain.ErrForbidden for any other actor, domain.ErrCollectorNotFound if the collector does not exist,
// and domain.ErrPersonaNotFound if the collector has no such persona, which includes personas of other collectors.
type GetPersona interface {
	Execute(ctx context.Context, actor domain.Actor, id domain.PersonaID) (*domain.Persona, error)
}

// ListPersonas retrieves all personas of the actor's collector. Only staff may.
// It returns domain.ErrForbidden for any other actor and domain.ErrCollectorNotFound if the collector does not
// exist.
type ListPersonas interface {
	Execute(ctx context.Context, actor domain.Actor) ([]*domain.Persona, error)
}
