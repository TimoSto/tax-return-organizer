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

func TestNewItemDefinition(t *testing.T) {
	docSpec := domain.DocumentSpec{MetadataFields: []domain.MetadataField{
		{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText}},
	}}
	dataSpec := domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true}

	for _, tc := range []struct {
		name       string
		title      string
		recurrence domain.Recurrence
		spec       domain.ItemSpec
		expTitle   string
		expKind    domain.Kind
		expErr     error
	}{
		{
			name: "Valid document item", title: "Payslip", recurrence: domain.RecurrenceMonthly,
			spec: docSpec, expTitle: "Payslip", expKind: domain.KindDocument,
		},
		{
			name: "Valid data item", title: "Commuting days", recurrence: domain.RecurrenceAnnually,
			spec: dataSpec, expTitle: "Commuting days", expKind: domain.KindData,
		},
		{
			name: "Document item with empty spec", title: "Tax notice", recurrence: domain.RecurrenceAnnually,
			spec: domain.DocumentSpec{}, expTitle: "Tax notice", expKind: domain.KindDocument,
		},
		{
			name: "Trims title", title: "  Payslip ", recurrence: domain.RecurrenceMonthly,
			spec: docSpec, expTitle: "Payslip", expKind: domain.KindDocument,
		},
		{name: "Blank title", title: " ", recurrence: domain.RecurrenceMonthly, spec: docSpec, expErr: domain.ErrItemTitleRequired},
		{name: "Unknown recurrence", title: "X", recurrence: "weekly", spec: docSpec, expErr: domain.ErrInvalidRecurrence},
		{name: "No spec", title: "X", recurrence: domain.RecurrenceMonthly, expErr: domain.ErrSpecRequired},
		{
			name: "Invalid data spec", title: "X", recurrence: domain.RecurrenceMonthly,
			spec: domain.ValueSpec{}, expErr: domain.ErrInvalidValueType,
		},
		{
			name: "Invalid document spec", title: "X", recurrence: domain.RecurrenceMonthly,
			spec:   domain.DocumentSpec{MetadataFields: []domain.MetadataField{{Name: ""}}},
			expErr: domain.ErrMetadataFieldNameRequired,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item, err := domain.NewItemDefinition(tc.title, tc.recurrence, tc.spec)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}

			if err != nil {
				return
			}

			if item.Title != tc.expTitle || item.Kind() != tc.expKind || item.Recurrence != tc.recurrence {
				t.Errorf("unexpected item %+v", item)
			}

			if item.ID == uuid.Nil {
				t.Error("item has no id")
			}
		})
	}
}
