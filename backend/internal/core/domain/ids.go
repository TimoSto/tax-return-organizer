package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// ErrInvalidPersonaID is returned for a string that is not a well-formed persona id.
var ErrInvalidPersonaID = invalid("invalid persona id")

// PersonaID identifies a persona.
type PersonaID uuid.UUID

// ParsePersonaID validates s and returns it as a PersonaID.
func ParsePersonaID(s string) (PersonaID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return PersonaID{}, ErrInvalidPersonaID
	}

	return PersonaID(id), nil
}

// String implements fmt.Stringer.
func (id PersonaID) String() string { return uuid.UUID(id).String() }

// ItemID identifies an item definition.
type ItemID uuid.UUID

// String implements fmt.Stringer.
func (id ItemID) String() string { return uuid.UUID(id).String() }

// IDGenerator creates the ids of new entities, so that the domain stays deterministic and callers can control ids.
type IDGenerator interface {
	NewCollectorID() (CollectorID, error)
	NewPersonaID() (PersonaID, error)
	NewItemID() (ItemID, error)
}

// RandomIDs is the IDGenerator that draws random ids. A collector id is not guaranteed to be unused.
type RandomIDs struct{}

// NewCollectorID returns a random, short collector id.
func (RandomIDs) NewCollectorID() (CollectorID, error) { return NewCollectorID() }

// NewPersonaID returns a random persona id.
func (RandomIDs) NewPersonaID() (PersonaID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return PersonaID{}, fmt.Errorf("generate persona id: %w", err)
	}

	return PersonaID(id), nil
}

// NewItemID returns a random item id.
func (RandomIDs) NewItemID() (ItemID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return ItemID{}, fmt.Errorf("generate item id: %w", err)
	}

	return ItemID(id), nil
}
