// Package inbound defines the ports through which the outside world drives the core: one interface per use case,
// and the input types they need. Inbound adapters (HTTP) depend on these interfaces only, never on a use case
// package; the use cases implement them and the composition root wires the two together.
//
// Use cases that act within a tax practice take the caller as a domain.Actor and act on its tax practice
// (Actor.TaxPracticeID); they decide what the actor may do. The inbound adapter only supplies the actor.
//
// Errors are the ones documented on the domain (see domain.Error): invalid input, not found, conflict and
// forbidden are meant for the client, anything else is an internal failure.
package inbound

import (
	"context"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// CreateTaxPracticeCommand is the input of CreateTaxPractice.
type CreateTaxPracticeCommand struct {
	Name  string
	Email string
	// Locale is the language of the predefined personas, empty for the default.
	Locale string
}

// CreateTaxPractice creates a tax practice together with the predefined personas and returns its id.
// Nothing is stored if it fails.
type CreateTaxPractice interface {
	Execute(ctx context.Context, cmd CreateTaxPracticeCommand) (domain.TaxPracticeID, error)
}

// GetTaxPractice retrieves the tax practice the actor belongs to. Staff and clients may.
// It returns domain.ErrTaxPracticeNotFound if there is none.
type GetTaxPractice interface {
	Execute(ctx context.Context, actor domain.Actor) (*domain.TaxPractice, error)
}

// GetPersona retrieves a persona of the actor's tax practice by its id. Only staff may.
// It returns domain.ErrForbidden for any other actor, domain.ErrTaxPracticeNotFound if the tax practice does not exist,
// and domain.ErrPersonaNotFound if the tax practice has no such persona, which includes personas of other tax practices.
type GetPersona interface {
	Execute(ctx context.Context, actor domain.Actor, id domain.PersonaID) (*domain.Persona, error)
}

// ListPersonas retrieves all personas of the actor's tax practice. Only staff may.
// It returns domain.ErrForbidden for any other actor and domain.ErrTaxPracticeNotFound if the tax practice does not
// exist.
type ListPersonas interface {
	Execute(ctx context.Context, actor domain.Actor) ([]*domain.Persona, error)
}
