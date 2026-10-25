package outbound

import "github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"

// IDGenerator creates the ids of new entities. Persona and item ids are random and effectively unique; a collector
// id is short, so it is not guaranteed to be unused and a caller that saves it must handle domain.ErrCollectorIDTaken.
type IDGenerator interface {
	NewCollectorID() (domain.CollectorID, error)
	NewPersonaID() (domain.PersonaID, error)
	NewItemID() (domain.ItemID, error)
}
