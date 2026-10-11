// Package postgrestest provides a shared PostgreSQL test container for repository tests.
package postgrestest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/pgtest"
)

// testPool is nil if Docker is unavailable; tests must call RequirePool.
var testPool *pgxpool.Pool

// Run starts a migrated PostgreSQL container, runs the tests and returns the exit code.
// Call it from TestMain: os.Exit(postgrestest.Run(m)).
func Run(m *testing.M) int {
	ctx := context.Background()

	if !pgtest.DockerAvailable(ctx) {
		fmt.Fprintln(os.Stderr, "docker unavailable, postgres tests will be skipped")
		return m.Run()
	}

	dsn, terminate, err := pgtest.Start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer terminate()

	if err := postgres.Migrate(ctx, dsn); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		return 1
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		return 1
	}
	defer testPool.Close()

	return m.Run()
}

// RequirePool skips the test if no database is available and otherwise returns the pool.
func RequirePool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if testPool == nil {
		t.Skip("docker unavailable")
	}

	return testPool
}
