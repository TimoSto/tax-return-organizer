// Package httpadapter exposes the inbound ports over HTTP.
package httpadapter

import (
	"log/slog"
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
func NewHandler(log *slog.Logger, d Deps) http.Handler {
	mux := http.NewServeMux()

	for _, r := range []httpx.Route{
		health.Liveness(log),
		collector.Create(log, d.CreateCollector),
		collector.Get(log, d.GetCollector),
		persona.Get(log, d.GetPersona),
	} {
		mux.Handle(r.Pattern, r.Handler)
	}

	return Middleware(log, mux)
}
