package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

var (
	// testTaxPracticeID is the tax practice the tests act for.
	testTaxPracticeID = domain.TaxPracticeID(uuid.New())
)

func TestNewTaxPractice(t *testing.T) {
	c, err := domain.NewTaxPractice(testTaxPracticeID, "  Jane ", " jane@example.com ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.ID() != testTaxPracticeID || c.Name() != "Jane" || c.Email() != "jane@example.com" {
		t.Errorf("unexpected tax practice %+v", c)
	}

	if _, err := domain.NewTaxPractice(testTaxPracticeID, " ", "jane@example.com"); !errors.Is(err, domain.ErrNameRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrNameRequired)
	}

	if _, err := domain.NewTaxPractice(testTaxPracticeID, "Jane", "nope"); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("err = %v, want %v", err, domain.ErrInvalidEmail)
	}
}

func TestRehydrateTaxPractice_RejectsInvalidStoredData(t *testing.T) {
	if _, err := domain.RehydrateTaxPractice(testTaxPracticeID, "", "jane@example.com"); !errors.Is(err, domain.ErrNameRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrNameRequired)
	}

	if _, err := domain.RehydrateTaxPractice(testTaxPracticeID, "Jane", "John <john@example.com>"); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("err = %v, want %v", err, domain.ErrInvalidEmail)
	}
}
