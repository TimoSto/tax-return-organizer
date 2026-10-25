package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestParsePersonaID(t *testing.T) {
	const s = "3f2b8c1e-5d4a-4e6f-9a7b-1c2d3e4f5a6b"

	id, err := domain.ParsePersonaID(s)
	if err != nil || id.String() != s {
		t.Errorf("ParsePersonaID(%q) = %v, %v", s, id, err)
	}

	for _, in := range []string{"", "not-a-uuid", s + "0"} {
		if _, err := domain.ParsePersonaID(in); !errors.Is(err, domain.ErrInvalidPersonaID) {
			t.Errorf("ParsePersonaID(%q) error = %v, want %v", in, err, domain.ErrInvalidPersonaID)
		}
	}
}

func TestParseItemID(t *testing.T) {
	const s = "9c1d7e2a-0b3f-4a58-8e6d-5f4c3b2a1d09"

	id, err := domain.ParseItemID(s)
	if err != nil || id.String() != s {
		t.Errorf("ParseItemID(%q) = %v, %v", s, id, err)
	}

	for _, in := range []string{"", "not-a-uuid", s + "0"} {
		if _, err := domain.ParseItemID(in); !errors.Is(err, domain.ErrInvalidItemID) {
			t.Errorf("ParseItemID(%q) error = %v, want %v", in, err, domain.ErrInvalidItemID)
		}
	}
}
