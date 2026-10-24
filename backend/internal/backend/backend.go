// Package backend is the composition root: it wires the adapters and usecases together and serves HTTP.
package backend

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres"
	pgcollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/collector"
	pgpersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/templates"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getpersona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/listpersonas"
)

const shutdownTimeout = 10 * time.Second

// postgresRepositories returns one repository per resource on db, which is the pool or a transaction.
// A new resource adds a field to port.Repositories and a line here.
func postgresRepositories(db pgdb.DBTX) port.Repositories {
	return port.Repositories{
		Collectors: pgcollector.NewRepository(db),
		Personas:   pgpersona.NewRepository(db),
	}
}

// wire builds the use cases on top of the given repositories and returns the HTTP handler exposing them.
// Reads use repos directly, writes that span several repositories go through tx.
func wire(repos port.Repositories, tx port.Transactor) http.Handler {
	return httpadapter.NewHandler(httpadapter.Deps{
		CreateCollector: createcollector.New(tx, templates.New()),
		GetCollector:    getcollector.New(repos.Collectors),
		GetPersona:      getpersona.New(repos.Personas),
		ListPersonas:    listpersonas.New(repos.Personas),
	})
}

// Run migrates the database and serves HTTP until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if err := postgres.Migrate(ctx, cfg.DatabaseURL); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	return serve(ctx, newServer(cfg.HTTPAddr, wire(postgresRepositories(pool), pgdb.NewTransactor(pool, postgresRepositories))))
}

func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// serve runs srv until it fails or ctx is cancelled, then shuts it down gracefully.
func serve(ctx context.Context, srv *http.Server) error {
	serveErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}
