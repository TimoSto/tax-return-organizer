package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewTemplateYearDeadline(t *testing.T) {
	dueDate := time.Date(2025, time.May, 31, 0, 0, 0, 0, time.UTC)

	t.Run("valid once (nil month)", func(t *testing.T) {
		d, err := domain.NewTemplateYearDeadline(42, 2024, nil, dueDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.TemplateID != 42 || d.Year != 2024 || d.Month != nil || !d.DueDate.Equal(dueDate) {
			t.Errorf("d = %+v, want TemplateID=42 Year=2024 Month=nil DueDate=%v", d, dueDate)
		}
	})

	t.Run("valid monthly", func(t *testing.T) {
		month := 3
		d, err := domain.NewTemplateYearDeadline(42, 2024, &month, dueDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.Month == nil || *d.Month != 3 {
			t.Errorf("Month = %v, want 3", d.Month)
		}
	})

	yearTests := []struct {
		name    string
		year    int
		wantErr bool
	}{
		{"lower bound", 1900, false},
		{"upper bound", 2200, false},
		{"too early", 1899, true},
		{"too late", 2201, true},
	}
	for _, tt := range yearTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewTemplateYearDeadline(42, tt.year, nil, dueDate)
			if tt.wantErr {
				if !errors.Is(err, domain.ErrTaxYearOutOfRange) {
					t.Errorf("err = %v, want %v", err, domain.ErrTaxYearOutOfRange)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	monthTests := []struct {
		name    string
		month   int
		wantErr bool
	}{
		{"lower bound", 1, false},
		{"upper bound", 12, false},
		{"too low", 0, true},
		{"too high", 13, true},
	}
	for _, tt := range monthTests {
		t.Run(tt.name, func(t *testing.T) {
			month := tt.month
			_, err := domain.NewTemplateYearDeadline(42, 2024, &month, dueDate)
			if tt.wantErr {
				if !errors.Is(err, domain.ErrInvalidMonth) {
					t.Errorf("err = %v, want %v", err, domain.ErrInvalidMonth)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	t.Run("zero due date", func(t *testing.T) {
		if _, err := domain.NewTemplateYearDeadline(42, 2024, nil, time.Time{}); !errors.Is(err, domain.ErrMissingDeadlineDate) {
			t.Errorf("err = %v, want %v", err, domain.ErrMissingDeadlineDate)
		}
	})
}
