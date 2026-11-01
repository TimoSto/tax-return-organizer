package httpadapter

import (
	"context"
	"net/http"
	"strings"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// SessionResolver resolves a bearer token to the actor it belongs to. ok is false for an unknown or expired token.
// It stands in for the inbound port inbound.ResolveSession of docs/AUTHENTICATION.md until there are login and
// server-side sessions; then that use case takes its place.
//
// TODO(auth): delete once inbound.ResolveSession exists; Deps.Sessions and authenticate take that port instead.
type SessionResolver interface {
	ResolveSession(ctx context.Context, token string) (actor domain.Actor, ok bool, err error)
}

// RejectAllSessions resolves no token, so every route that needs a session answers 401. It is the resolver of the
// production binary until there is login.
//
// TODO(auth): delete with SessionResolver; production then resolves through the ResolveSession use case.
type RejectAllSessions struct{}

// ResolveSession implements SessionResolver.
func (RejectAllSessions) ResolveSession(context.Context, string) (domain.Actor, bool, error) {
	return domain.Actor{}, false, nil
}

// authenticate resolves the bearer token of the request and puts its actor on the context for httpx.Actor. A
// missing, malformed or unknown token is answered with 401.
func authenticate(sessions SessionResolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			httpx.WriteUnauthenticated(w, r)

			return
		}

		actor, ok, err := sessions.ResolveSession(r.Context(), token)
		if err != nil {
			httpx.WriteError(w, r, err)

			return
		}

		if !ok {
			httpx.WriteUnauthenticated(w, r)

			return
		}

		next.ServeHTTP(w, r.WithContext(httpx.WithActor(r.Context(), actor)))
	})
}

// bearerToken returns the token of an "Authorization: Bearer <token>" header. The scheme is case-insensitive
// (RFC 9110).
func bearerToken(r *http.Request) (string, bool) {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}

	return token, true
}
