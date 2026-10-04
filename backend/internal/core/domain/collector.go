package domain

import (
	"strings"

	"github.com/google/uuid"
)

// Collector is a tenant: e.g. a tax advisory that manages its own set of
// clients, categories, and checklist templates.
type Collector struct {
	ID   uuid.UUID
	Name string
}

func NewCollector(name string) (*Collector, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyCollectorName
	}
	return &Collector{ID: uuid.New(), Name: name}, nil
}
