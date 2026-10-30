package domain_test

import (
	"errors"
	"fmt"
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
			item, err := domain.NewItemDefinition(domain.ItemID(uuid.New()), tc.title, tc.recurrence, tc.spec, nil)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}

			if err != nil {
				return
			}

			if item.Title() != tc.expTitle || item.Kind() != tc.expKind || item.Recurrence() != tc.recurrence {
				t.Errorf("unexpected item %+v", item)
			}

			if item.ID() == (domain.ItemID{}) {
				t.Error("item has no id")
			}
		})
	}
}

func TestNewItemDefinition_InstanceFields(t *testing.T) {
	iban := domain.MetadataField{Name: "IBAN", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}}
	bank := domain.MetadataField{Name: "Bank", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText}}

	for _, tc := range []struct {
		name          string
		fields        []domain.MetadataField
		expRepeatable bool
		expErr        error
	}{
		{name: "No instance fields", fields: nil},
		{name: "Empty instance fields", fields: []domain.MetadataField{}},
		{name: "Required label", fields: []domain.MetadataField{iban}, expRepeatable: true},
		{name: "Optional field after the label", fields: []domain.MetadataField{iban, bank}, expRepeatable: true},
		{name: "Optional label", fields: []domain.MetadataField{bank, iban}, expErr: domain.ErrInstanceLabelNotRequired},
		{name: "Blank name", fields: []domain.MetadataField{{Name: " ", ValueSpec: iban.ValueSpec}}, expErr: domain.ErrMetadataFieldNameRequired},
		{name: "Duplicate name", fields: []domain.MetadataField{iban, iban}, expErr: domain.ErrDuplicateMetadataField},
		{name: "Invalid type", fields: []domain.MetadataField{{Name: "IBAN", ValueSpec: domain.ValueSpec{Required: true}}}, expErr: domain.ErrInvalidValueType},
		{name: "Too many", fields: tooManyFields(iban), expErr: domain.ErrTooManyMetadataFields},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item, err := domain.NewItemDefinition(domain.ItemID(uuid.New()), "Bank statement", domain.RecurrenceMonthly,
				domain.DocumentSpec{}, tc.fields)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}

			if err != nil {
				return
			}

			if item.Repeatable() != tc.expRepeatable {
				t.Errorf("Repeatable() = %v, want %v", item.Repeatable(), tc.expRepeatable)
			}

			if got := item.InstanceFields(); !tc.expRepeatable && got != nil {
				t.Errorf("InstanceFields() = %v, want nil", got)
			}
		})
	}
}

func TestItemDefinition_InstanceFieldsAreCopied(t *testing.T) {
	fields := []domain.MetadataField{{Name: "IBAN", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}}}

	item, err := domain.NewItemDefinition(domain.ItemID(uuid.New()), "Bank statement", domain.RecurrenceMonthly,
		domain.DocumentSpec{}, fields)
	if err != nil {
		t.Fatal(err)
	}

	fields[0].Name = "Changed"
	item.InstanceFields()[0].Name = "Changed too"

	if got := item.InstanceFields()[0].Name; got != "IBAN" {
		t.Errorf("instance field name = %q, want IBAN: the item shares a slice", got)
	}
}

func tooManyFields(label domain.MetadataField) []domain.MetadataField {
	fields := []domain.MetadataField{label}
	for i := range domain.MaxMetadataFields {
		fields = append(fields, domain.MetadataField{Name: fmt.Sprintf("Field %d", i), ValueSpec: label.ValueSpec})
	}

	return fields
}
