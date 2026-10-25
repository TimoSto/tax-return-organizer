package personatemplate_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/idgen"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	pt "github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
)

func TestPredefined_InstantiateInEveryLocale(t *testing.T) {
	for _, l := range pt.Locales() {
		for _, tmpl := range pt.Predefined() {
			p, err := instantiate(tmpl, l)
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

func TestPredefined_AllTextsTranslated(t *testing.T) {
	check := func(what string, text pt.Localized) {
		for _, l := range pt.Locales() {
			if text[l] == "" {
				t.Errorf("%s has no %s text", what, l)
			}
		}
	}

	for _, tmpl := range pt.Predefined() {
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

func TestPredefined_InstantiateUsesLocale(t *testing.T) {
	p, err := instantiate(pt.Predefined()[0], pt.LocaleDE)
	if err != nil {
		t.Fatal(err)
	}

	if p.Name() != "Angestellt" {
		t.Errorf("name = %q, want Angestellt", p.Name())
	}
}

// instantiate copies tmpl to a collector with random ids.
func instantiate(tmpl pt.PersonaTemplate, l pt.Locale) (*domain.Persona, error) {
	ids := idgen.Random{}

	personaID, err := ids.NewPersonaID()
	if err != nil {
		return nil, err
	}

	itemIDs := make([]domain.ItemID, len(tmpl.Items))
	for i := range itemIDs {
		if itemIDs[i], err = ids.NewItemID(); err != nil {
			return nil, err
		}
	}

	return tmpl.Instantiate("k7m2", l, personaID, itemIDs)
}
