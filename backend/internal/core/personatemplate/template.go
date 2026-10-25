// Package personatemplate holds predefined personas and the logic to copy them to a collector.
package personatemplate

import (
	"errors"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// IDs creates the ids of the persona and items an Instantiate call makes.
type IDs interface {
	NewPersonaID() (domain.PersonaID, error)
	NewItemID() (domain.ItemID, error)
}

// PersonaTemplate is a predefined persona that is copied to a collector, in the collector's locale.
// The copy is a normal persona that the collector owns and may change, nothing links it back to the template.
type PersonaTemplate struct {
	Name        Localized
	Description Localized
	Items       []ItemTemplate
}

// ItemTemplate is a predefined item definition. A document item has MetadataFields, a data item has Value.
type ItemTemplate struct {
	Title          Localized
	Kind           domain.Kind
	Recurrence     domain.Recurrence
	MetadataFields []MetadataFieldTemplate
	Value          *domain.ValueSpec
}

// MetadataFieldTemplate is a predefined metadata field of a document item.
type MetadataFieldTemplate struct {
	Name      Localized
	ValueSpec domain.ValueSpec
}

// Instantiate creates a persona of the collector from the template, with all texts in locale l and ids from ids.
func (t PersonaTemplate) Instantiate(ids IDs, collectorID domain.CollectorID, l Locale) (*domain.Persona, error) {
	personaID, err := ids.NewPersonaID()
	if err != nil {
		return nil, err
	}

	p, err := domain.NewPersona(personaID, collectorID, t.Name.For(l), t.Description.For(l))
	if err != nil {
		return nil, err
	}

	for _, it := range t.Items {
		spec, err := it.spec(l)
		if err != nil {
			return nil, fmt.Errorf("item %q: %w", it.Title.For(DefaultLocale), err)
		}

		itemID, err := ids.NewItemID()
		if err != nil {
			return nil, err
		}

		if _, err := p.AddItem(itemID, it.Title.For(l), it.Recurrence, spec); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (it ItemTemplate) spec(l Locale) (domain.ItemSpec, error) {
	switch it.Kind {
	case domain.KindDocument:
		doc := domain.DocumentSpec{}
		for _, f := range it.MetadataFields {
			doc.MetadataFields = append(doc.MetadataFields, domain.MetadataField{Name: f.Name.For(l), ValueSpec: f.ValueSpec})
		}

		return doc, nil
	case domain.KindData:
		if it.Value == nil {
			return nil, errors.New("data item template has no value")
		}

		return *it.Value, nil
	default:
		return nil, fmt.Errorf("unknown item kind %q", it.Kind)
	}
}
