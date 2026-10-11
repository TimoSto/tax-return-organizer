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
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

const shutdownTimeout = 10 * time.Second

// newMux registers the routes of all inbound handlers.
func newMux(createCollector ports.CreateCollectorPort) *http.ServeMux {
	mux := http.NewServeMux()
	httpadapter.RegisterCollectorRoutes(mux, createCollector)

	return mux
}

// Run migrates the database and serves HTTP until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	if cfg.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultHTTPAddr
	}

	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	collectors := postgres.NewCollectorRepository(pool)
	createCollector := createcollector.New(collectors)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           newMux(createCollector.Execute),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr)
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
