package domain

import (
	"net/mail"
)

// ErrInvalidEmail is returned for anything but a plain email address.
var ErrInvalidEmail = invalid("invalid email address")

// ValidateEmail accepts only a plain address such as "john@example.com".
// Display names ("John <john@example.com>") and angle brackets are rejected.
func ValidateEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return ErrInvalidEmail
	}

	return nil
}
