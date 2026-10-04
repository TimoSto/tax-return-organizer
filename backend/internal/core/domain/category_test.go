package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewCategory(t *testing.T) {
	collectorID := uuid.New()

	t.Run("valid", func(t *testing.T) {
		c, err := domain.NewCategory(collectorID, "Private finances", "bank statements, ETFs, ...")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.CollectorID != collectorID {
			t.Errorf("CollectorID = %v, want %v", c.CollectorID, collectorID)
		}
		if c.Name != "Private finances" {
			t.Errorf("Name = %q, want %q", c.Name, "Private finances")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		if _, err := domain.NewCategory(collectorID, "  ", ""); !errors.Is(err, domain.ErrEmptyCategoryName) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyCategoryName)
		}
	})
}
