// Package httpadapter exposes the inbound ports over HTTP.
package httpadapter

import (
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/health"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

// Deps are the inbound ports the endpoints call.
type Deps struct {
	CreateCollector ports.CreateCollectorPort
	GetCollector    ports.GetCollectorPort
	GetPersona      ports.GetPersonaPort
}

// NewHandler registers all endpoints and wraps them in the middleware.
func NewHandler(d Deps) http.Handler {
	mux := http.NewServeMux()

	for _, r := range []httpx.Route{
		health.Liveness(),
		collector.Create(d.CreateCollector),
		collector.Get(d.GetCollector),
		persona.Get(d.GetPersona),
	} {
		mux.Handle(r.Pattern, r.Handler)
	}

	return Middleware(mux)
}
