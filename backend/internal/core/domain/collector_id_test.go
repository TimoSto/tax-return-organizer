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
		{"aaaa", nil},
		{"k7m2", nil},
		{"zz99", nil},
		{"", domain.ErrInvalidCollectorID},
		{"aaa", domain.ErrInvalidCollectorID},
		{"aaaaa", domain.ErrInvalidCollectorID},
		{"AAAA", domain.ErrInvalidCollectorID},
		{"aa0a", domain.ErrInvalidCollectorID},
		{"aa1a", domain.ErrInvalidCollectorID},
		{"aaoa", domain.ErrInvalidCollectorID},
		{"aa-a", domain.ErrInvalidCollectorID},
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
