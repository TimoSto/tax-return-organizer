// Command docker runs the backend as configured by environment variables.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/TimoSto/tax-return-organizer/backend/internal/backend"
)

func main() {
	if err := run(); err != nil {
		slog.Error("backend stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return backend.Run(ctx, backend.ConfigFromEnv(os.Getenv))
}
