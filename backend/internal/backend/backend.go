// Package backend is the composition root: it wires the adapters and usecases together and serves HTTP.
package backend

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/health"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/idgen"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres"
	pgcollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/collector"
	pgpersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getpersona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/listpersonas"
)

const shutdownTimeout = 10 * time.Second

// Repositories is the full set of repositories, one per resource. It exists only in the composition root: use cases
// get single repositories, or a transactor narrowed to the repositories they declare (see narrow).
// Resources owned by a collector are added as stores, never as plain repositories.
type Repositories struct {
	Collectors outbound.CollectorRepository
	Personas   outbound.PersonaStore
}

// postgresRepositories returns one repository per resource on db, which is the pool or a transaction.
// A new resource adds a field to Repositories and a line here.
func postgresRepositories(db pgdb.DBTX) Repositories {
	return Repositories{
		Collectors: pgcollector.NewRepository(db),
		Personas:   pgpersona.NewStore(db),
	}
}

// NewTransactor returns the Postgres transactor that use cases run their writes through.
func NewTransactor(pool *pgxpool.Pool) outbound.Transactor[Repositories] {
	return pgdb.NewTransactor(pool, postgresRepositories)
}

// narrowed is a transactor whose callbacks get only the repositories pick selects from the full set.
type narrowed[R any] struct {
	tx   outbound.Transactor[Repositories]
	pick func(Repositories) R
}

// narrow returns tx restricted to the repositories pick selects, for a use case that declares its own set.
func narrow[R any](tx outbound.Transactor[Repositories], pick func(Repositories) R) outbound.Transactor[R] {
	return narrowed[R]{tx: tx, pick: pick}
}

// InTx implements outbound.Transactor.
func (n narrowed[R]) InTx(ctx context.Context, fn func(ctx context.Context, repos R) error) error {
	return n.tx.InTx(ctx, func(ctx context.Context, repos Repositories) error {
		return fn(ctx, n.pick(repos))
	})
}

// NewCreateCollector returns the createcollector use case on tx with the predefined personas. ids creates the new ids.
func NewCreateCollector(tx outbound.Transactor[Repositories], ids outbound.IDGenerator) *createcollector.Usecase {
	repos := narrow(tx, func(r Repositories) createcollector.Repositories {
		return createcollector.Repositories{Collectors: r.Collectors, Personas: r.Personas}
	})

	return createcollector.New(repos, personatemplate.Predefined(), ids)
}

// wire builds the use cases on top of the given repositories and returns the HTTP handler exposing them.
// Reads use repos directly, writes that span several repositories go through tx. ready backs /readyz.
func wire(repos Repositories, tx outbound.Transactor[Repositories], ready health.Pinger) (http.Handler, error) {
	return httpadapter.NewHandler(httpadapter.Deps{
		CreateCollector: NewCreateCollector(tx, idgen.Random{}),
		GetCollector:    getcollector.New(repos.Collectors),
		GetPersona:      getpersona.New(repos.Personas),
		ListPersonas:    listpersonas.New(repos.Personas),
		Readiness:       ready,
	})
}

// Run migrates the database and serves HTTP until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if !cfg.SkipMigrate {
		if err := postgres.Migrate(ctx, cfg.DatabaseURL); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	pool, err := newPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	ready := &drainingPinger{inner: pool}

	handler, err := wire(postgresRepositories(pool), NewTransactor(pool), ready)
	if err != nil {
		return fmt.Errorf("wire: %w", err)
	}

	return serve(ctx, newServer(cfg.HTTPAddr, handler), ready.startDraining, cfg.DrainDelay)
}

// Migrate applies the pending schema migrations and returns, for running them as a separate job
// with DDL rights while the app itself runs with MIGRATE_ON_START=false.
func Migrate(ctx context.Context, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if err := postgres.Migrate(ctx, cfg.DatabaseURL); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}

// newPool opens the connection pool tuned by cfg and fails fast if the database is unreachable.
func newPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}

	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}

	if cfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return pool, nil
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

// serve runs srv until it fails or ctx is cancelled, then shuts it down gracefully: it calls drain, waits
// drainDelay so the orchestrator notices, and only then stops accepting connections and waits for in-flight requests.
func serve(ctx context.Context, srv *http.Server, drain func(), drainDelay time.Duration) error {
	// Requests keep the values of ctx but are not cancelled by it, which would abort them during the drain.
	// They are cancelled only when the graceful shutdown gives up.
	baseCtx, cancelRequests := context.WithCancel(context.WithoutCancel(ctx))
	defer cancelRequests()

	srv.BaseContext = func(net.Listener) context.Context { return baseCtx }

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

	slog.InfoContext(ctx, "draining", "delay", drainDelay)
	drain()

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve: %w", err)
	case <-time.After(drainDelay):
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		cancelRequests()

		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}
