package personatemplate

import (
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// Predefined returns the built-in personas every collector starts with. Each call returns a fresh slice.
func Predefined() []PersonaTemplate {
	return []PersonaTemplate{
		{
			Name:        Localized{LocaleEN: "Employed", LocaleDE: "Angestellt"},
			Description: Localized{LocaleEN: "Works as an employee", LocaleDE: "Arbeitet als Angestellte/r"},
			Items: []ItemTemplate{
				{
					Title:      Localized{LocaleEN: "Payslip", LocaleDE: "Gehaltsabrechnung"},
					Kind:       domain.KindDocument,
					Recurrence: domain.RecurrenceMonthly,
				},
				{
					Title:      Localized{LocaleEN: "Annual wage tax statement", LocaleDE: "Lohnsteuerbescheinigung"},
					Kind:       domain.KindDocument,
					Recurrence: domain.RecurrenceAnnually,
				},
				{
					Title:      Localized{LocaleEN: "Days worked at the office", LocaleDE: "Arbeitstage im Büro"},
					Kind:       domain.KindData,
					Recurrence: domain.RecurrenceAnnually,
					Value:      &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days", Required: true},
				},
			},
		},
		{
			Name:        Localized{LocaleEN: "Landlord", LocaleDE: "Vermieter"},
			Description: Localized{LocaleEN: "Rents out property", LocaleDE: "Vermietet Wohnraum"},
			Items: []ItemTemplate{
				{
					Title:      Localized{LocaleEN: "Rent receipt", LocaleDE: "Mietzahlung"},
					Kind:       domain.KindData,
					Recurrence: domain.RecurrenceMonthly,
					Value:      &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true},
				},
				{
					Title:      Localized{LocaleEN: "Maintenance invoice", LocaleDE: "Handwerkerrechnung"},
					Kind:       domain.KindDocument,
					Recurrence: domain.RecurrenceAnnually,
					MetadataFields: []MetadataFieldTemplate{
						{Name: Localized{LocaleEN: "Amount", LocaleDE: "Betrag"}, ValueSpec: domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true}},
					},
				},
			},
		},
	}
}
