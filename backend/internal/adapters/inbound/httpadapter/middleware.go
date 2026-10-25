package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/logging"
)

// Middleware wraps h with request-ID assignment, request logging and panic recovery.
func Middleware(h http.Handler) http.Handler {
	return requestID(requestLog(recoverPanic(h)))
}

// validRequestID limits incoming ids to a safe alphabet and length, so they cannot inject into the logs.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// requestID takes the id from the incoming header if it is valid, otherwise it generates one.
// The id is returned in the response header and put in the context, where the logger picks it up.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(httpx.RequestIDHeader)
		if !validRequestID.MatchString(id) {
			id = uuid.NewString()
		}

		w.Header().Set(httpx.RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(logging.WithRequestID(r.Context(), id)))
	})
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		// The mux sets Pattern on r; the raw path would have unbounded cardinality.
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}

		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"route", route,
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

				slog.ErrorContext(r.Context(), "panic", "value", v)
				httpx.WriteProblem(w, r, http.StatusInternalServerError, httpx.CodeInternal, "internal error")
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
