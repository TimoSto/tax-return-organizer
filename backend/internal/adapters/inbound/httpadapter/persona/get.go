// Package persona contains the HTTP handlers of the persona resource.
package persona

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type getResponse struct {
	ID          string         `json:"id"`
	CollectorID string         `json:"collectorId"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Items       []itemResponse `json:"items"`
}

// itemResponse carries exactly the spec that matches kind, the other one is null.
type itemResponse struct {
	ID           string                `json:"id"`
	Title        string                `json:"title"`
	Kind         string                `json:"kind"`
	Recurrence   string                `json:"recurrence"`
	DocumentSpec *documentSpecResponse `json:"documentSpec"`
	ValueSpec    *valueSpecResponse    `json:"valueSpec"`
}

type documentSpecResponse struct {
	MetadataFields []metadataFieldResponse `json:"metadataFields"`
}

type metadataFieldResponse struct {
	Name      string            `json:"name"`
	ValueSpec valueSpecResponse `json:"valueSpec"`
}

type valueSpecResponse struct {
	Type     string `json:"type"`
	Unit     string `json:"unit"`
	Required bool   `json:"required"`
}

// Getter retrieves a persona of a collector by its id.
type Getter interface {
	Execute(ctx context.Context, collectorID domain.CollectorID, id uuid.UUID) (*domain.Persona, error)
}

// Get is GET /collectors/{collectorId}/personas/{id}.
func Get(getter Getter) httpx.Route {
	return httpx.Route{
		Pattern: "GET /collectors/{collectorId}/personas/{id}",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			collectorID, err := domain.ParseCollectorID(r.PathValue("collectorId"))
			if err != nil {
				httpx.WriteJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid collector id"})

				return
			}

			id, err := uuid.Parse(r.PathValue("id"))
			if err != nil {
				httpx.WriteJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid persona id"})

				return
			}

			p, err := getter.Execute(r.Context(), collectorID, id)
			if err != nil {
				httpx.WriteError(w, r, err)

				return
			}

			httpx.WriteJSON(w, http.StatusOK, newGetResponse(p))
		}),
	}
}

func newGetResponse(p *domain.Persona) getResponse {
	// Never null, so clients can iterate without a check.
	items := make([]itemResponse, 0, len(p.Items))

	for _, item := range p.Items {
		items = append(items, newItemResponse(item))
	}

	return getResponse{
		ID:          p.ID.String(),
		CollectorID: p.CollectorID.String(),
		Name:        p.Name,
		Description: p.Description,
		Items:       items,
	}
}

func newItemResponse(item domain.ItemDefinition) itemResponse {
	resp := itemResponse{
		ID:         item.ID.String(),
		Title:      item.Title,
		Kind:       string(item.Kind),
		Recurrence: string(item.Recurrence),
	}

	if item.DocumentSpec != nil {
		fields := make([]metadataFieldResponse, 0, len(item.DocumentSpec.MetadataFields))
		for _, f := range item.DocumentSpec.MetadataFields {
			fields = append(fields, metadataFieldResponse{Name: f.Name, ValueSpec: newValueSpecResponse(f.ValueSpec)})
		}

		resp.DocumentSpec = &documentSpecResponse{MetadataFields: fields}
	}

	if item.ValueSpec != nil {
		spec := newValueSpecResponse(*item.ValueSpec)
		resp.ValueSpec = &spec
	}

	return resp
}

func newValueSpecResponse(s domain.ValueSpec) valueSpecResponse {
	return valueSpecResponse{Type: string(s.Type), Unit: s.Unit, Required: s.Required}
}
