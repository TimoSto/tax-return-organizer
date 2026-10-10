package domain

import (
	"errors"
	"net/mail"
)

var ErrInvalidEmail = errors.New("invalid email address")

func ValidateEMail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.Join(ErrInvalidEmail, err)
	}

	return nil
}
