//go:build e2e

// Command e2e runs the backend against a throwaway Postgres testcontainer.
// The database is wiped on every run; it needs a running Docker daemon.
// Every request with "Authorization: Bearer e2e-employee-a" acts as employeeA, an employee of the seeded collector.
// It is excluded from normal builds: use -tags e2e.
package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/idgen"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres"
	"github.com/TimoSto/tax-return-organizer/backend/internal/backend"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/logging"
)

// defaultCollectorID is the id of the collector seeded on every run, so e2e tests can reference it.
var defaultCollectorID = domain.CollectorID(uuid.MustParse("00000000-0000-4000-8000-00000000aaaa"))

// employeeAID is the user id of employeeA, the employee of the default collector that sessionToken logs in as.
// There are no accounts yet, so it exists only in the actor.
var employeeAID = domain.UserID(uuid.MustParse("00000000-0000-4000-8000-0000000000a1"))

// sessionToken is the bearer token of employeeA's session. It never expires. It stands in for a login until
// cmd/e2e seeds real accounts (docs/AUTHENTICATION.md).
const sessionToken = "e2e-employee-a" //nolint:gosec // a public test token, compiled only with the e2e tag

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
	cfg.Sessions = staticSessions{}

	return backend.Run(ctx, cfg)
}

// seed migrates the database and creates the default collector, with its predefined personas, through the
// create-collector use case so the seed follows any change to how collectors are created.
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

	create := backend.NewCreateCollector(backend.NewTransactor(pool), seedIDs{})

	if _, err := create.Execute(ctx, inbound.CreateCollectorCommand{Name: "E2E Collector", Email: "e2e@example.com"}); err != nil {
		return fmt.Errorf("create collector: %w", err)
	}

	return nil
}

// seedIDs draws the well-known collector id and random ids otherwise.
type seedIDs struct{ idgen.Random }

func (seedIDs) NewCollectorID() (domain.CollectorID, error) { return defaultCollectorID, nil }

// staticSessions resolves sessionToken to employeeA and no other token.
type staticSessions struct{}

func (staticSessions) ResolveSession(_ context.Context, token string) (domain.Actor, bool, error) {
	if subtle.ConstantTimeCompare([]byte(token), []byte(sessionToken)) != 1 {
		return domain.Actor{}, false, nil
	}

	return domain.NewStaffActor(defaultCollectorID, employeeAID), true, nil
}
