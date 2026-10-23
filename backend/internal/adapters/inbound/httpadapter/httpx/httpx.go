// Package httpx contains the HTTP helpers shared by all handler packages.
package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

const (
	// RequestIDHeader carries the id that ties a response to its log lines.
	RequestIDHeader = "X-Request-ID"
	// MaxBodyBytes limits the size of request bodies handlers read.
	MaxBodyBytes = 1 << 20
)

// Route is an endpoint: a ServeMux pattern ("POST /collectors") and its handler.
type Route struct {
	Pattern string
	Handler http.Handler
}

// ErrorResponse is the JSON body of every error answer.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON writes body as JSON with the given status.
func WriteJSON(log *slog.Logger, w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error("write response", "error", err)
	}
}

// WriteError maps err to an HTTP status. Domain errors are reported to the client;
// everything else is logged and answered with a generic 500.
func WriteError(log *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNameRequired), errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidCollectorID):
		WriteJSON(log, w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrCollectorNotFound), errors.Is(err, domain.ErrPersonaNotFound):
		WriteJSON(log, w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	default:
		log.ErrorContext(r.Context(), "request failed", "error", err, "request_id", w.Header().Get(RequestIDHeader))
		WriteJSON(log, w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
	}
}
