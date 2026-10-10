package postgres_test

import (
	"context"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports/porttest"
)

func TestCollectorRepository(t *testing.T) {
	pool := requirePool(t)

	porttest.TestCollectorRepository(t, func(t *testing.T) ports.CollectorRepository {
		if _, err := pool.Exec(context.Background(), `TRUNCATE collectors`); err != nil {
			t.Fatalf("truncate collectors: %v", err)
		}

		return postgres.NewCollectorRepository(pool)
	})
}
