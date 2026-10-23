package domain_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestPersonaTemplates_InstantiateInEveryLocale(t *testing.T) {
	for _, l := range domain.Locales() {
		for _, tmpl := range domain.PersonaTemplates() {
			p, err := tmpl.Instantiate("k7m2", l)
			if err != nil {
				t.Errorf("%s/%s: %v", l, tmpl.Name.For(domain.DefaultLocale), err)

				continue
			}

			if len(p.Items) != len(tmpl.Items) {
				t.Errorf("%s/%s: %d items, want %d", l, p.Name, len(p.Items), len(tmpl.Items))
			}
		}
	}
}

func TestPersonaTemplates_AllTextsTranslated(t *testing.T) {
	check := func(what string, text domain.Localized) {
		for _, l := range domain.Locales() {
			if text[l] == "" {
				t.Errorf("%s has no %s text", what, l)
			}
		}
	}

	for _, tmpl := range domain.PersonaTemplates() {
		check("persona name", tmpl.Name)
		check("description of "+tmpl.Name.For(domain.DefaultLocale), tmpl.Description)

		for _, it := range tmpl.Items {
			check("item title", it.Title)

			for _, f := range it.MetadataFields {
				check("metadata field of "+it.Title.For(domain.DefaultLocale), f.Name)
			}
		}
	}
}

func TestPersonaTemplate_InstantiateUsesLocale(t *testing.T) {
	p, err := domain.PersonaTemplates()[0].Instantiate("k7m2", domain.LocaleDE)
	if err != nil {
		t.Fatal(err)
	}

	if p.Name != "Angestellt" {
		t.Errorf("name = %q, want Angestellt", p.Name)
	}
}

func TestPersonaTemplate_InstantiateCreatesFreshIDs(t *testing.T) {
	tmpl := domain.PersonaTemplates()[0]

	a, _ := tmpl.Instantiate("k7m2", domain.LocaleEN)
	b, _ := tmpl.Instantiate("k7m2", domain.LocaleEN)

	if a.ID == b.ID || a.Items[0].ID == b.Items[0].ID {
		t.Error("two instances share ids")
	}
}

func TestParseLocale(t *testing.T) {
	for in, want := range map[string]domain.Locale{"": domain.LocaleEN, "en": domain.LocaleEN, "de": domain.LocaleDE} {
		if got, err := domain.ParseLocale(in); err != nil || got != want {
			t.Errorf("ParseLocale(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	if _, err := domain.ParseLocale("fr"); err == nil {
		t.Error("expected an error for fr")
	}
}

func TestLocalized_FallsBackToDefault(t *testing.T) {
	if got := (domain.Localized{domain.LocaleEN: "x"}).For(domain.LocaleDE); got != "x" {
		t.Errorf("got %q, want x", got)
	}
}
