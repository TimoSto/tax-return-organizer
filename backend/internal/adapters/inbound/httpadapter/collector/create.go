// Package collector contains the HTTP handlers of the collector resource.
package collector

import (
	"context"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
)

type createRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	// Locale is optional.
	Locale string `json:"locale"`
}

type createResponse struct {
	ID string `json:"id"`
}

// Creator creates a collector and returns its id.
type Creator interface {
	Execute(ctx context.Context, cmd createcollector.Command) (domain.CollectorID, error)
}

// Create is POST /collectors.
func Create(creator Creator) httpx.Route {
	return httpx.Route{
		Pattern: "POST /collectors",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req createRequest

			if !httpx.DecodeJSON(w, r, &req) {
				return
			}

			id, err := creator.Execute(r.Context(), createcollector.Command{Name: req.Name, Email: req.Email, Locale: req.Locale})
			if err != nil {
				httpx.WriteError(w, r, err)

				return
			}

			w.Header().Set("Location", "/collectors/"+id.String())
			httpx.WriteJSON(w, http.StatusCreated, createResponse{ID: id.String()})
		}),
	}
}
