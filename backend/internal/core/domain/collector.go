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
// Create collectors with NewCollector or RehydrateCollector, so they are always valid.
type Collector struct {
	id    CollectorID
	name  string
	email Email
}

// NewCollector validates the input and returns a collector with the given id.
func NewCollector(id CollectorID, name, email string) (*Collector, error) {
	addr, err := ParseEmail(strings.TrimSpace(email))
	if err != nil {
		return nil, err
	}

	return RehydrateCollector(id, name, addr)
}

// RehydrateCollector rebuilds a collector that was stored earlier, e.g. when a repository loads it. It applies the
// same rules as NewCollector, so corrupt stored data is reported instead of becoming an invalid collector.
func RehydrateCollector(id CollectorID, name string, email Email) (*Collector, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrNameRequired
	}

	addr, err := ParseEmail(string(email))
	if err != nil {
		return nil, err
	}

	return &Collector{id: id, name: name, email: addr}, nil
}

// ID returns the id of the collector.
func (c *Collector) ID() CollectorID { return c.id }

// Name returns the name of the collector.
func (c *Collector) Name() string { return c.name }

// Email returns the email address of the collector.
func (c *Collector) Email() Email { return c.email }
