package httpadapter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(log *slog.Logger, w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error("write response", "error", err)
	}
}

// writeError maps err to an HTTP status. Domain errors are reported to the client;
// everything else is logged and answered with a generic 500.
func writeError(log *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNameRequired), errors.Is(err, domain.ErrInvalidEmail):
		writeJSON(log, w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrCollectorNotFound):
		writeJSON(log, w, http.StatusNotFound, errorResponse{Error: err.Error()})
	default:
		log.ErrorContext(r.Context(), "request failed", "error", err, "request_id", w.Header().Get(requestIDHeader))
		writeJSON(log, w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}
}
