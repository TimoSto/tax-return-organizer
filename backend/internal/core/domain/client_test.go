package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewClient(t *testing.T) {
	collectorID := uuid.New()

	t.Run("valid name", func(t *testing.T) {
		c, err := domain.NewClient(collectorID, "Jane Doe")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.CollectorID != collectorID {
			t.Errorf("CollectorID = %v, want %v", c.CollectorID, collectorID)
		}
		if c.Name != "Jane Doe" {
			t.Errorf("Name = %q, want %q", c.Name, "Jane Doe")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		if _, err := domain.NewClient(collectorID, ""); !errors.Is(err, domain.ErrEmptyClientName) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyClientName)
		}
	})
}
