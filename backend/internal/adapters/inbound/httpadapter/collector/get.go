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

// Getter retrieves the collector the request acts for (see tenant.WithCollector).
type Getter interface {
	Execute(ctx context.Context) (*domain.Collector, error)
}

// Get is GET /collectors/{collectorId}.
func Get(getter Getter) httpx.Route {
	return httpx.Route{
		Pattern: "GET /collectors/{collectorId}",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := getter.Execute(r.Context())
			if err != nil {
				httpx.WriteError(w, r, err)

				return
			}

			httpx.WriteJSON(w, http.StatusOK, getResponse{ID: c.ID.String(), Name: c.Name, Email: c.Email})
		}),
	}
}
