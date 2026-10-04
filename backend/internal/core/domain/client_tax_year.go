package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// Plausible range for any year the domain accepts (tax years, year values).
const (
	minYear = 1900
	maxYear = 2200
)

// ClientTaxYear is one client's tax year — the container documents and
// checklist item completions for that year hang off.
type ClientTaxYear struct {
	ID       int
	ClientID uuid.UUID
	Year     int
}

func NewClientTaxYear(clientID uuid.UUID, year int) (*ClientTaxYear, error) {
	if year < minYear || year > maxYear {
		return nil, fmt.Errorf("%w: %d", ErrTaxYearOutOfRange, year)
	}
	return &ClientTaxYear{ClientID: clientID, Year: year}, nil
}
