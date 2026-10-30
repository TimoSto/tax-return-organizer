package personatemplate_test

import (
	"slices"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/locale"
	pt "github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

var (
	// testCollectorID is the collector the tests act for.
	testCollectorID = domaintest.CollectorID()
)

func TestInstantiate_RejectsItemWithoutSpec(t *testing.T) {
	tmpl := pt.PersonaTemplate{
		Name:  locale.Localized{locale.LocaleEN: "p"},
		Items: []pt.ItemTemplate{{Title: locale.Localized{locale.LocaleEN: "x"}, Recurrence: domain.RecurrenceMonthly}},
	}

	if _, err := instantiate(tmpl, locale.LocaleEN); err == nil {
		t.Error("expected an error")
	}
}

func TestInstantiate_KindFollowsSpec(t *testing.T) {
	amount := domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR"}
	tmpl := pt.PersonaTemplate{
		Name: locale.Localized{locale.LocaleEN: "p"},
		Items: []pt.ItemTemplate{
			{
				Title:      locale.Localized{locale.LocaleEN: "doc"},
				Recurrence: domain.RecurrenceMonthly,
				Spec: pt.DocumentSpecTemplate{MetadataFields: []pt.MetadataFieldTemplate{
					{Name: locale.Localized{locale.LocaleEN: "Amount", locale.LocaleDE: "Betrag"}, ValueSpec: amount},
				}},
			},
			{
				Title:      locale.Localized{locale.LocaleEN: "data"},
				Recurrence: domain.RecurrenceMonthly,
				Spec:       pt.ValueSpecTemplate(amount),
			},
		},
	}

	p, err := instantiate(tmpl, locale.LocaleDE)
	if err != nil {
		t.Fatal(err)
	}

	doc, ok := p.Items()[0].Spec().(domain.DocumentSpec)
	if !ok {
		t.Fatalf("spec of the document item = %T, want domain.DocumentSpec", p.Items()[0].Spec())
	}

	if want := []domain.MetadataField{{Name: "Betrag", ValueSpec: amount}}; !slices.Equal(doc.MetadataFields, want) {
		t.Errorf("metadata fields = %v, want %v", doc.MetadataFields, want)
	}

	if got := p.Items()[1].Spec(); got != domain.ItemSpec(amount) {
		t.Errorf("spec of the data item = %#v, want %#v", got, amount)
	}
}

func TestInstantiate_LocalizesInstanceFields(t *testing.T) {
	label := domain.ValueSpec{Type: domain.ValueTypeText, Required: true}
	tmpl := pt.PersonaTemplate{
		Name: locale.Localized{locale.LocaleEN: "p"},
		Items: []pt.ItemTemplate{
			{
				Title:      locale.Localized{locale.LocaleEN: "Bank statement"},
				Recurrence: domain.RecurrenceMonthly,
				Spec:       pt.DocumentSpecTemplate{},
				InstanceFields: []pt.MetadataFieldTemplate{
					{Name: locale.Localized{locale.LocaleEN: "Account number", locale.LocaleDE: "Kontonummer"}, ValueSpec: label},
				},
			},
			{Title: locale.Localized{locale.LocaleEN: "Single"}, Recurrence: domain.RecurrenceMonthly, Spec: pt.DocumentSpecTemplate{}},
		},
	}

	p, err := instantiate(tmpl, locale.LocaleDE)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := p.Items()[0].InstanceFields(), []domain.MetadataField{{Name: "Kontonummer", ValueSpec: label}}; !slices.Equal(got, want) {
		t.Errorf("instance fields = %v, want %v", got, want)
	}

	if p.Items()[1].Repeatable() {
		t.Error("an item template without instance fields became repeatable")
	}
}

func TestInstantiate_UsesTheGivenIDs(t *testing.T) {
	personaID := domain.PersonaID{1}
	itemIDs := []domain.ItemID{{2}, {3}}

	tmpl := pt.PersonaTemplate{
		Name: locale.Localized{locale.LocaleEN: "p"},
		Items: []pt.ItemTemplate{
			{Title: locale.Localized{locale.LocaleEN: "a"}, Recurrence: domain.RecurrenceMonthly, Spec: pt.DocumentSpecTemplate{}},
			{Title: locale.Localized{locale.LocaleEN: "b"}, Recurrence: domain.RecurrenceMonthly, Spec: pt.DocumentSpecTemplate{}},
		},
	}

	p, err := tmpl.Instantiate(testCollectorID, locale.LocaleEN, personaID, itemIDs)
	if err != nil {
		t.Fatal(err)
	}

	if p.ID() != personaID || p.Items()[0].ID() != itemIDs[0] || p.Items()[1].ID() != itemIDs[1] {
		t.Errorf("ids = %v %v %v, want the given ones", p.ID(), p.Items()[0].ID(), p.Items()[1].ID())
	}
}

func TestInstantiate_RejectsWrongNumberOfItemIDs(t *testing.T) {
	tmpl := pt.Predefined()[0]

	if _, err := tmpl.Instantiate(testCollectorID, locale.LocaleEN, domain.PersonaID{1}, nil); err == nil {
		t.Error("expected an error")
	}
}
