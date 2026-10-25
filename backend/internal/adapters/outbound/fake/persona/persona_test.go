package persona_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/porttest"
)

func newStore(*testing.T, ...domain.CollectorID) outbound.PersonaStore {
	return persona.NewStore()
}

func TestPersonaRepository(t *testing.T) {
	porttest.TestPersonaRepository(t, newStore)
}

func TestPersonaRepositoryList(t *testing.T) {
	porttest.TestPersonaRepositoryList(t, newStore)
}

func TestPersonaRepositoryIsolation(t *testing.T) {
	porttest.TestPersonaRepositoryIsolation(t, newStore)
}
