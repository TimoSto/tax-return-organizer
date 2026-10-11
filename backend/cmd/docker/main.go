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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := backend.Run(ctx, backend.ConfigFromEnv(os.Getenv)); err != nil {
		slog.Error("backend stopped", "error", err)
		stop()
		os.Exit(1)
	}
}
