package personatemplate_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	pt "github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
)

func TestInstantiate_RejectsBrokenItemTemplates(t *testing.T) {
	for name, it := range map[string]pt.ItemTemplate{
		"data item without value": {Title: pt.Localized{pt.LocaleEN: "x"}, Kind: domain.KindData, Recurrence: domain.RecurrenceMonthly},
		"unknown kind":            {Title: pt.Localized{pt.LocaleEN: "x"}, Kind: "other", Recurrence: domain.RecurrenceMonthly},
	} {
		tmpl := pt.PersonaTemplate{Name: pt.Localized{pt.LocaleEN: "p"}, Items: []pt.ItemTemplate{it}}

		if _, err := instantiate(tmpl, pt.LocaleEN); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestInstantiate_UsesTheGivenIDs(t *testing.T) {
	personaID := domain.PersonaID{1}
	itemIDs := []domain.ItemID{{2}, {3}}

	tmpl := pt.PersonaTemplate{
		Name: pt.Localized{pt.LocaleEN: "p"},
		Items: []pt.ItemTemplate{
			{Title: pt.Localized{pt.LocaleEN: "a"}, Kind: domain.KindDocument, Recurrence: domain.RecurrenceMonthly},
			{Title: pt.Localized{pt.LocaleEN: "b"}, Kind: domain.KindDocument, Recurrence: domain.RecurrenceMonthly},
		},
	}

	p, err := tmpl.Instantiate("k7m2", pt.LocaleEN, personaID, itemIDs)
	if err != nil {
		t.Fatal(err)
	}

	if p.ID() != personaID || p.Items()[0].ID() != itemIDs[0] || p.Items()[1].ID() != itemIDs[1] {
		t.Errorf("ids = %v %v %v, want the given ones", p.ID(), p.Items()[0].ID(), p.Items()[1].ID())
	}
}

func TestInstantiate_RejectsWrongNumberOfItemIDs(t *testing.T) {
	tmpl := pt.Predefined()[0]

	if _, err := tmpl.Instantiate("k7m2", pt.LocaleEN, domain.PersonaID{1}, nil); err == nil {
		t.Error("expected an error")
	}
}
