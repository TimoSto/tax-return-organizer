package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// ClientTaxYear is one client's tax year — the container documents and
// checklist item completions for that year hang off.
type ClientTaxYear struct {
	ID       int
	ClientID uuid.UUID
	Year     int
}

func NewClientTaxYear(clientID uuid.UUID, year int) (*ClientTaxYear, error) {
	if year < 1900 || year > 2200 {
		return nil, fmt.Errorf("%w: %d", ErrTaxYearOutOfRange, year)
	}
	return &ClientTaxYear{ClientID: clientID, Year: year}, nil
}
