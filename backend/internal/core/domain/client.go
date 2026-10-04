package domain

import (
	"strings"

	"github.com/google/uuid"
)

// Client belongs to exactly one Collector.
type Client struct {
	ID          uuid.UUID
	CollectorID uuid.UUID
	Name        string
}

func NewClient(collectorID uuid.UUID, name string) (*Client, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyClientName
	}
	return &Client{ID: uuid.New(), CollectorID: collectorID, Name: name}, nil
}
