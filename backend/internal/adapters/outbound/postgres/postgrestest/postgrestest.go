// Package postgrestest provides a shared PostgreSQL test container for repository tests.
package postgrestest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres"
)

// testPool is nil if Docker is unavailable; tests must call RequirePool.
var testPool *pgxpool.Pool

// Run starts a migrated PostgreSQL container, runs the tests and returns the exit code.
// Call it from TestMain: os.Exit(postgrestest.Run(m)).
func Run(m *testing.M) int {
	ctx := context.Background()

	if !dockerAvailable(ctx) {
		fmt.Fprintln(os.Stderr, "docker unavailable, postgres tests will be skipped")
		return m.Run()
	}

	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres container: %v\n", err)
		return 1
	}
	defer func() { _ = container.Terminate(ctx) }()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection string: %v\n", err)
		return 1
	}

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

func dockerAvailable(ctx context.Context) bool {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return false
	}
	defer provider.Close()

	return provider.Health(ctx) == nil
}

// RequirePool skips the test if no database is available and otherwise returns the pool.
func RequirePool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if testPool == nil {
		t.Skip("docker unavailable")
	}

	return testPool
}
