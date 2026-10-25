package persona

import (
	"context"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type listResponse struct {
	Personas []getResponse `json:"personas"`
}

// Lister retrieves all personas of the collector the request acts for.
type Lister interface {
	Execute(ctx context.Context) ([]*domain.Persona, error)
}

// List is GET /collectors/{collectorId}/personas.
func List(lister Lister) httpx.Route {
	return httpx.Route{
		Pattern: "GET /collectors/{collectorId}/personas",
		Tenant:  true,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			personas, err := lister.Execute(r.Context())
			if err != nil {
				httpx.WriteError(w, r, err)

				return
			}

			// Never null, so clients can iterate without a check.
			resp := listResponse{Personas: make([]getResponse, 0, len(personas))}
			for _, p := range personas {
				resp.Personas = append(resp.Personas, newGetResponse(p))
			}

			httpx.WriteJSON(w, http.StatusOK, resp)
		}),
	}
}
