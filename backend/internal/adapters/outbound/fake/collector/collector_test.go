package collector_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/porttest"
)

func TestCollectorRepository(t *testing.T) {
	porttest.TestCollectorRepository(t, func(*testing.T) port.CollectorRepository {
		return collector.NewRepository()
	})
}
