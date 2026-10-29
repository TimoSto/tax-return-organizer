// Package idgen draws random ids for new entities.
package idgen

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// Random generates random (version 4) UUIDs, which uuid draws from crypto/rand. It implements outbound.IDGenerator.
type Random struct{}

var _ outbound.IDGenerator = Random{}

// NewCollectorID returns a random collector id.
func (Random) NewCollectorID() (domain.CollectorID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return domain.CollectorID{}, fmt.Errorf("generate collector id: %w", err)
	}

	return domain.CollectorID(id), nil
}

// NewPersonaID returns a random persona id.
func (Random) NewPersonaID() (domain.PersonaID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return domain.PersonaID{}, fmt.Errorf("generate persona id: %w", err)
	}

	return domain.PersonaID(id), nil
}

// NewItemID returns a random item id.
func (Random) NewItemID() (domain.ItemID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return domain.ItemID{}, fmt.Errorf("generate item id: %w", err)
	}

	return domain.ItemID(id), nil
}
