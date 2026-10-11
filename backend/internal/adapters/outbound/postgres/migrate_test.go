package postgres_test

import (
	"context"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres"
)

func TestMigrate_RejectsBadDSN(t *testing.T) {
	for _, dsn := range []string{"mysql://x", "not a url", ""} {
		if err := postgres.Migrate(context.Background(), dsn); err == nil || !strings.Contains(err.Error(), "dsn") {
			t.Errorf("Migrate(%q) err = %v, want dsn error", dsn, err)
		}
	}
}
