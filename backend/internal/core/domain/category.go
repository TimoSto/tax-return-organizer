package domain

import (
	"strings"

	"github.com/google/uuid"
)

// Category groups checklist templates (e.g. "Private finances", "Work") and
// is defined by a Collector, reused across that collector's clients.
type Category struct {
	ID          int
	CollectorID uuid.UUID
	Name        string
	Description string
}

func NewCategory(collectorID uuid.UUID, name, description string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyCategoryName
	}
	return &Category{CollectorID: collectorID, Name: name, Description: description}, nil
}
