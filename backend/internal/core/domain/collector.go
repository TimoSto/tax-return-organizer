package domain

import (
	"errors"

	"github.com/google/uuid"
)

var ErrCollectorNotFound = errors.New("collector not found")

type Collector struct {
	ID    uuid.UUID
	Name  string
	Email string
}
