package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestValidateEMail(t *testing.T) {
	for _, tc := range []struct {
		name   string
		email  string
		expErr error
	}{
		{
			name:   "Valid email",
			email:  "john.doe@example.com",
			expErr: nil,
		},
		{
			name:   "Valid email with plus tag",
			email:  "john+tax@example.com",
			expErr: nil,
		},
		{
			name:   "Valid email with subdomain",
			email:  "a@mail.example.co.uk",
			expErr: nil,
		},
		{
			name:   "Valid email with display name",
			email:  "John <john@example.com>",
			expErr: nil,
		},
		{
			name:   "Empty string",
			email:  "",
			expErr: domain.ErrInvalidEmail,
		},
		{
			name:   "Missing at sign",
			email:  "john.doe.example.com",
			expErr: domain.ErrInvalidEmail,
		},
		{
			name:   "Missing local part",
			email:  "@example.com",
			expErr: domain.ErrInvalidEmail,
		},
		{
			name:   "Missing domain",
			email:  "john@",
			expErr: domain.ErrInvalidEmail,
		},
		{
			name:   "Space in local part",
			email:  "john doe@example.com",
			expErr: domain.ErrInvalidEmail,
		},
		{
			name:   "Double at sign",
			email:  "a@@example.com",
			expErr: domain.ErrInvalidEmail,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateEMail(tc.email)

			if !errors.Is(err, tc.expErr) {
				t.Errorf("ValidateEMail(%q) error = %v, wantErr %v", tc.email, err, tc.expErr)
			}
		})
	}
}
