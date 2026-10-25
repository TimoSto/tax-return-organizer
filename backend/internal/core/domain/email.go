package domain

import (
	"net/mail"
	"strings"
)

// ErrInvalidEmail is returned for anything but a plain email address.
var ErrInvalidEmail = invalid("invalid email address")

// Email is a validated, plain email address such as "john@example.com".
type Email string

// ParseEmail accepts only a plain address of at most MaxEmailLength characters. Display names
// ("John <john@example.com>") and angle brackets are rejected. The address is lowercased, so the same address is
// always stored the same way.
func ParseEmail(s string) (Email, error) {
	s = strings.ToLower(s)

	if len(s) > MaxEmailLength {
		return "", ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return "", ErrInvalidEmail
	}

	return Email(s), nil
}

// String implements fmt.Stringer.
func (e Email) String() string { return string(e) }
