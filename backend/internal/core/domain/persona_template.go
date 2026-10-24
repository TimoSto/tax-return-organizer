package domain

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
	Kind           Kind
	Recurrence     Recurrence
	MetadataFields []MetadataFieldTemplate
	Value          *ValueSpec
}

// MetadataFieldTemplate is a predefined metadata field of a document item.
type MetadataFieldTemplate struct {
	Name      Localized
	ValueSpec ValueSpec
}

// Instantiate creates a persona of the collector from the template, with all texts in locale l.
func (t PersonaTemplate) Instantiate(collectorID CollectorID, l Locale) (*Persona, error) {
	p, err := NewPersona(collectorID, t.Name.For(l), t.Description.For(l))
	if err != nil {
		return nil, err
	}

	for _, it := range t.Items {
		var spec ItemSpec

		switch it.Kind {
		case KindDocument:
			doc := DocumentSpec{}
			for _, f := range it.MetadataFields {
				doc.MetadataFields = append(doc.MetadataFields, MetadataField{Name: f.Name.For(l), ValueSpec: f.ValueSpec})
			}

			spec = doc
		case KindData:
			spec = *it.Value
		}

		if _, err := p.AddItem(it.Title.For(l), it.Recurrence, spec); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// PersonaTemplates returns the predefined personas. Every collector starts with them.
func PersonaTemplates() []PersonaTemplate {
	return []PersonaTemplate{
		{
			Name:        Localized{LocaleEN: "Employed", LocaleDE: "Angestellt"},
			Description: Localized{LocaleEN: "Works as an employee", LocaleDE: "Arbeitet als Angestellte/r"},
			Items: []ItemTemplate{
				{
					Title:      Localized{LocaleEN: "Payslip", LocaleDE: "Gehaltsabrechnung"},
					Kind:       KindDocument,
					Recurrence: RecurrenceMonthly,
				},
				{
					Title:      Localized{LocaleEN: "Annual wage tax statement", LocaleDE: "Lohnsteuerbescheinigung"},
					Kind:       KindDocument,
					Recurrence: RecurrenceAnnually,
				},
				{
					Title:      Localized{LocaleEN: "Days worked at the office", LocaleDE: "Arbeitstage im Büro"},
					Kind:       KindData,
					Recurrence: RecurrenceAnnually,
					Value:      &ValueSpec{Type: ValueTypeNumber, Unit: "days", Required: true},
				},
			},
		},
		{
			Name:        Localized{LocaleEN: "Landlord", LocaleDE: "Vermieter"},
			Description: Localized{LocaleEN: "Rents out property", LocaleDE: "Vermietet Wohnraum"},
			Items: []ItemTemplate{
				{
					Title:      Localized{LocaleEN: "Rent receipt", LocaleDE: "Mietzahlung"},
					Kind:       KindData,
					Recurrence: RecurrenceMonthly,
					Value:      &ValueSpec{Type: ValueTypeNumber, Unit: "EUR", Required: true},
				},
				{
					Title:      Localized{LocaleEN: "Maintenance invoice", LocaleDE: "Handwerkerrechnung"},
					Kind:       KindDocument,
					Recurrence: RecurrenceAnnually,
					MetadataFields: []MetadataFieldTemplate{
						{Name: Localized{LocaleEN: "Amount", LocaleDE: "Betrag"}, ValueSpec: ValueSpec{Type: ValueTypeNumber, Unit: "EUR", Required: true}},
					},
				},
			},
		},
	}
}
