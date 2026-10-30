package personatemplate

import (
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/locale"
)

// Predefined returns the built-in personas every collector starts with. Each call returns a fresh slice.
func Predefined() []PersonaTemplate {
	return []PersonaTemplate{
		{
			Name:        locale.Localized{locale.LocaleEN: "Employed", locale.LocaleDE: "Angestellt"},
			Description: locale.Localized{locale.LocaleEN: "Works as an employee", locale.LocaleDE: "Arbeitet als Angestellte/r"},
			Items: []ItemTemplate{
				{
					Title:      locale.Localized{locale.LocaleEN: "Payslip", locale.LocaleDE: "Gehaltsabrechnung"},
					Recurrence: domain.RecurrenceMonthly,
					Spec:       DocumentSpecTemplate{},
				},
				{
					Title:      locale.Localized{locale.LocaleEN: "Annual wage tax statement", locale.LocaleDE: "Lohnsteuerbescheinigung"},
					Recurrence: domain.RecurrenceAnnually,
					Spec:       DocumentSpecTemplate{},
				},
				{
					Title:      locale.Localized{locale.LocaleEN: "Days worked at the office", locale.LocaleDE: "Arbeitstage im Büro"},
					Recurrence: domain.RecurrenceAnnually,
					Spec:       ValueSpecTemplate{Type: domain.ValueTypeNumber, Unit: "days", Required: true},
				},
			},
		},
		{
			Name:        locale.Localized{locale.LocaleEN: "Landlord", locale.LocaleDE: "Vermieter"},
			Description: locale.Localized{locale.LocaleEN: "Rents out property", locale.LocaleDE: "Vermietet Wohnraum"},
			Items: []ItemTemplate{
				{
					Title:      locale.Localized{locale.LocaleEN: "Rent receipt", locale.LocaleDE: "Mietzahlung"},
					Recurrence: domain.RecurrenceMonthly,
					Spec:       ValueSpecTemplate{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true},
				},
				{
					Title:      locale.Localized{locale.LocaleEN: "Maintenance invoice", locale.LocaleDE: "Handwerkerrechnung"},
					Recurrence: domain.RecurrenceAnnually,
					Spec: DocumentSpecTemplate{MetadataFields: []MetadataFieldTemplate{
						{Name: locale.Localized{locale.LocaleEN: "Amount", locale.LocaleDE: "Betrag"}, ValueSpec: domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true}},
					}},
				},
			},
		},
	}
}
