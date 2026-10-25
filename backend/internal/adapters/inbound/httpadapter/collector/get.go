package collector

import (
	"context"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type getResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Getter retrieves the collector the request acts for.
type Getter interface {
	Execute(ctx context.Context, id domain.CollectorID) (*domain.Collector, error)
}

// Get is GET /collectors/{collectorId}.
func Get(getter Getter) httpx.Route {
	return httpx.Route{
		Pattern: "GET /collectors/{collectorId}",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			collectorID, ok := httpx.Tenant(w, r)
			if !ok {
				return
			}

			c, err := getter.Execute(r.Context(), collectorID)
			if err != nil {
				httpx.WriteError(w, r, err)

				return
			}

			httpx.WriteJSON(w, http.StatusOK, getResponse{ID: c.ID().String(), Name: c.Name(), Email: c.Email().String()})
		}),
	}
}
