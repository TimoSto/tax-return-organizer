// Package httpadapter exposes the inbound ports over HTTP.
package httpadapter

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

const maxBodyBytes = 1 << 20

type createCollectorRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type createCollectorResponse struct {
	ID string `json:"id"`
}

// RegisterCollectorRoutes registers the collector routes on mux.
func RegisterCollectorRoutes(mux *http.ServeMux, log *slog.Logger, createCollector ports.CreateCollector) {
	mux.HandleFunc("POST /collectors", createCollectorHandler(log, createCollector))
}

func createCollectorHandler(log *slog.Logger, create ports.CreateCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createCollectorRequest

		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
		dec.DisallowUnknownFields()

		if err := dec.Decode(&req); err != nil {
			writeJSON(log, w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})

			return
		}

		id, err := create.Execute(r.Context(), ports.CreateCollectorCmd{Name: req.Name, Email: req.Email})
		if err != nil {
			writeError(log, w, r, err)

			return
		}

		w.Header().Set("Location", "/collectors/"+id.String())
		writeJSON(log, w, http.StatusCreated, createCollectorResponse{ID: id.String()})
	}
}
