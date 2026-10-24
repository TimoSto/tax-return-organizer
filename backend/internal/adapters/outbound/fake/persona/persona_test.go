package persona_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports/porttest"
)

func newRepo(_ *testing.T, collectors ...domain.CollectorID) ports.PersonaRepository {
	return persona.NewRepository(collectors...)
}

func TestPersonaRepository(t *testing.T) {
	porttest.TestPersonaRepository(t, newRepo)
}

func TestPersonaRepositoryList(t *testing.T) {
	porttest.TestPersonaRepositoryList(t, newRepo)
}
