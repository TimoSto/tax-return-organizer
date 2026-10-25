package domain

import "github.com/google/uuid"

var (
	// ErrInvalidPersonaID is returned for a string that is not a well-formed persona id.
	ErrInvalidPersonaID = invalid("invalid persona id")
	// ErrInvalidItemID is returned for a string that is not a well-formed item definition id.
	ErrInvalidItemID = invalid("invalid item id")
)

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

// ParseItemID validates s and returns it as an ItemID.
func ParseItemID(s string) (ItemID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ItemID{}, ErrInvalidItemID
	}

	return ItemID(id), nil
}

// String implements fmt.Stringer.
func (id ItemID) String() string { return uuid.UUID(id).String() }
