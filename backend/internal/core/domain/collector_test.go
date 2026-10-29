package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

var (
	// testCollectorID is the collector the tests act for.
	testCollectorID = domain.CollectorID(uuid.New())
)

func TestNewCollector(t *testing.T) {
	c, err := domain.NewCollector(testCollectorID, "  Jane ", " jane@example.com ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.ID() != testCollectorID || c.Name() != "Jane" || c.Email() != "jane@example.com" {
		t.Errorf("unexpected collector %+v", c)
	}

	if _, err := domain.NewCollector(testCollectorID, " ", "jane@example.com"); !errors.Is(err, domain.ErrNameRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrNameRequired)
	}

	if _, err := domain.NewCollector(testCollectorID, "Jane", "nope"); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("err = %v, want %v", err, domain.ErrInvalidEmail)
	}
}

func TestRehydrateCollector_RejectsInvalidStoredData(t *testing.T) {
	if _, err := domain.RehydrateCollector(testCollectorID, "", "jane@example.com"); !errors.Is(err, domain.ErrNameRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrNameRequired)
	}

	if _, err := domain.RehydrateCollector(testCollectorID, "Jane", "John <john@example.com>"); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("err = %v, want %v", err, domain.ErrInvalidEmail)
	}
}
