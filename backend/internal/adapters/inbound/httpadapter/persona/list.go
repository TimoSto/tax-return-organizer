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

// Lister retrieves all personas of a collector.
type Lister interface {
	Execute(ctx context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error)
}

// List is GET /collectors/{collectorId}/personas.
func List(lister Lister) httpx.Route {
	return httpx.Route{
		Pattern: "GET /collectors/{collectorId}/personas",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			collectorID, err := domain.ParseCollectorID(r.PathValue("collectorId"))
			if err != nil {
				httpx.WriteJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid collector id"})

				return
			}

			personas, err := lister.Execute(r.Context(), collectorID)
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
