package collector

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

type getResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Get is GET /collectors/{id}.
func Get(log *slog.Logger, get ports.GetCollectorPort) httpx.Route {
	return httpx.Route{
		Pattern: "GET /collectors/{id}",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(r.PathValue("id"))
			if err != nil {
				httpx.WriteJSON(log, w, http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid collector id"})

				return
			}

			c, err := get(r.Context(), id)
			if err != nil {
				httpx.WriteError(log, w, r, err)

				return
			}

			httpx.WriteJSON(log, w, http.StatusOK, getResponse{ID: c.ID.String(), Name: c.Name, Email: c.Email})
		}),
	}
}
