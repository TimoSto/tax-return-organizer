// Package personatemplate holds predefined personas and the logic to copy them to a collector.
package personatemplate

import (
	"errors"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

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

// Instantiate creates a persona of the collector from the template, with all texts in locale l.
// It is pure: the caller supplies the persona id and one item id per item of the template, in order.
func (t PersonaTemplate) Instantiate(collectorID domain.CollectorID, l Locale, personaID domain.PersonaID, itemIDs []domain.ItemID) (*domain.Persona, error) {
	if len(itemIDs) != len(t.Items) {
		return nil, fmt.Errorf("persona template %q has %d items, got %d item ids", t.Name.For(DefaultLocale), len(t.Items), len(itemIDs))
	}

	p, err := domain.NewPersona(personaID, collectorID, t.Name.For(l), t.Description.For(l))
	if err != nil {
		return nil, err
	}

	for i, it := range t.Items {
		spec, err := it.spec(l)
		if err != nil {
			return nil, fmt.Errorf("item %q: %w", it.Title.For(DefaultLocale), err)
		}

		if _, err := p.AddItem(itemIDs[i], it.Title.For(l), it.Recurrence, spec); err != nil {
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
