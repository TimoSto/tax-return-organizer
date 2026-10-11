// Package httpadapter exposes the inbound ports over HTTP.
package httpadapter

import (
	"log/slog"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

const maxBodyBytes = 1 << 20

// RegisterCollectorRoutes registers the collector routes on mux.
func RegisterCollectorRoutes(mux *http.ServeMux, log *slog.Logger, createCollector ports.CreateCollector) {
	mux.HandleFunc("POST /collectors", createCollectorHandler(log, createCollector))
}
