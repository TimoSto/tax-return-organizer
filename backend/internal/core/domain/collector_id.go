package domain

import "github.com/google/uuid"

// ErrInvalidCollectorID is returned for an id that is not a well-formed collector id.
var ErrInvalidCollectorID = invalid("invalid collector id")

// CollectorID identifies a collector. It is a random UUID, so it cannot be guessed.
type CollectorID uuid.UUID

// ParseCollectorID validates s and returns it as a CollectorID.
func ParseCollectorID(s string) (CollectorID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return CollectorID{}, ErrInvalidCollectorID
	}

	return CollectorID(id), nil
}

// String implements fmt.Stringer.
func (id CollectorID) String() string { return uuid.UUID(id).String() }
