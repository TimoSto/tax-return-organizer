package persona_test

import (
	"context"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/porttest"
)

func newStore(t *testing.T, collectors ...domain.CollectorID) outbound.PersonaStore {
	repo := collector.NewRepository()

	for _, id := range collectors {
		if err := repo.SaveCollector(context.Background(), domaintest.Collector(id, "Collector "+id.String(), domain.Email(id.String()+"@example.com"))); err != nil {
			t.Fatalf("seed collector %s: %v", id, err)
		}
	}

	return persona.NewStore(repo)
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

func TestPersonaRepositoryUnknownCollector(t *testing.T) {
	porttest.TestPersonaRepositoryUnknownCollector(t, newStore)
}
