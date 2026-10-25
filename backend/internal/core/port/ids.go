package port

import "github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"

// CollectorIDGenerator creates ids for new collectors. The id is not guaranteed to be unused.
type CollectorIDGenerator interface {
	NewCollectorID() (domain.CollectorID, error)
}

// PersonaIDGenerator creates ids for new personas.
type PersonaIDGenerator interface {
	NewPersonaID() (domain.PersonaID, error)
}

// ItemIDGenerator creates ids for new item definitions.
type ItemIDGenerator interface {
	NewItemID() (domain.ItemID, error)
}
