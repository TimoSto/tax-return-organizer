// Package domain contains the business entities and their rules. It depends on no other package of this module.
package domain

import (
	"errors"
	"strings"
)

var (
	// ErrTaxPracticeNotFound is returned when a tax practice does not exist.
	ErrTaxPracticeNotFound = notFound("tax practice not found")
	// ErrTaxPracticeIDTaken is returned when saving a tax practice whose id is already in use. TaxPractice ids are random
	// UUIDs, so a clash is a bug, not a client error: it is a plain error, not a *Error.
	ErrTaxPracticeIDTaken = errors.New("tax practice id already taken")
	// ErrNameRequired is returned for a blank tax practice name.
	ErrNameRequired = invalid("tax practice name is required")
	// ErrNameTooLong is returned for a tax practice name longer than MaxTaxPracticeNameLength.
	ErrNameTooLong = invalid("tax practice name is too long")
)

// TaxPractice is a person who collects documents for a tax return.
// Create tax practices with NewTaxPractice or RehydrateTaxPractice, so they are always valid.
type TaxPractice struct {
	id    TaxPracticeID
	name  string
	email Email
}

// NewTaxPractice validates the input and returns a tax practice with the given id.
func NewTaxPractice(id TaxPracticeID, name, email string) (*TaxPractice, error) {
	addr, err := ParseEmail(strings.TrimSpace(email))
	if err != nil {
		return nil, err
	}

	return RehydrateTaxPractice(id, name, addr)
}

// RehydrateTaxPractice rebuilds a tax practice that was stored earlier, e.g. when a repository loads it. It applies the
// same rules as NewTaxPractice, so corrupt stored data is reported instead of becoming an invalid tax practice.
func RehydrateTaxPractice(id TaxPracticeID, name string, email Email) (*TaxPractice, error) {
	if id == (TaxPracticeID{}) {
		return nil, ErrInvalidTaxPracticeID
	}

	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrNameRequired
	}

	if tooLong(name, MaxTaxPracticeNameLength) {
		return nil, ErrNameTooLong
	}

	addr, err := ParseEmail(string(email))
	if err != nil {
		return nil, err
	}

	return &TaxPractice{id: id, name: name, email: addr}, nil
}

// ID returns the id of the tax practice.
func (c *TaxPractice) ID() TaxPracticeID { return c.id }

// Name returns the name of the tax practice.
func (c *TaxPractice) Name() string { return c.name }

// Email returns the email address of the tax practice.
func (c *TaxPractice) Email() Email { return c.email }
