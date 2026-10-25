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

	mux := http.NewServeMux()

	for _, r := range Routes(d) {
		mux.Handle(r.Pattern, r.Handler)
	}

	return Middleware(mux), nil
}
