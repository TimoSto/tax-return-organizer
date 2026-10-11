package httpadapter

import (
	"log/slog"
	"net/http"
)

// RegisterHealthRoutes registers GET /healthz, a liveness probe that does not touch dependencies.
func RegisterHealthRoutes(mux *http.ServeMux, log *slog.Logger) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(log, w, http.StatusOK, map[string]string{"status": "ok"})
	})
}
