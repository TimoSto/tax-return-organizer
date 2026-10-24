// Package httpadapter exposes the use cases over HTTP.
package httpadapter

import (
	"net/http"
	"strings"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/health"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/tenant"
)

// Deps are the use cases the endpoints call.
type Deps struct {
	CreateCollector collector.Creator
	GetCollector    collector.Getter
	GetPersona      persona.Getter
	ListPersonas    persona.Lister
}

// NewHandler registers all endpoints and wraps them in the middleware.
func NewHandler(d Deps) http.Handler {
	mux := http.NewServeMux()

	for _, r := range []httpx.Route{
		health.Liveness(),
		collector.Create(d.CreateCollector),
		collector.Get(d.GetCollector),
		persona.Get(d.GetPersona),
		persona.List(d.ListPersonas),
	} {
		mux.Handle(r.Pattern, tenantFromPath(r))
	}

	return Middleware(mux)
}

const collectorIDParam = "collectorId"

// tenantFromPath puts the {collectorId} path value of routes that have one into the context. Once there is
// authentication, this is the only place that changes: the tenant then comes from the authenticated principal.
func tenantFromPath(r httpx.Route) http.Handler {
	if !strings.Contains(r.Pattern, "{"+collectorIDParam+"}") {
		return r.Handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, err := domain.ParseCollectorID(req.PathValue(collectorIDParam))
		if err != nil {
			httpx.WriteJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid collector id"})

			return
		}

		r.Handler.ServeHTTP(w, req.WithContext(tenant.WithCollector(req.Context(), id)))
	})
}
