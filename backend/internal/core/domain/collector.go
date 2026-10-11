package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrCollectorNotFound = errors.New("collector not found")
	ErrNameRequired      = errors.New("collector name is required")
)

type Collector struct {
	ID    uuid.UUID
	Name  string
	Email string
}

// NewCollector validates the input and returns a collector with the given id.
func NewCollector(id uuid.UUID, name, email string) (*Collector, error) {
	name, email = strings.TrimSpace(name), strings.TrimSpace(email)

	if name == "" {
		return nil, ErrNameRequired
	}

	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	return &Collector{ID: id, Name: name, Email: email}, nil
}
