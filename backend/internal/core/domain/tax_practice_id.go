package domain

import "github.com/google/uuid"

// ErrInvalidTaxPracticeID is returned for an id that is not a well-formed tax practice id.
var ErrInvalidTaxPracticeID = invalid("invalid tax practice id")

// TaxPracticeID identifies a tax practice. It is a random UUID, so it cannot be guessed.
type TaxPracticeID uuid.UUID

// ParseTaxPracticeID validates s and returns it as a TaxPracticeID.
func ParseTaxPracticeID(s string) (TaxPracticeID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TaxPracticeID{}, ErrInvalidTaxPracticeID
	}

	return TaxPracticeID(id), nil
}

// String implements fmt.Stringer.
func (id TaxPracticeID) String() string { return uuid.UUID(id).String() }
