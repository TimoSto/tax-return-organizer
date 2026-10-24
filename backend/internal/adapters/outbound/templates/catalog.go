// Package templates provides the predefined persona templates every collector starts with.
package templates

import (
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	pt "github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
)

// Catalog is the built-in set of persona templates.
type Catalog struct{}

// New returns the built-in catalog.
func New() Catalog { return Catalog{} }

// PersonaTemplates returns the predefined personas. Every collector starts with them.
func (Catalog) PersonaTemplates() []pt.PersonaTemplate {
	return []pt.PersonaTemplate{
		{
			Name:        pt.Localized{pt.LocaleEN: "Employed", pt.LocaleDE: "Angestellt"},
			Description: pt.Localized{pt.LocaleEN: "Works as an employee", pt.LocaleDE: "Arbeitet als Angestellte/r"},
			Items: []pt.ItemTemplate{
				{
					Title:      pt.Localized{pt.LocaleEN: "Payslip", pt.LocaleDE: "Gehaltsabrechnung"},
					Kind:       domain.KindDocument,
					Recurrence: domain.RecurrenceMonthly,
				},
				{
					Title:      pt.Localized{pt.LocaleEN: "Annual wage tax statement", pt.LocaleDE: "Lohnsteuerbescheinigung"},
					Kind:       domain.KindDocument,
					Recurrence: domain.RecurrenceAnnually,
				},
				{
					Title:      pt.Localized{pt.LocaleEN: "Days worked at the office", pt.LocaleDE: "Arbeitstage im Büro"},
					Kind:       domain.KindData,
					Recurrence: domain.RecurrenceAnnually,
					Value:      &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days", Required: true},
				},
			},
		},
		{
			Name:        pt.Localized{pt.LocaleEN: "Landlord", pt.LocaleDE: "Vermieter"},
			Description: pt.Localized{pt.LocaleEN: "Rents out property", pt.LocaleDE: "Vermietet Wohnraum"},
			Items: []pt.ItemTemplate{
				{
					Title:      pt.Localized{pt.LocaleEN: "Rent receipt", pt.LocaleDE: "Mietzahlung"},
					Kind:       domain.KindData,
					Recurrence: domain.RecurrenceMonthly,
					Value:      &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true},
				},
				{
					Title:      pt.Localized{pt.LocaleEN: "Maintenance invoice", pt.LocaleDE: "Handwerkerrechnung"},
					Kind:       domain.KindDocument,
					Recurrence: domain.RecurrenceAnnually,
					MetadataFields: []pt.MetadataFieldTemplate{
						{Name: pt.Localized{pt.LocaleEN: "Amount", pt.LocaleDE: "Betrag"}, ValueSpec: domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR", Required: true}},
					},
				},
			},
		},
	}
}
