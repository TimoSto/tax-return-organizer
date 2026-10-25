// Package domaintest builds valid domain entities for tests. The constructors panic on invalid input, so a broken
// fixture fails loudly instead of being silently ignored.
package domaintest

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// Collector returns a valid collector.
func Collector(id domain.CollectorID, name string, email domain.Email) *domain.Collector {
	c, err := domain.RehydrateCollector(id, name, email)
	if err != nil {
		panic(err)
	}

	return c
}

// Item returns a valid item definition.
func Item(id domain.ItemID, title string, recurrence domain.Recurrence, spec domain.ItemSpec) domain.ItemDefinition {
	item, err := domain.RehydrateItemDefinition(id, title, recurrence, spec)
	if err != nil {
		panic(err)
	}

	return item
}

// Persona returns a valid persona with the given items.
func Persona(
	id domain.PersonaID, collectorID domain.CollectorID, name, description string, items ...domain.ItemDefinition,
) *domain.Persona {
	p, err := domain.RehydratePersona(id, collectorID, name, description, items)
	if err != nil {
		panic(err)
	}

	return p
}

// PersonaID returns a random persona id.
func PersonaID() domain.PersonaID {
	return domain.PersonaID(uuid.New())
}

// ItemID returns a random item id.
func ItemID() domain.ItemID {
	return domain.ItemID(uuid.New())
}

// CmpOptions makes go-cmp compare entities, which have no exported fields, by their contents. A nil and an empty
// slice are equal, because repositories may report either.
func CmpOptions() cmp.Option {
	return cmp.Options{
		cmp.AllowUnexported(domain.Collector{}, domain.Persona{}, domain.ItemDefinition{}),
		cmpopts.EquateEmpty(),
	}
}
