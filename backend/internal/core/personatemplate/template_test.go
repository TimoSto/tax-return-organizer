package personatemplate_test

import (
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

func TestInstantiate_RejectsBrokenItemTemplates(t *testing.T) {
	for name, it := range map[string]pt.ItemTemplate{
		"data item without value": {Title: locale.Localized{locale.LocaleEN: "x"}, Kind: domain.KindData, Recurrence: domain.RecurrenceMonthly},
		"unknown kind":            {Title: locale.Localized{locale.LocaleEN: "x"}, Kind: "other", Recurrence: domain.RecurrenceMonthly},
	} {
		tmpl := pt.PersonaTemplate{Name: locale.Localized{locale.LocaleEN: "p"}, Items: []pt.ItemTemplate{it}}

		if _, err := instantiate(tmpl, locale.LocaleEN); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestInstantiate_UsesTheGivenIDs(t *testing.T) {
	personaID := domain.PersonaID{1}
	itemIDs := []domain.ItemID{{2}, {3}}

	tmpl := pt.PersonaTemplate{
		Name: locale.Localized{locale.LocaleEN: "p"},
		Items: []pt.ItemTemplate{
			{Title: locale.Localized{locale.LocaleEN: "a"}, Kind: domain.KindDocument, Recurrence: domain.RecurrenceMonthly},
			{Title: locale.Localized{locale.LocaleEN: "b"}, Kind: domain.KindDocument, Recurrence: domain.RecurrenceMonthly},
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
