// Package httpadapter exposes the inbound ports over HTTP.
package httpadapter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
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

type errorResponse struct {
	Error string `json:"error"`
}

// NewHandler returns the HTTP routes backed by the given inbound ports.
func NewHandler(createCollector ports.CreateCollectorPort) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /collectors", createCollectorHandler(createCollector))

	return mux
}

func createCollectorHandler(create ports.CreateCollectorPort) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createCollectorRequest

		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
		dec.DisallowUnknownFields()

		if err := dec.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})

			return
		}

		id, err := create(r.Context(), req.Name, req.Email)
		if err != nil {
			if errors.Is(err, createcollector.ErrNameRequired) || errors.Is(err, domain.ErrInvalidEmail) {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})

				return
			}

			slog.ErrorContext(r.Context(), "create collector", "error", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

			return
		}

		w.Header().Set("Location", "/collectors/"+id)
		writeJSON(w, http.StatusCreated, createCollectorResponse{ID: id})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "error", err)
	}
}
