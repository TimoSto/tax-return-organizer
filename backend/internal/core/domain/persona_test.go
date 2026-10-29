package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewPersona(t *testing.T) {
	for _, tc := range []struct {
		name        string
		pName       string
		description string
		expName     string
		expDesc     string
		expErr      error
	}{
		{name: "Valid", pName: "Employed", description: "Has a job", expName: "Employed", expDesc: "Has a job"},
		{name: "Trims input", pName: "  Letting ", description: " rents out ", expName: "Letting", expDesc: "rents out"},
		{name: "Empty description is allowed", pName: "ETF investing", expName: "ETF investing"},
		{name: "Empty name", pName: "", expErr: domain.ErrPersonaNameRequired},
		{name: "Blank name", pName: "  \t", expErr: domain.ErrPersonaNameRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := domain.NewPersona(domain.PersonaID(uuid.New()), testCollectorID, tc.pName, tc.description)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}

			if err != nil {
				return
			}

			if p.Name() != tc.expName || p.Description() != tc.expDesc {
				t.Errorf("persona = %q/%q, want %q/%q", p.Name(), p.Description(), tc.expName, tc.expDesc)
			}

			if p.CollectorID() != testCollectorID {
				t.Errorf("collector id = %q, want %q", p.CollectorID(), testCollectorID)
			}
		})
	}
}

func TestNewPersona_GeneratesUniqueIDs(t *testing.T) {
	a, _ := domain.NewPersona(domain.PersonaID(uuid.New()), testCollectorID, "A", "")
	b, _ := domain.NewPersona(domain.PersonaID(uuid.New()), testCollectorID, "A", "")

	if a.ID() == b.ID() {
		t.Error("two personas share an id")
	}
}

func newPersonaWithItems(t *testing.T) (*domain.Persona, domain.ItemDefinition, domain.ItemDefinition) {
	t.Helper()

	p, err := domain.NewPersona(domain.PersonaID(uuid.New()), testCollectorID, "Employed", "")
	if err != nil {
		t.Fatal(err)
	}

	payslip, err := p.AddItem(domain.ItemID(uuid.New()), "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{})
	if err != nil {
		t.Fatal(err)
	}

	days, err := p.AddItem(domain.ItemID(uuid.New()), "Commuting days", domain.RecurrenceAnnually,
		domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days"})
	if err != nil {
		t.Fatal(err)
	}

	return p, payslip, days
}

func TestPersona_AddItem(t *testing.T) {
	p, payslip, days := newPersonaWithItems(t)

	if items := p.Items(); len(items) != 2 || items[0].ID() != payslip.ID() || items[1].ID() != days.ID() {
		t.Fatalf("items = %+v, want payslip then days in order", items)
	}

	if _, err := p.AddItem(domain.ItemID(uuid.New()), " ", domain.RecurrenceMonthly, domain.DocumentSpec{}); !errors.Is(err, domain.ErrItemTitleRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrItemTitleRequired)
	}

	if len(p.Items()) != 2 {
		t.Errorf("rejected item was added, items = %d", len(p.Items()))
	}

	// Items hands out copies: changing them does not change the persona.
	items := p.Items()
	items[0] = days

	if p.Items()[0].ID() != payslip.ID() {
		t.Errorf("changing the returned slice changed the persona: %+v", p.Items()[0])
	}
}

func TestPersona_UpdateItem(t *testing.T) {
	p, payslip, days := newPersonaWithItems(t)
	number := domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "km", Required: true}

	for _, tc := range []struct {
		name   string
		id     domain.ItemID
		title  string
		spec   domain.ItemSpec
		expErr error
	}{
		{name: "Unknown item", id: domain.ItemID(uuid.New()), title: "X", spec: domain.DocumentSpec{}, expErr: domain.ErrItemNotFound},
		{name: "Spec of other kind", id: days.ID(), title: "X", spec: domain.DocumentSpec{}, expErr: domain.ErrSpecMismatch},
		{name: "Blank title", id: payslip.ID(), title: "", spec: domain.DocumentSpec{}, expErr: domain.ErrItemTitleRequired},
		{name: "No spec", id: payslip.ID(), title: "X", expErr: domain.ErrSpecRequired},
		{name: "Valid data update", id: days.ID(), title: " Distance ", spec: number},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := p.UpdateItem(tc.id, tc.title, domain.RecurrenceQuarterly, tc.spec)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}
		})
	}

	got := p.Items()[1]
	if got.ID() != days.ID() || got.Title() != "Distance" || got.Kind() != domain.KindData ||
		got.Recurrence() != domain.RecurrenceQuarterly || got.Spec() != number {
		t.Errorf("updated item = %+v", got)
	}

	if first := p.Items()[0]; first.Title() != "Payslip" || first.Recurrence() != domain.RecurrenceMonthly {
		t.Errorf("failed updates changed the first item: %+v", p.Items()[0])
	}
}

func TestPersona_RemoveItem(t *testing.T) {
	p, payslip, days := newPersonaWithItems(t)

	if err := p.RemoveItem(domain.ItemID(uuid.New())); !errors.Is(err, domain.ErrItemNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrItemNotFound)
	}

	if err := p.RemoveItem(payslip.ID()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if items := p.Items(); len(items) != 1 || items[0].ID() != days.ID() {
		t.Errorf("items = %+v, want only days", items)
	}
}

func TestPersona_AddItemCopiesTheSpec(t *testing.T) {
	p, _, _ := newPersonaWithItems(t)
	fields := []domain.MetadataField{{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText}}}

	item, err := p.AddItem(domain.ItemID(uuid.New()), "Notice", domain.RecurrenceAnnually, domain.DocumentSpec{MetadataFields: fields})
	if err != nil {
		t.Fatal(err)
	}

	fields[0].Name = "Changed"
	item.Spec().(domain.DocumentSpec).MetadataFields[0].Name = "Changed too"

	got := p.Items()[2].Spec().(domain.DocumentSpec)
	if got.MetadataFields[0].Name != "Employer" {
		t.Errorf("the persona's spec was changed through a shared slice: %+v", got)
	}
}

func TestRehydratePersona(t *testing.T) {
	id, collectorID := domain.PersonaID(uuid.New()), testCollectorID
	item, err := domain.NewItemDefinition(domain.ItemID(uuid.New()), "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{})
	if err != nil {
		t.Fatal(err)
	}

	items := []domain.ItemDefinition{item}

	p, err := domain.RehydratePersona(id, collectorID, "Employed", "", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.ID() != id || p.CollectorID() != collectorID || len(p.Items()) != 1 {
		t.Errorf("unexpected persona %+v", p)
	}

	items[0] = domain.ItemDefinition{}

	if p.Items()[0].ID() != item.ID() {
		t.Error("the persona shares the caller's slice")
	}

	if _, err := domain.RehydratePersona(id, collectorID, " ", "", nil); !errors.Is(err, domain.ErrPersonaNameRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrPersonaNameRequired)
	}
}
