// Package postgres contains the PostgreSQL implementations of the outbound ports.
package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies all pending schema migrations. dsn must be a postgres:// or postgresql:// URL.
// If ctx is cancelled, the migration in progress finishes and the remaining ones are skipped.
func Migrate(ctx context.Context, dsn string) error {
	migrateURL, err := pgx5URL(dsn)
	if err != nil {
		return err
	}

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, migrateURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	done := make(chan error, 1)
	go func() { done <- m.Up() }()

	select {
	case err = <-done:
	case <-ctx.Done():
		go func() { m.GracefulStop <- true }()

		<-done

		return ctx.Err()
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// pgx5URL rewrites a postgres(ql):// URL to the scheme of golang-migrate's pgx5 driver.
func pgx5URL(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", errors.New("invalid dsn")
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", errors.New("dsn must start with postgres:// or postgresql://")
	}

	u.Scheme = "pgx5"

	return u.String(), nil
}
