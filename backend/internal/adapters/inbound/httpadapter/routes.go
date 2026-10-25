// Package httpadapter exposes the use cases over HTTP.
package httpadapter

import (
	"fmt"
	"net/http"
	"slices"
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
// A new field of Deps needs a line here.
func (d Deps) validate() error {
	var missing []string

	for name, isNil := range map[string]bool{
		"CreateCollector": d.CreateCollector == nil,
		"GetCollector":    d.GetCollector == nil,
		"GetPersona":      d.GetPersona == nil,
		"ListPersonas":    d.ListPersonas == nil,
		"Readiness":       d.Readiness == nil,
	} {
		if isNil {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		slices.Sort(missing)
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

// NewHandler registers all endpoints and wraps them in the middleware. It fails if a dependency is nil or a
// route is misconfigured.
func NewHandler(d Deps) (http.Handler, error) {
	if err := d.validate(); err != nil {
		return nil, err
	}

	routes := Routes(d)
	if err := checkPathTenants(routes); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	for _, r := range routes {
		mux.Handle(r.Pattern, withTenant(r, pathTenant))
	}

	return Middleware(mux), nil
}

// tenantResolver determines the collector a request acts for.
type tenantResolver func(w http.ResponseWriter, r *http.Request) (domain.CollectorID, bool)

// withTenant puts the resolved collector into the context of routes marked Tenant. Once there is authentication,
// only the resolver changes: the tenant then comes from the authenticated principal (and checkPathTenants goes).
func withTenant(r httpx.Route, resolve tenantResolver) http.Handler {
	if !r.Tenant {
		return r.Handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, ok := resolve(w, req)
		if !ok {
			return
		}

		r.Handler.ServeHTTP(w, req.WithContext(tenant.With(req.Context(), id)))
	})
}

// pathTenant reads the tenant from the {collectorId} path value.
func pathTenant(w http.ResponseWriter, r *http.Request) (domain.CollectorID, bool) {
	return httpx.ParsePathID(w, r, httpx.CollectorIDParam, domain.ParseCollectorID)
}

// checkPathTenants fails at startup when a route's Tenant flag and its pattern disagree, which pathTenant
// would otherwise turn into a 400 or a missing tenant at request time.
func checkPathTenants(routes []httpx.Route) error {
	var bad []string

	for _, r := range routes {
		if r.Tenant != strings.Contains(r.Pattern, "{"+httpx.CollectorIDParam+"}") {
			bad = append(bad, r.Pattern)
		}
	}

	if len(bad) > 0 {
		return fmt.Errorf("routes whose Tenant flag disagrees with {%s} in the pattern: %s", httpx.CollectorIDParam, strings.Join(bad, ", "))
	}

	return nil
}
