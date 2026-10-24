//go:build e2e

// Command e2e runs the backend against a throwaway Postgres testcontainer.
// The database is wiped on every run; it needs a running Docker daemon.
// It is excluded from normal builds: use -tags e2e.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres"
	pgcollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/backend"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/logging"
)

// defaultCollectorID is the id of the collector seeded on every run, so e2e tests can reference it.
var defaultCollectorID = domain.CollectorID("aaaa")

func main() {
	logging.Setup(os.Stderr, slog.LevelInfo)

	if err := run(); err != nil {
		slog.Error("e2e stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return runWithContext(ctx, os.Getenv)
}

func runWithContext(ctx context.Context, getenv func(string) string) error {
	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("e2e"),
		tcpostgres.WithUsername("e2e"),
		tcpostgres.WithPassword("e2e"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return fmt.Errorf("start postgres container: %w", err)
	}
	defer func() {
		termCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()

		if err := container.Terminate(termCtx); err != nil {
			slog.Error("terminate postgres container", "error", err)
		}
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("connection string: %w", err)
	}

	if err := seed(ctx, dsn); err != nil {
		return fmt.Errorf("seed: %w", err)
	}

	cfg := backend.ConfigFromEnv(getenv)
	cfg.DatabaseURL = dsn

	return backend.Run(ctx, cfg)
}

// seed migrates the database and inserts the default collector.
// backend.Run migrates again afterwards, which is a no-op.
func seed(ctx context.Context, dsn string) error {
	if err := postgres.Migrate(ctx, dsn); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	c, err := domain.NewCollector(defaultCollectorID, "E2E Collector", "e2e@example.com")
	if err != nil {
		return fmt.Errorf("new collector: %w", err)
	}

	if err := pgcollector.NewRepository(pool).SaveCollector(ctx, c); err != nil {
		return fmt.Errorf("save collector: %w", err)
	}

	return nil
}
