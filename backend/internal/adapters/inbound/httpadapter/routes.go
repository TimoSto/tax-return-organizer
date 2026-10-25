// Package httpadapter exposes the use cases over HTTP.
package httpadapter

import (
	"fmt"
	"net/http"
	"reflect"
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
	// Readiness is checked by GET /readyz, normally the database pool.
	Readiness health.Pinger
}

// validate reports the fields of d that are nil, so a missing wiring fails at startup and not on the first request.
func (d Deps) validate() error {
	v := reflect.ValueOf(d)

	var missing []string

	for i := range v.NumField() {
		if v.Field(i).IsNil() {
			missing = append(missing, v.Type().Field(i).Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing dependencies: %s", strings.Join(missing, ", "))
	}

	return nil
}

// Routes lists all endpoints. A new endpoint is added here and described in api/openapi.yaml.
func Routes(d Deps) []httpx.Route {
	return []httpx.Route{
		health.Liveness(),
		health.Readiness(d.Readiness),
		collector.Create(d.CreateCollector),
		collector.Get(d.GetCollector),
		persona.Get(d.GetPersona),
		persona.List(d.ListPersonas),
	}
}

// NewHandler registers all endpoints and wraps them in the middleware. It fails if a dependency is nil.
func NewHandler(d Deps) (http.Handler, error) {
	if err := d.validate(); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	for _, r := range Routes(d) {
		mux.Handle(r.Pattern, tenantFromPath(r))
	}

	return Middleware(mux), nil
}

const collectorIDParam = "collectorId"

// tenantFromPath puts the {collectorId} path value of routes that have one into the context. Once there is
// authentication, this is the only place that changes: the tenant then comes from the authenticated principal.
func tenantFromPath(r httpx.Route) http.Handler {
	if !strings.Contains(r.Pattern, "{"+collectorIDParam+"}") {
		return r.Handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, ok := httpx.ParsePathID(w, req, collectorIDParam, domain.ParseCollectorID)
		if !ok {
			return
		}

		r.Handler.ServeHTTP(w, req.WithContext(tenant.WithCollector(req.Context(), id)))
	})
}
