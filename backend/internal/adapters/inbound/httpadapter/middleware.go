package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// Middleware wraps h with request-ID assignment, request logging and panic recovery.
func Middleware(log *slog.Logger, h http.Handler) http.Handler {
	return requestID(requestLog(log, recoverPanic(log, h)))
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(requestIDHeader, uuid.NewString())
		next.ServeHTTP(w, r)
	})
}

func requestLog(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		log.InfoContext(r.Context(), "request",
			"request_id", w.Header().Get(requestIDHeader),
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start),
		)
	})
}

func recoverPanic(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}

				log.ErrorContext(r.Context(), "panic", "value", v, "request_id", w.Header().Get(requestIDHeader))
				writeJSON(log, w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
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
