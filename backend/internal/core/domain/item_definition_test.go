package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestKind_Valid(t *testing.T) {
	for _, tc := range []struct {
		in   domain.Kind
		want bool
	}{
		{domain.KindDocument, true},
		{domain.KindData, true},
		{"", false},
		{"file", false},
	} {
		if got := tc.in.Valid(); got != tc.want {
			t.Errorf("Kind(%q).Valid() = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestRecurrence(t *testing.T) {
	for _, tc := range []struct {
		in      domain.Recurrence
		periods int
		valid   bool
	}{
		{domain.RecurrenceMonthly, 12, true},
		{domain.RecurrenceQuarterly, 4, true},
		{domain.RecurrenceSemiannually, 2, true},
		{domain.RecurrenceAnnually, 1, true},
		{"", 0, false},
		{"weekly", 0, false},
	} {
		if got := tc.in.Periods(); got != tc.periods {
			t.Errorf("Recurrence(%q).Periods() = %d, want %d", tc.in, got, tc.periods)
		}

		if got := tc.in.Valid(); got != tc.valid {
			t.Errorf("Recurrence(%q).Valid() = %v, want %v", tc.in, got, tc.valid)
		}
	}
}

func TestNewItemDefinition_NormalizesEmptyMetadataFields(t *testing.T) {
	in := &domain.DocumentSpec{MetadataFields: []domain.MetadataField{}}

	item, err := domain.NewItemDefinition(uuid.New(), "Notice", domain.KindDocument, domain.RecurrenceAnnually, in, nil)
	if err != nil {
		t.Fatal(err)
	}

	if item.DocumentSpec.MetadataFields != nil {
		t.Errorf("MetadataFields = %#v, want nil", item.DocumentSpec.MetadataFields)
	}

	if in.MetadataFields == nil {
		t.Error("the caller's spec was modified")
	}
}

func TestNewItemDefinition(t *testing.T) {
	personaID := uuid.New()
	docSpec := &domain.DocumentSpec{MetadataFields: []domain.MetadataField{
		{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText}},
	}}
	dataSpec := &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true}

	for _, tc := range []struct {
		name       string
		title      string
		kind       domain.Kind
		recurrence domain.Recurrence
		doc        *domain.DocumentSpec
		data       *domain.ValueSpec
		expTitle   string
		expErr     error
	}{
		{
			name: "Valid document item", title: "Payslip", kind: domain.KindDocument,
			recurrence: domain.RecurrenceMonthly, doc: docSpec, expTitle: "Payslip",
		},
		{
			name: "Valid data item", title: "Commuting days", kind: domain.KindData,
			recurrence: domain.RecurrenceAnnually, data: dataSpec, expTitle: "Commuting days",
		},
		{
			name: "Document item with empty spec", title: "Tax notice", kind: domain.KindDocument,
			recurrence: domain.RecurrenceAnnually, doc: &domain.DocumentSpec{}, expTitle: "Tax notice",
		},
		{
			name: "Trims title", title: "  Payslip ", kind: domain.KindDocument,
			recurrence: domain.RecurrenceMonthly, doc: docSpec, expTitle: "Payslip",
		},
		{
			name: "Blank title", title: " ", kind: domain.KindDocument,
			recurrence: domain.RecurrenceMonthly, doc: docSpec, expErr: domain.ErrItemTitleRequired,
		},
		{
			name: "Unknown kind", title: "X", kind: "file",
			recurrence: domain.RecurrenceMonthly, doc: docSpec, expErr: domain.ErrInvalidKind,
		},
		{
			name: "Unknown recurrence", title: "X", kind: domain.KindDocument,
			recurrence: "weekly", doc: docSpec, expErr: domain.ErrInvalidRecurrence,
		},
		{
			name: "Document item without spec", title: "X", kind: domain.KindDocument,
			recurrence: domain.RecurrenceMonthly, expErr: domain.ErrSpecMismatch,
		},
		{
			name: "Document item with data spec", title: "X", kind: domain.KindDocument,
			recurrence: domain.RecurrenceMonthly, data: dataSpec, expErr: domain.ErrSpecMismatch,
		},
		{
			name: "Document item with both specs", title: "X", kind: domain.KindDocument,
			recurrence: domain.RecurrenceMonthly, doc: docSpec, data: dataSpec, expErr: domain.ErrSpecMismatch,
		},
		{
			name: "Data item without spec", title: "X", kind: domain.KindData,
			recurrence: domain.RecurrenceMonthly, expErr: domain.ErrSpecMismatch,
		},
		{
			name: "Data item with document spec", title: "X", kind: domain.KindData,
			recurrence: domain.RecurrenceMonthly, doc: docSpec, expErr: domain.ErrSpecMismatch,
		},
		{
			name: "Invalid data spec", title: "X", kind: domain.KindData,
			recurrence: domain.RecurrenceMonthly, data: &domain.ValueSpec{}, expErr: domain.ErrInvalidValueType,
		},
		{
			name: "Invalid document spec", title: "X", kind: domain.KindDocument,
			recurrence: domain.RecurrenceMonthly,
			doc:        &domain.DocumentSpec{MetadataFields: []domain.MetadataField{{Name: ""}}},
			expErr:     domain.ErrMetadataFieldNameRequired,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item, err := domain.NewItemDefinition(personaID, tc.title, tc.kind, tc.recurrence, tc.doc, tc.data)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}

			if err != nil {
				return
			}

			if item.Title != tc.expTitle || item.PersonaID != personaID || item.Kind != tc.kind || item.Recurrence != tc.recurrence {
				t.Errorf("unexpected item %+v", item)
			}

			if item.ID == uuid.Nil {
				t.Error("item has no id")
			}
		})
	}
}
