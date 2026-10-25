package personatemplate_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/idgen"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	pt "github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
)

func TestInstantiate_RejectsBrokenItemTemplates(t *testing.T) {
	for name, it := range map[string]pt.ItemTemplate{
		"data item without value": {Title: pt.Localized{pt.LocaleEN: "x"}, Kind: domain.KindData, Recurrence: domain.RecurrenceMonthly},
		"unknown kind":            {Title: pt.Localized{pt.LocaleEN: "x"}, Kind: "other", Recurrence: domain.RecurrenceMonthly},
	} {
		tmpl := pt.PersonaTemplate{Name: pt.Localized{pt.LocaleEN: "p"}, Items: []pt.ItemTemplate{it}}

		if _, err := tmpl.Instantiate(idgen.Random{}, "k7m2", pt.LocaleEN); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
