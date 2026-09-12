package domain

import "fmt"

type TaxYear struct {
	ID   int
	Year int
}

func NewTaxYear(year int) (*TaxYear, error) {
	if year < 1900 || year > 2200 {
		return nil, fmt.Errorf("tax year %d out of plausible range", year)
	}
	return &TaxYear{Year: year}, nil
}
