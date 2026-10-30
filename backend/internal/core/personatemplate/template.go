// Package personatemplate holds predefined personas and the logic to copy them to a collector.
package personatemplate

import (
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/locale"
)

// PersonaTemplate is a predefined persona that is copied to a collector, in the collector's locale.
// The copy is a normal persona that the collector owns and may change, nothing links it back to the template.
type PersonaTemplate struct {
	Name        locale.Localized
	Description locale.Localized
	Items       []ItemTemplate
}

// ItemTemplate is a predefined item definition. Its kind follows from Spec.
type ItemTemplate struct {
	Title      locale.Localized
	Recurrence domain.Recurrence
	Spec       ItemSpecTemplate
}

// ItemSpecTemplate is what a predefined item asks for, the localized counterpart of domain.ItemSpec. It is a sum
// type that only DocumentSpecTemplate (a document item) and ValueSpecTemplate (a data item) implement.
type ItemSpecTemplate interface {
	// spec returns the domain spec with all texts in locale l.
	spec(l locale.Locale) domain.ItemSpec
}

// DocumentSpecTemplate makes a document item.
type DocumentSpecTemplate struct {
	// MetadataFields may be nil or empty if the document has none.
	MetadataFields []MetadataFieldTemplate
}

func (s DocumentSpecTemplate) spec(l locale.Locale) domain.ItemSpec {
	doc := domain.DocumentSpec{}
	for _, f := range s.MetadataFields {
		doc.MetadataFields = append(doc.MetadataFields, domain.MetadataField{Name: f.Name.For(l), ValueSpec: f.ValueSpec})
	}

	return doc
}

// ValueSpecTemplate makes a data item. A value spec has no texts to localize, so it is a domain.ValueSpec.
type ValueSpecTemplate domain.ValueSpec

func (s ValueSpecTemplate) spec(locale.Locale) domain.ItemSpec { return domain.ValueSpec(s) }

// MetadataFieldTemplate is a predefined metadata field of a document item.
type MetadataFieldTemplate struct {
	Name      locale.Localized
	ValueSpec domain.ValueSpec
}

// Instantiate creates a persona of the collector from the template, with all texts in locale l.
// It is pure: the caller supplies the persona id and one item id per item of the template, in order.
func (t PersonaTemplate) Instantiate(collectorID domain.CollectorID, l locale.Locale, personaID domain.PersonaID, itemIDs []domain.ItemID) (*domain.Persona, error) {
	if len(itemIDs) != len(t.Items) {
		return nil, fmt.Errorf("persona template %q has %d items, got %d item ids", t.Name.For(locale.DefaultLocale), len(t.Items), len(itemIDs))
	}

	p, err := domain.NewPersona(personaID, collectorID, t.Name.For(l), t.Description.For(l))
	if err != nil {
		return nil, err
	}

	for i, it := range t.Items {
		if it.Spec == nil {
			return nil, fmt.Errorf("item template %q has no spec", it.Title.For(locale.DefaultLocale))
		}

		if _, err := p.AddItem(itemIDs[i], it.Title.For(l), it.Recurrence, it.Spec.spec(l)); err != nil {
			return nil, err
		}
	}

	return p, nil
}
