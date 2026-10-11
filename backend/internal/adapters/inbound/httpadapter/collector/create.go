// Package collector contains the HTTP handlers of the collector resource.
package collector

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

type createRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type createResponse struct {
	ID string `json:"id"`
}

// Create is POST /collectors.
func Create(log *slog.Logger, create ports.CreateCollector) httpx.Route {
	return httpx.Route{
		Pattern: "POST /collectors",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req createRequest

			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, httpx.MaxBodyBytes))
			dec.DisallowUnknownFields()

			if err := dec.Decode(&req); err != nil {
				httpx.WriteJSON(log, w, http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid request body"})

				return
			}

			id, err := create.Execute(r.Context(), ports.CreateCollectorCmd{Name: req.Name, Email: req.Email})
			if err != nil {
				httpx.WriteError(log, w, r, err)

				return
			}

			w.Header().Set("Location", "/collectors/"+id.String())
			httpx.WriteJSON(log, w, http.StatusCreated, createResponse{ID: id.String()})
		}),
	}
}
