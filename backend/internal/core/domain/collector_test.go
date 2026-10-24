package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewCollector(t *testing.T) {
	c, err := domain.NewCollector("aaaa", "  Jane ", " jane@example.com ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.ID() != "aaaa" || c.Name() != "Jane" || c.Email() != "jane@example.com" {
		t.Errorf("unexpected collector %+v", c)
	}

	if _, err := domain.NewCollector("aaaa", " ", "jane@example.com"); !errors.Is(err, domain.ErrNameRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrNameRequired)
	}

	if _, err := domain.NewCollector("aaaa", "Jane", "nope"); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("err = %v, want %v", err, domain.ErrInvalidEmail)
	}
}

func TestRehydrateCollector_RejectsInvalidStoredData(t *testing.T) {
	if _, err := domain.RehydrateCollector("aaaa", "", "jane@example.com"); !errors.Is(err, domain.ErrNameRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrNameRequired)
	}

	if _, err := domain.RehydrateCollector("aaaa", "Jane", "John <john@example.com>"); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("err = %v, want %v", err, domain.ErrInvalidEmail)
	}
}
