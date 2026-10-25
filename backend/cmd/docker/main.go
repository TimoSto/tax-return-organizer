// Command docker runs the backend as configured by environment variables.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/TimoSto/tax-return-organizer/backend/internal/backend"
	"github.com/TimoSto/tax-return-organizer/backend/internal/logging"
)

func main() {
	cfg := backend.ConfigFromEnv(os.Getenv)
	logging.Setup(os.Stderr, cfg.LogLevel)

	if err := run(cfg, os.Args[1:]); err != nil {
		slog.Error("backend stopped", "error", err)
		os.Exit(1)
	}
}

// run starts the server, or only applies the migrations when called with the `migrate` argument.
func run(cfg backend.Config, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(args) > 0 && args[0] == "migrate" {
		return backend.Migrate(ctx, cfg)
	}

	return backend.Run(ctx, cfg)
}
