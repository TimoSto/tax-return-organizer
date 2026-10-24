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
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "error", err)
	}
}

// ParsePathID parses the path value param with parse. On failure it answers 400 "invalid <what>" and returns false.
func ParsePathID[T any](w http.ResponseWriter, r *http.Request, param string, parse func(string) (T, error), what string) (T, bool) {
	id, err := parse(r.PathValue(param))
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid " + what})

		var zero T

		return zero, false
	}

	return id, true
}

// WriteError maps err to an HTTP status. Domain errors are reported to the client with their own
// message only (never the wrapped chain); everything else is logged and answered with a generic 500.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var de *domain.Error
	if errors.As(err, &de) {
		WriteJSON(w, statusFor(de.Kind), ErrorResponse{Error: de.Error()})

		return
	}

	slog.ErrorContext(r.Context(), "request failed", "error", err)
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
}

func statusFor(k domain.ErrorKind) int {
	switch k {
	case domain.ErrorInvalid:
		return http.StatusBadRequest
	case domain.ErrorNotFound:
		return http.StatusNotFound
	case domain.ErrorConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
