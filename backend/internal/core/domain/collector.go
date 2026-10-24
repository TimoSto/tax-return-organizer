// Package domain contains the business entities and their rules. It depends on no other package of this module.
package domain

import (
	"strings"
)

var (
	// ErrCollectorNotFound is returned when a collector does not exist.
	ErrCollectorNotFound = notFound("collector not found")
	// ErrCollectorIDTaken is returned when saving a collector whose id is already in use.
	ErrCollectorIDTaken = conflict("collector id already taken")
	// ErrNameRequired is returned for a blank collector name.
	ErrNameRequired = invalid("collector name is required")
)

// Collector is a person who collects documents for a tax return.
type Collector struct {
	ID    CollectorID
	Name  string
	Email string
}

// NewCollector validates the input and returns a collector with the given id.
func NewCollector(id CollectorID, name, email string) (*Collector, error) {
	name, email = strings.TrimSpace(name), strings.TrimSpace(email)

	if name == "" {
		return nil, ErrNameRequired
	}

	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	return &Collector{ID: id, Name: name, Email: email}, nil
}
