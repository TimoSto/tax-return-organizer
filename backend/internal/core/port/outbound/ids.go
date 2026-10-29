package outbound

import "github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"

// IDGenerator creates the ids of new entities. The ids are random and effectively unique.
type IDGenerator interface {
	NewCollectorID() (domain.CollectorID, error)
	NewPersonaID() (domain.PersonaID, error)
	NewItemID() (domain.ItemID, error)
}
