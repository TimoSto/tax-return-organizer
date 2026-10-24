// Package logging configures the process-wide logger. The entrypoints call Setup once on start;
// everything else logs through log/slog's package-level functions.
package logging

import (
	"io"
	"log/slog"
)

// Setup makes a text logger that adds the request id of the log context,  writing to w at the given level the default slog logger.
func Setup(w io.Writer, level slog.Level) {
	slog.SetDefault(slog.New(NewHandler(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))))
}
