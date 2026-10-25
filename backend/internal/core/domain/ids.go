package domain

import "github.com/google/uuid"

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
