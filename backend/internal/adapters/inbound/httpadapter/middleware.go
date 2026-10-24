package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
)

// Middleware wraps h with request-ID assignment, request logging and panic recovery.
func Middleware(h http.Handler) http.Handler {
	return requestID(requestLog(recoverPanic(h)))
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpx.RequestIDHeader, uuid.NewString())
		next.ServeHTTP(w, r)
	})
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		slog.InfoContext(r.Context(), "request",
			"request_id", w.Header().Get(httpx.RequestIDHeader),
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start),
		)
	})
}

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if errors.Is(asError(v), http.ErrAbortHandler) {
					panic(v)
				}

				slog.ErrorContext(r.Context(), "panic", "value", v, "request_id", w.Header().Get(httpx.RequestIDHeader))
				httpx.WriteJSON(w, http.StatusInternalServerError, httpx.ErrorResponse{Error: "internal error"})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// asError returns v as an error, or nil if it is not one.
func asError(v any) error {
	err, _ := v.(error)

	return err
}
