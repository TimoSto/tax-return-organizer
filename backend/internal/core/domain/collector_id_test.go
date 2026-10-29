package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestParseCollectorID(t *testing.T) {
	tests := []struct {
		in      string
		wantErr error
	}{
		{"5f0c8a3e-9b1d-4c2e-8f7a-6d4b3c2a1e0f", nil},
		{"", domain.ErrInvalidCollectorID},
		{"aaaa", domain.ErrInvalidCollectorID},
		{"5f0c8a3e-9b1d-4c2e-8f7a-6d4b3c2a1e0", domain.ErrInvalidCollectorID},
		{"5f0c8a3e-9b1d-4c2e-8f7a-6d4b3c2a1e0g", domain.ErrInvalidCollectorID},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := domain.ParseCollectorID(tt.in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}

			if err == nil && got.String() != tt.in {
				t.Errorf("id = %q, want %q", got, tt.in)
			}
		})
	}
}
