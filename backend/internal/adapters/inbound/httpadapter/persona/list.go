package persona

import (
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

type listResponse struct {
	Personas []getResponse `json:"personas"`
}

// List is GET /collectors/{collectorId}/personas.
func List(list ports.ListPersonasPort) httpx.Route {
	return httpx.Route{
		Pattern: "GET /collectors/{collectorId}/personas",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			collectorID, err := domain.ParseCollectorID(r.PathValue("collectorId"))
			if err != nil {
				httpx.WriteJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid collector id"})

				return
			}

			personas, err := list(r.Context(), collectorID)
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
