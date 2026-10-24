package domain

import (
	"net/mail"
)

// ErrInvalidEmail is returned for anything but a plain email address.
var ErrInvalidEmail = invalid("invalid email address")

// Email is a validated, plain email address such as "john@example.com".
type Email string

// ParseEmail accepts only a plain address. Display names ("John <john@example.com>") and angle brackets are rejected.
func ParseEmail(s string) (Email, error) {
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return "", ErrInvalidEmail
	}

	return Email(s), nil
}

// String implements fmt.Stringer.
func (e Email) String() string { return string(e) }
