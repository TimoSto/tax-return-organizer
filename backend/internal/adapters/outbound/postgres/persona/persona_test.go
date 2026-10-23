package persona_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

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

// A JSON null literal passes the "spec matches kind" check, so only the object check can reject it.
func TestItemDefinitions_SpecsMustBeJSONObjects(t *testing.T) {
	pool := postgrestest.RequirePool(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `TRUNCATE personas, collectors CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	personaID := uuid.New()

	if _, err := pool.Exec(ctx, `INSERT INTO collectors (id, name, email) VALUES ('aaaa', 'A', 'a@example.com')`); err != nil {
		t.Fatalf("seed collector: %v", err)
	}

	if _, err := pool.Exec(ctx, `INSERT INTO personas (id, collector_id, name) VALUES ($1, 'aaaa', 'Employed')`, personaID); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	for _, tc := range []struct {
		name, kind, column, value string
		wantErr                   bool
	}{
		{"document spec object", "document", "document_spec", `{}`, false},
		{"document spec json null", "document", "document_spec", `null`, true},
		{"document spec array", "document", "document_spec", `[]`, true},
		{"value spec object", "data", "value_spec", `{}`, false},
		{"value spec json null", "data", "value_spec", `null`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pool.Exec(ctx,
				`INSERT INTO item_definitions (id, persona_id, position, title, kind, recurrence, `+tc.column+`)
				 VALUES ($1, $2, 0, 'Item', $3, 'monthly', $4::jsonb)`,
				uuid.New(), personaID, tc.kind, tc.value)

			if (err != nil) != tc.wantErr {
				t.Errorf("insert error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
