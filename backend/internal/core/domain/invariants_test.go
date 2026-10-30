package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func runes(n int) string { return strings.Repeat("ä", n) }

func newItemID() domain.ItemID { return domain.ItemID(uuid.New()) }

func TestParseEmail_LowercasesAndBoundsLength(t *testing.T) {
	got, err := domain.ParseEmail("John.Doe@Example.COM")
	if err != nil || got != "john.doe@example.com" {
		t.Errorf("ParseEmail() = %q, %v, want the lowercased address", got, err)
	}

	local := strings.Repeat("a", 64)
	atLimit := local + "@" + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 57) + ".com"

	if len(atLimit) != domain.MaxEmailLength {
		t.Fatalf("fixture has %d chars, want %d", len(atLimit), domain.MaxEmailLength)
	}

	if _, err := domain.ParseEmail(atLimit); err != nil {
		t.Errorf("address at the limit: %v", err)
	}

	if _, err := domain.ParseEmail("a" + atLimit); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("address over the limit: error = %v, want ErrInvalidEmail", err)
	}
}

func TestNewCollector_NormalizesEmailAndBoundsName(t *testing.T) {
	c, err := domain.NewCollector(testCollectorID, "Ann", " Ann@Example.com ")
	if err != nil || c.Email() != "ann@example.com" {
		t.Errorf("NewCollector() = %v, %v, want a lowercased email", c, err)
	}

	if _, err := domain.NewCollector(testCollectorID, runes(domain.MaxCollectorNameLength), "a@example.com"); err != nil {
		t.Errorf("name at the limit: %v", err)
	}

	if _, err := domain.NewCollector(testCollectorID, runes(domain.MaxCollectorNameLength+1), "a@example.com"); !errors.Is(err, domain.ErrNameTooLong) {
		t.Errorf("name over the limit: error = %v, want ErrNameTooLong", err)
	}
}

func TestRehydrate_RejectsInvalidCollectorID(t *testing.T) {
	var zero domain.CollectorID

	if _, err := domain.RehydrateCollector(zero, "Ann", "a@example.com"); !errors.Is(err, domain.ErrInvalidCollectorID) {
		t.Errorf("RehydrateCollector(zero id) error = %v, want ErrInvalidCollectorID", err)
	}

	if _, err := domain.RehydratePersona(domain.PersonaID(uuid.New()), zero, "Employed", "", nil); !errors.Is(err, domain.ErrInvalidCollectorID) {
		t.Errorf("RehydratePersona(zero id) error = %v, want ErrInvalidCollectorID", err)
	}
}

func TestNewPersona_BoundsTexts(t *testing.T) {
	id := domain.PersonaID(uuid.New())

	for _, tc := range []struct {
		name, personaName, description string
		expErr                         error
	}{
		{"at the limits", runes(domain.MaxPersonaNameLength), runes(domain.MaxPersonaDescriptionLength), nil},
		{"name too long", runes(domain.MaxPersonaNameLength + 1), "", domain.ErrPersonaNameTooLong},
		{"description too long", "Employed", runes(domain.MaxPersonaDescriptionLength + 1), domain.ErrPersonaDescriptionTooLong},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := domain.NewPersona(id, testCollectorID, tc.personaName, tc.description); !errors.Is(err, tc.expErr) {
				t.Errorf("error = %v, want %v", err, tc.expErr)
			}
		})
	}
}

func TestPersona_AddItem_RejectsDuplicateID(t *testing.T) {
	p, _ := domain.NewPersona(domain.PersonaID(uuid.New()), testCollectorID, "Employed", "")
	id := newItemID()

	if _, err := p.AddItem(id, "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{}, nil); err != nil {
		t.Fatal(err)
	}

	_, err := p.AddItem(id, "Other", domain.RecurrenceMonthly, domain.DocumentSpec{}, nil)
	if !errors.Is(err, domain.ErrDuplicateItemID) {
		t.Errorf("error = %v, want ErrDuplicateItemID", err)
	}

	// A server-side fault, not a client error: it must not be a classified domain error.
	var de *domain.Error
	if errors.As(err, &de) {
		t.Errorf("ErrDuplicateItemID must not be a *domain.Error, got kind %v", de.Kind)
	}

	if len(p.Items()) != 1 {
		t.Errorf("items = %d, want 1", len(p.Items()))
	}
}

func TestPersona_AddItem_BoundsItemCount(t *testing.T) {
	p, _ := domain.NewPersona(domain.PersonaID(uuid.New()), testCollectorID, "Employed", "")

	for range domain.MaxItemsPerPersona {
		if _, err := p.AddItem(newItemID(), "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{}, nil); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := p.AddItem(newItemID(), "One too many", domain.RecurrenceMonthly, domain.DocumentSpec{}, nil); !errors.Is(err, domain.ErrTooManyItems) {
		t.Errorf("error = %v, want ErrTooManyItems", err)
	}
}

func TestRehydratePersona_RejectsDuplicateIDsAndTooManyItems(t *testing.T) {
	spec := domain.DocumentSpec{}
	id := newItemID()
	item, _ := domain.NewItemDefinition(id, "Payslip", domain.RecurrenceMonthly, spec, nil)
	pid := domain.PersonaID(uuid.New())

	if _, err := domain.RehydratePersona(pid, testCollectorID, "Employed", "", []domain.ItemDefinition{item, item}); !errors.Is(err, domain.ErrDuplicateItemID) {
		t.Errorf("duplicate: error = %v, want ErrDuplicateItemID", err)
	}

	items := make([]domain.ItemDefinition, domain.MaxItemsPerPersona+1)
	for i := range items {
		items[i], _ = domain.NewItemDefinition(newItemID(), "Payslip", domain.RecurrenceMonthly, spec, nil)
	}

	if _, err := domain.RehydratePersona(pid, testCollectorID, "Employed", "", items); !errors.Is(err, domain.ErrTooManyItems) {
		t.Errorf("too many: error = %v, want ErrTooManyItems", err)
	}

	if _, err := domain.RehydratePersona(pid, testCollectorID, "Employed", "", items[:domain.MaxItemsPerPersona]); err != nil {
		t.Errorf("at the limit: %v", err)
	}
}

func TestNewItemDefinition_BoundsTitle(t *testing.T) {
	spec := domain.DocumentSpec{}

	if _, err := domain.NewItemDefinition(newItemID(), runes(domain.MaxItemTitleLength), domain.RecurrenceMonthly, spec, nil); err != nil {
		t.Errorf("title at the limit: %v", err)
	}

	if _, err := domain.NewItemDefinition(newItemID(), runes(domain.MaxItemTitleLength+1), domain.RecurrenceMonthly, spec, nil); !errors.Is(err, domain.ErrItemTitleTooLong) {
		t.Errorf("title over the limit: error = %v, want ErrItemTitleTooLong", err)
	}
}

func TestSpecs_BoundSizes(t *testing.T) {
	text := domain.ValueSpec{Type: domain.ValueTypeText}

	if err := (domain.ValueSpec{Type: domain.ValueTypeText, Unit: runes(domain.MaxUnitLength)}).Validate(); err != nil {
		t.Errorf("unit at the limit: %v", err)
	}

	if err := (domain.ValueSpec{Type: domain.ValueTypeText, Unit: runes(domain.MaxUnitLength + 1)}).Validate(); !errors.Is(err, domain.ErrUnitTooLong) {
		t.Errorf("unit over the limit: error = %v, want ErrUnitTooLong", err)
	}

	long := domain.DocumentSpec{MetadataFields: []domain.MetadataField{{Name: runes(domain.MaxMetadataFieldNameLength + 1), ValueSpec: text}}}
	if err := long.Validate(); !errors.Is(err, domain.ErrMetadataFieldNameTooLong) {
		t.Errorf("field name over the limit: error = %v, want ErrMetadataFieldNameTooLong", err)
	}

	fields := make([]domain.MetadataField, domain.MaxMetadataFields+1)
	for i := range fields {
		fields[i] = domain.MetadataField{Name: strings.Repeat("f", i+1), ValueSpec: text}
	}

	if err := (domain.DocumentSpec{MetadataFields: fields[:domain.MaxMetadataFields]}).Validate(); err != nil {
		t.Errorf("fields at the limit: %v", err)
	}

	if err := (domain.DocumentSpec{MetadataFields: fields}).Validate(); !errors.Is(err, domain.ErrTooManyMetadataFields) {
		t.Errorf("fields over the limit: error = %v, want ErrTooManyMetadataFields", err)
	}
}
