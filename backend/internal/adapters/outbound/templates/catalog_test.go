package templates_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/templates"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	pt "github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
)

func TestCatalog_InstantiateInEveryLocale(t *testing.T) {
	for _, l := range pt.Locales() {
		for _, tmpl := range templates.New().PersonaTemplates() {
			p, err := tmpl.Instantiate(domain.RandomIDs{}, "k7m2", l)
			if err != nil {
				t.Errorf("%s/%s: %v", l, tmpl.Name.For(pt.DefaultLocale), err)

				continue
			}

			if len(p.Items()) != len(tmpl.Items) {
				t.Errorf("%s/%s: %d items, want %d", l, p.Name(), len(p.Items()), len(tmpl.Items))
			}
		}
	}
}

func TestCatalog_AllTextsTranslated(t *testing.T) {
	check := func(what string, text pt.Localized) {
		for _, l := range pt.Locales() {
			if text[l] == "" {
				t.Errorf("%s has no %s text", what, l)
			}
		}
	}

	for _, tmpl := range templates.New().PersonaTemplates() {
		check("persona name", tmpl.Name)
		check("description of "+tmpl.Name.For(pt.DefaultLocale), tmpl.Description)

		for _, it := range tmpl.Items {
			check("item title", it.Title)

			for _, f := range it.MetadataFields {
				check("metadata field of "+it.Title.For(pt.DefaultLocale), f.Name)
			}
		}
	}
}

func TestCatalog_InstantiateUsesLocale(t *testing.T) {
	p, err := templates.New().PersonaTemplates()[0].Instantiate(domain.RandomIDs{}, "k7m2", pt.LocaleDE)
	if err != nil {
		t.Fatal(err)
	}

	if p.Name() != "Angestellt" {
		t.Errorf("name = %q, want Angestellt", p.Name())
	}
}

func TestCatalog_InstantiateCreatesFreshIDs(t *testing.T) {
	tmpl := templates.New().PersonaTemplates()[0]

	a, _ := tmpl.Instantiate(domain.RandomIDs{}, "k7m2", pt.LocaleEN)
	b, _ := tmpl.Instantiate(domain.RandomIDs{}, "k7m2", pt.LocaleEN)

	if a.ID() == b.ID() || a.Items()[0].ID() == b.Items()[0].ID() {
		t.Error("two instances share ids")
	}
}
