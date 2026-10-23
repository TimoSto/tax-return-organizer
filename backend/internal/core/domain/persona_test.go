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
			p, err := domain.NewPersona("aaaa", tc.pName, tc.description)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}

			if err != nil {
				return
			}

			if p.Name != tc.expName || p.Description != tc.expDesc {
				t.Errorf("persona = %q/%q, want %q/%q", p.Name, p.Description, tc.expName, tc.expDesc)
			}

			if p.CollectorID != "aaaa" {
				t.Errorf("collector id = %q, want aaaa", p.CollectorID)
			}
		})
	}
}

func TestNewPersona_GeneratesUniqueIDs(t *testing.T) {
	a, _ := domain.NewPersona("aaaa", "A", "")
	b, _ := domain.NewPersona("aaaa", "A", "")

	if a.ID == b.ID {
		t.Error("two personas share an id")
	}
}

func newPersonaWithItems(t *testing.T) (*domain.Persona, *domain.ItemDefinition, *domain.ItemDefinition) {
	t.Helper()

	p, err := domain.NewPersona("aaaa", "Employed", "")
	if err != nil {
		t.Fatal(err)
	}

	payslip, err := p.AddItem("Payslip", domain.KindDocument, domain.RecurrenceMonthly, &domain.DocumentSpec{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	days, err := p.AddItem("Commuting days", domain.KindData, domain.RecurrenceAnnually, nil,
		&domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days"})
	if err != nil {
		t.Fatal(err)
	}

	return p, payslip, days
}

func TestPersona_AddItem(t *testing.T) {
	p, payslip, days := newPersonaWithItems(t)

	if len(p.Items) != 2 || p.Items[0].ID != payslip.ID || p.Items[1].ID != days.ID {
		t.Fatalf("items = %+v, want payslip then days in order", p.Items)
	}

	if payslip.PersonaID != p.ID {
		t.Errorf("item persona id = %v, want %v", payslip.PersonaID, p.ID)
	}

	if _, err := p.AddItem(" ", domain.KindDocument, domain.RecurrenceMonthly, &domain.DocumentSpec{}, nil); !errors.Is(err, domain.ErrItemTitleRequired) {
		t.Errorf("err = %v, want %v", err, domain.ErrItemTitleRequired)
	}

	if len(p.Items) != 2 {
		t.Errorf("rejected item was added, items = %d", len(p.Items))
	}
}

func TestPersona_UpdateItem(t *testing.T) {
	p, payslip, days := newPersonaWithItems(t)
	number := &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "km", Required: true}

	for _, tc := range []struct {
		name   string
		id     uuid.UUID
		title  string
		doc    *domain.DocumentSpec
		val    *domain.ValueSpec
		expErr error
	}{
		{name: "Unknown item", id: uuid.New(), title: "X", doc: &domain.DocumentSpec{}, expErr: domain.ErrItemNotFound},
		{name: "Spec of other kind", id: days.ID, title: "X", doc: &domain.DocumentSpec{}, expErr: domain.ErrSpecMismatch},
		{name: "Blank title", id: payslip.ID, title: "", doc: &domain.DocumentSpec{}, expErr: domain.ErrItemTitleRequired},
		{name: "Valid data update", id: days.ID, title: " Distance ", val: number},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := p.UpdateItem(tc.id, tc.title, domain.RecurrenceQuarterly, tc.doc, tc.val)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}
		})
	}

	got := p.Items[1]
	if got.ID != days.ID || got.Title != "Distance" || got.Kind != domain.KindData ||
		got.Recurrence != domain.RecurrenceQuarterly || got.ValueSpec.Unit != "km" {
		t.Errorf("updated item = %+v", got)
	}

	if p.Items[0].Title != "Payslip" || p.Items[0].Recurrence != domain.RecurrenceMonthly {
		t.Errorf("failed updates changed the first item: %+v", p.Items[0])
	}
}

func TestPersona_RemoveItem(t *testing.T) {
	p, payslip, days := newPersonaWithItems(t)

	if err := p.RemoveItem(uuid.New()); !errors.Is(err, domain.ErrItemNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrItemNotFound)
	}

	if err := p.RemoveItem(payslip.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Items) != 1 || p.Items[0].ID != days.ID {
		t.Errorf("items = %+v, want only days", p.Items)
	}
}
