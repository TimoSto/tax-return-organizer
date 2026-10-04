package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestRequirement_Accepts(t *testing.T) {
	docID := uuid.New()

	tests := []struct {
		name    string
		req     domain.Requirement
		value   domain.Value
		wantErr error
	}{
		{"document ok", domain.DocumentRequirement{}, domain.DocumentValue{DocumentID: docID}, nil},
		{"document nil id", domain.DocumentRequirement{}, domain.DocumentValue{}, domain.ErrInvalidValue},
		{"document gets text", domain.DocumentRequirement{}, domain.TextValue{Text: "x"}, domain.ErrWrongFulfillmentKind},

		{"text ok", domain.TextRequirement{}, domain.TextValue{Text: "Berlin"}, nil},
		{"text blank", domain.TextRequirement{}, domain.TextValue{Text: "  "}, domain.ErrInvalidValue},
		{"text gets bool", domain.TextRequirement{}, domain.BoolValue{Bool: true}, domain.ErrWrongFulfillmentKind},

		{"number ok", domain.NewNumberRequirement("EUR"), domain.NumberValue{Number: 1450}, nil},
		{"number gets year", domain.NewNumberRequirement(""), domain.YearValue{Year: 2024}, domain.ErrWrongFulfillmentKind},

		{"year ok", domain.YearRequirement{}, domain.YearValue{Year: 2024}, nil},
		{"year lower bound", domain.YearRequirement{}, domain.YearValue{Year: 1900}, nil},
		{"year too early", domain.YearRequirement{}, domain.YearValue{Year: 1899}, domain.ErrInvalidValue},
		{"year too late", domain.YearRequirement{}, domain.YearValue{Year: 2201}, domain.ErrInvalidValue},
		{"year gets number", domain.YearRequirement{}, domain.NumberValue{Number: 2024}, domain.ErrWrongFulfillmentKind},

		{"bool ok", domain.BoolRequirement{}, domain.BoolValue{Bool: false}, nil},
		{"bool gets text", domain.BoolRequirement{}, domain.TextValue{Text: "yes"}, domain.ErrWrongFulfillmentKind},

		{"nil value", domain.BoolRequirement{}, nil, domain.ErrInvalidValue},

		// Pointers to the right type must be rejected, not panic.
		{"document pointer", domain.DocumentRequirement{}, &domain.DocumentValue{DocumentID: docID}, domain.ErrWrongFulfillmentKind},
		{"text pointer", domain.TextRequirement{}, &domain.TextValue{Text: "x"}, domain.ErrWrongFulfillmentKind},
		{"number pointer", domain.NewNumberRequirement(""), &domain.NumberValue{Number: 1}, domain.ErrWrongFulfillmentKind},
		{"year pointer", domain.YearRequirement{}, &domain.YearValue{Year: 2024}, domain.ErrWrongFulfillmentKind},
		{"bool pointer", domain.BoolRequirement{}, &domain.BoolValue{Bool: true}, domain.ErrWrongFulfillmentKind},
		{"nil text pointer", domain.TextRequirement{}, (*domain.TextValue)(nil), domain.ErrWrongFulfillmentKind},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Accepts(tt.value)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewNumberRequirement_TrimsUnit(t *testing.T) {
	if got := domain.NewNumberRequirement("  EUR ").Unit; got != "EUR" {
		t.Errorf("Unit = %q, want %q", got, "EUR")
	}
}
