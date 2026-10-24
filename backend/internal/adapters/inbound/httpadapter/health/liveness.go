// Package health contains the HTTP handlers of the health probes.
package health

import (
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
)

// Liveness is GET /healthz. It does not touch dependencies such as the database.
func Liveness() httpx.Route {
	return httpx.Route{
		Pattern: "GET /healthz",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		}),
	}
}
