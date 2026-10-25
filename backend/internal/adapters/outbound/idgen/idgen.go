// Package idgen draws random ids for new entities.
package idgen

import (
	"crypto/rand"
	"fmt"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// Random generates random ids with crypto/rand. It implements outbound.IDGenerator.
type Random struct{}

var _ outbound.IDGenerator = Random{}

// NewCollectorID returns a random, short collector id. It is not guaranteed to be unused.
func (Random) NewCollectorID() (domain.CollectorID, error) {
	// Bytes at or above this limit are rejected, so the remaining range is a multiple of the
	// alphabet size and every symbol is equally likely.
	const limit = 256 - 256%len(domain.CollectorIDAlphabet)

	id := make([]byte, 0, domain.CollectorIDLength)
	buf := make([]byte, domain.CollectorIDLength*2)

	for len(id) < domain.CollectorIDLength {
		if _, err := rand.Read(buf); err != nil {
			return "", fmt.Errorf("generate collector id: %w", err)
		}

		for _, b := range buf {
			if int(b) >= limit {
				continue
			}

			id = append(id, domain.CollectorIDAlphabet[int(b)%len(domain.CollectorIDAlphabet)])
			if len(id) == domain.CollectorIDLength {
				break
			}
		}
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
