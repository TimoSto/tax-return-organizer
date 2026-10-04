package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewClientTaxYear(t *testing.T) {
	clientID := uuid.New()

	tests := []struct {
		name    string
		year    int
		wantErr bool
	}{
		{"typical year", 2024, false},
		{"lower bound", 1900, false},
		{"upper bound", 2200, false},
		{"too early", 1899, true},
		{"too late", 2201, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ty, err := domain.NewClientTaxYear(clientID, tt.year)
			if tt.wantErr {
				if !errors.Is(err, domain.ErrTaxYearOutOfRange) {
					t.Errorf("err = %v, want %v", err, domain.ErrTaxYearOutOfRange)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ty.ClientID != clientID {
				t.Errorf("ClientID = %v, want %v", ty.ClientID, clientID)
			}
			if ty.Year != tt.year {
				t.Errorf("Year = %d, want %d", ty.Year, tt.year)
			}
		})
	}
}
