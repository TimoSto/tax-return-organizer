package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// ErrPersonaNameRequired is returned for a blank persona name.
var ErrPersonaNameRequired = errors.New("persona name is required")

// Persona groups the items a client has to provide if they do something tax relevant, e.g. "Employed".
type Persona struct {
	ID          uuid.UUID
	CollectorID CollectorID
	Name        string
	Description string
}

// NewPersona validates the input and returns a persona with a fresh id.
func NewPersona(collectorID CollectorID, name, description string) (*Persona, error) {
	name, description = strings.TrimSpace(name), strings.TrimSpace(description)

	if name == "" {
		return nil, ErrPersonaNameRequired
	}

	return &Persona{ID: uuid.New(), CollectorID: collectorID, Name: name, Description: description}, nil
}
