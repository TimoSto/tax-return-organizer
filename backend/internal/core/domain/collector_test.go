package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewCollector(t *testing.T) {
	t.Run("valid name", func(t *testing.T) {
		c, err := domain.NewCollector("Acme Tax Advisory")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Name != "Acme Tax Advisory" {
			t.Errorf("Name = %q, want %q", c.Name, "Acme Tax Advisory")
		}
		if c.ID.String() == "" {
			t.Error("expected a generated ID")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		if _, err := domain.NewCollector("   "); !errors.Is(err, domain.ErrEmptyCollectorName) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyCollectorName)
		}
	})
}
