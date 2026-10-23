package persona_test

import (
	"context"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/postgrestest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports/porttest"
)

func TestPersonaRepository(t *testing.T) {
	pool := postgrestest.RequirePool(t)

	porttest.TestPersonaRepository(t, func(t *testing.T, collectors ...domain.CollectorID) ports.PersonaRepository {
		ctx := context.Background()

		// personas reference collectors, so both tables are reset and the collectors re-seeded.
		if _, err := pool.Exec(ctx, `TRUNCATE personas, collectors CASCADE`); err != nil {
			t.Fatalf("truncate: %v", err)
		}

		for _, id := range collectors {
			if _, err := pool.Exec(ctx,
				`INSERT INTO collectors (id, name, email) VALUES ($1, $2, $3)`,
				string(id), "Collector "+string(id), string(id)+"@example.com"); err != nil {
				t.Fatalf("seed collector %s: %v", id, err)
			}
		}

		return persona.NewRepository(pool)
	})
}
