// Package httpadapter exposes the use cases over HTTP.
package httpadapter

import (
	"net/http"

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
		mux.Handle(r.Pattern, r.Handler)
	}

	return Middleware(mux)
}
