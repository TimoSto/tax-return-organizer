package domain

import (
	"errors"
	"net/mail"
)

// ErrInvalidEmail is returned for anything but a plain email address.
var ErrInvalidEmail = errors.New("invalid email address")

// ValidateEmail accepts only a plain address such as "john@example.com".
// Display names ("John <john@example.com>") and angle brackets are rejected.
func ValidateEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return errors.Join(ErrInvalidEmail, err)
	}

	if addr.Address != email {
		return errors.Join(ErrInvalidEmail, errors.New("must be a plain address without display name"))
	}

	return nil
}
