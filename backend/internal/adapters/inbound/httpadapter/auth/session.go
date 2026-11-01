// Package auth contains the HTTP handlers of the session resource.
package auth

import (
	"fmt"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type sessionResponse struct {
	Role          string `json:"role"`
	UserID        string `json:"userId"`
	TaxPracticeID string `json:"taxPracticeId"`
}

// Session is GET /auth/session, the actor of the caller's session. The BFF reads it to pick the area (staff or
// client) to show. It needs no use case: the authenticate middleware has resolved the actor already.
//
// TODO(auth): add the tax practice's slug and name (docs/AUTHENTICATION.md) once tax practices have a slug.
func Session() httpx.Route {
	return httpx.Route{
		Pattern: "GET /auth/session",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := httpx.Actor(w, r)
			if !ok {
				return
			}

			role, err := roleName(actor.Role())
			if err != nil {
				httpx.WriteError(w, r, err)

				return
			}

			httpx.WriteJSON(w, http.StatusOK, sessionResponse{
				Role:          role,
				UserID:        actor.UserID().String(),
				TaxPracticeID: actor.TaxPracticeID().String(),
			})
		}),
	}
}

func roleName(r domain.Role) (string, error) {
	switch r {
	case domain.RoleStaff:
		return "staff", nil
	case domain.RoleClient:
		return "client", nil
	default:
		return "", fmt.Errorf("actor has unknown role %d", r)
	}
}
