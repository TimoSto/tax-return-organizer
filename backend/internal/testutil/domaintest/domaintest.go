// Package domaintest builds valid domain entities for tests. The constructors panic on invalid input, so a broken
// fixture fails loudly instead of being silently ignored.
package domaintest

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// TaxPractice returns a valid tax practice.
func TaxPractice(id domain.TaxPracticeID, name string, email domain.Email) *domain.TaxPractice {
	c, err := domain.RehydrateTaxPractice(id, name, email)
	if err != nil {
		panic(err)
	}

	return c
}

// Item returns a valid item definition. Instance fields make it repeatable.
func Item(
	id domain.ItemID, title string, recurrence domain.Recurrence, spec domain.ItemSpec,
	instanceFields ...domain.MetadataField,
) domain.ItemDefinition {
	item, err := domain.RehydrateItemDefinition(id, title, recurrence, spec, instanceFields)
	if err != nil {
		panic(err)
	}

	return item
}

// Persona returns a valid persona with the given items.
func Persona(
	id domain.PersonaID, taxPracticeID domain.TaxPracticeID, name, description string, items ...domain.ItemDefinition,
) *domain.Persona {
	p, err := domain.RehydratePersona(id, taxPracticeID, name, description, items)
	if err != nil {
		panic(err)
	}

	return p
}

// TaxPracticeID returns a random tax practice id.
func TaxPracticeID() domain.TaxPracticeID {
	return domain.TaxPracticeID(uuid.New())
}

// PersonaID returns a random persona id.
func PersonaID() domain.PersonaID {
	return domain.PersonaID(uuid.New())
}

// ItemID returns a random item id.
func ItemID() domain.ItemID {
	return domain.ItemID(uuid.New())
}

// Staff returns an employee of the tax practice with a random user id.
func Staff(taxPracticeID domain.TaxPracticeID) domain.Actor {
	return domain.NewStaffActor(taxPracticeID, domain.UserID(uuid.New()))
}

// Client returns a client of the tax practice with a random user id.
func Client(taxPracticeID domain.TaxPracticeID) domain.Actor {
	return domain.NewClientActor(taxPracticeID, domain.UserID(uuid.New()))
}

// CmpOptions makes go-cmp compare entities, which have no exported fields, by their contents. A nil and an empty
// slice are equal, because repositories may report either.
func CmpOptions() cmp.Option {
	return cmp.Options{
		cmp.AllowUnexported(domain.TaxPractice{}, domain.Persona{}, domain.ItemDefinition{}),
		cmpopts.EquateEmpty(),
	}
}
