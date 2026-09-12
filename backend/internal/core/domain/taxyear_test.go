package domain_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewTaxYear_RejectsOutOfRange(t *testing.T) {
	cases := []int{1899, 2201, 0, -1}
	for _, year := range cases {
		if _, err := domain.NewTaxYear(year); err == nil {
			t.Errorf("expected error for year %d, got nil", year)
		}
	}
}

func TestNewTaxYear_AcceptsPlausibleYear(t *testing.T) {
	ty, err := domain.NewTaxYear(2025)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ty.Year != 2025 {
		t.Fatalf("expected year 2025, got %d", ty.Year)
	}
}
