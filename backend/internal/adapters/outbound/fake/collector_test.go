package fake_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports/porttest"
)

func TestCollectorRepository(t *testing.T) {
	porttest.TestCollectorRepository(t, func(t *testing.T) ports.CollectorRepository {
		return fake.NewCollectorRepository()
	})
}
