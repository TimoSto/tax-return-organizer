// Package pgtest starts a throwaway Postgres container for tests.
package pgtest

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// DockerAvailable reports whether a Docker daemon is reachable.
func DockerAvailable(ctx context.Context) bool {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return false
	}
	defer func() { _ = provider.Close() }()

	return provider.Health(ctx) == nil
}

// Start runs a Postgres container and returns its DSN and a function that terminates it.
func Start(ctx context.Context) (dsn string, terminate func(), err error) {
	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("start postgres container: %w", err)
	}

	terminate = func() { _ = container.Terminate(context.WithoutCancel(ctx)) }

	dsn, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		terminate()

		return "", nil, fmt.Errorf("connection string: %w", err)
	}

	return dsn, terminate, nil
}
