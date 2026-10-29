package collector

import (
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
)

type getResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Get is GET /collectors/{collectorId}.
func Get(getter inbound.GetCollector) httpx.Route {
	return httpx.Route{
		Pattern: "GET /collectors/{collectorId}",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := httpx.Actor(w, r)
			if !ok {
				return
			}

			c, err := getter.Execute(r.Context(), actor)
			if err != nil {
				httpx.WriteError(w, r, err)

				return
			}

			httpx.WriteJSON(w, http.StatusOK, getResponse{ID: c.ID().String(), Name: c.Name(), Email: c.Email().String()})
		}),
	}
}
