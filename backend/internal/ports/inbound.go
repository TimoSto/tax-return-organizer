// Package ports defines the interfaces between the core and its adapters.
package ports

import (
	"context"

	"github.com/google/uuid"
)

// CreateCollectorCmd is the input of the CreateCollector use case.
type CreateCollectorCmd struct {
	Name  string
	Email string
}

// CreateCollector creates a collector and returns its id.
type CreateCollector interface {
	Execute(ctx context.Context, cmd CreateCollectorCmd) (uuid.UUID, error)
}
