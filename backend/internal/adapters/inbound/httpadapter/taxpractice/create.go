// Package taxpractice contains the HTTP handlers of the tax practice resource.
package taxpractice

import (
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
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

// Create is POST /tax-practices.
func Create(creator inbound.CreateTaxPractice) httpx.Route {
	return httpx.Route{
		Pattern: "POST /tax-practices",
		// Public until the operator creates tax practices (docs/AUTHENTICATION.md).
		Public: true,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req createRequest

			if !httpx.DecodeJSON(w, r, &req) {
				return
			}

			id, err := creator.Execute(r.Context(), inbound.CreateTaxPracticeCommand{Name: req.Name, Email: req.Email, Locale: req.Locale})
			if err != nil {
				httpx.WriteError(w, r, err)

				return
			}

			httpx.WriteJSON(w, http.StatusCreated, createResponse{ID: id.String()})
		}),
	}
}
