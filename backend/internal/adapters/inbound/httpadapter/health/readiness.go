package health

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
)

// readinessTimeout bounds the dependency check, so a hanging database fails the probe instead of stalling it.
const readinessTimeout = 2 * time.Second

// ErrDraining is what a Pinger returns while the server is shutting down. It makes the probe fail without being
// logged as an error.
var ErrDraining = errors.New("draining")

// Pinger reports whether a dependency, such as the database, can serve requests. *pgxpool.Pool satisfies it.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Readiness is GET /readyz. It answers 503 while the dependency does not respond, so that no traffic is routed here.
func Readiness(dep Pinger) httpx.Route {
	return httpx.Route{
		Pattern: "GET /readyz",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
			defer cancel()

			if err := dep.Ping(ctx); err != nil {
				// The cause is only logged, like unexpected errors elsewhere.
				if !errors.Is(err, ErrDraining) {
					slog.ErrorContext(r.Context(), "readiness check failed", "error", err)
				}

				httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})

				return
			}

			httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		}),
	}
}
