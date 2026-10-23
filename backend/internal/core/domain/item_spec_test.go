package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestValueType_Valid(t *testing.T) {
	for _, tc := range []struct {
		in   domain.ValueType
		want bool
	}{
		{domain.ValueTypeText, true},
		{domain.ValueTypeNumber, true},
		{domain.ValueTypeDate, true},
		{domain.ValueTypeBool, true},
		{"", false},
		{"EUR", false},
		{"Text", false},
	} {
		if got := tc.in.Valid(); got != tc.want {
			t.Errorf("ValueType(%q).Valid() = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestValueSpec_Validate(t *testing.T) {
	if err := (domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true}).Validate(); err != nil {
		t.Errorf("valid spec: unexpected error %v", err)
	}

	if err := (domain.ValueSpec{Type: "decimal"}).Validate(); !errors.Is(err, domain.ErrInvalidValueType) {
		t.Errorf("err = %v, want %v", err, domain.ErrInvalidValueType)
	}
}

func TestDocumentSpec_Validate(t *testing.T) {
	text := domain.ValueSpec{Type: domain.ValueTypeText}

	for _, tc := range []struct {
		name   string
		spec   domain.DocumentSpec
		expErr error
	}{
		{name: "No metadata fields", spec: domain.DocumentSpec{}},
		{
			name: "Valid fields",
			spec: domain.DocumentSpec{MetadataFields: []domain.MetadataField{
				{Name: "Employer", ValueSpec: text},
				{Name: "Period start", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeDate, Required: true}},
			}},
		},
		{
			name:   "Blank name",
			spec:   domain.DocumentSpec{MetadataFields: []domain.MetadataField{{Name: "  ", ValueSpec: text}}},
			expErr: domain.ErrMetadataFieldNameRequired,
		},
		{
			name: "Duplicate name",
			spec: domain.DocumentSpec{MetadataFields: []domain.MetadataField{
				{Name: "Employer", ValueSpec: text}, {Name: "Employer", ValueSpec: text},
			}},
			expErr: domain.ErrDuplicateMetadataField,
		},
		{
			name: "Duplicate name differing in surrounding whitespace",
			spec: domain.DocumentSpec{MetadataFields: []domain.MetadataField{
				{Name: "Employer", ValueSpec: text}, {Name: " Employer ", ValueSpec: text},
			}},
			expErr: domain.ErrDuplicateMetadataField,
		},
		{
			name:   "Invalid nested spec",
			spec:   domain.DocumentSpec{MetadataFields: []domain.MetadataField{{Name: "Employer", ValueSpec: domain.ValueSpec{}}}},
			expErr: domain.ErrInvalidValueType,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.spec.Validate(); !errors.Is(err, tc.expErr) {
				t.Errorf("err = %v, want %v", err, tc.expErr)
			}
		})
	}
}
