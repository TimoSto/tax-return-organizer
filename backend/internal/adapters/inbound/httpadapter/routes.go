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
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
)

// Deps are the use cases the endpoints call.
type Deps struct {
	CreateCollector inbound.CreateCollector
	GetCollector    inbound.GetCollector
	GetPersona      inbound.GetPersona
	ListPersonas    inbound.ListPersonas
	// Readiness is checked by GET /readyz, normally the database pool.
	Readiness health.Pinger
}

// validate reports the fields of d that are nil, in declaration order, so a missing wiring fails at startup and not
// on the first request. It checks every field by reflection, so a new field of Deps needs no extra code here. All
// fields must be interfaces (or another nillable kind), otherwise IsNil panics at startup.
func (d Deps) validate() error {
	var missing []string

	v := reflect.ValueOf(d)
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
