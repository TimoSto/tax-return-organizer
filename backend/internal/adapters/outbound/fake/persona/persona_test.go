package persona_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/porttest"
)

func newRepo(_ *testing.T, collectors ...domain.CollectorID) port.PersonaRepository {
	return persona.NewRepository(collectors...)
}

func TestPersonaRepository(t *testing.T) {
	porttest.TestPersonaRepository(t, newRepo)
}

func TestPersonaRepositoryList(t *testing.T) {
	porttest.TestPersonaRepositoryList(t, newRepo)
}
