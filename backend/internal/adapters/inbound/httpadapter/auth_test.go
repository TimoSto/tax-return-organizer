package httpadapter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

const validToken = "valid-token"

// sessionStub resolves validToken to actor and fails every lookup with err if it is set.
type sessionStub struct {
	actor domain.Actor
	err   error
}

func (s sessionStub) ResolveSession(_ context.Context, token string) (domain.Actor, bool, error) {
	if s.err != nil {
		return domain.Actor{}, false, s.err
	}

	return s.actor, token == validToken, nil
}

// getCollectorFunc records the actor the handler passed, to check what the middleware resolved.
type getCollectorFunc func(context.Context, domain.Actor) (*domain.Collector, error)

func (f getCollectorFunc) Execute(ctx context.Context, actor domain.Actor) (*domain.Collector, error) {
	return f(ctx, actor)
}

func authHandler(t *testing.T, sessions httpadapter.SessionResolver, got *domain.Actor) http.Handler {
	t.Helper()

	d := fullDeps()
	d.Sessions = sessions
	d.GetCollector = getCollectorFunc(func(_ context.Context, a domain.Actor) (*domain.Collector, error) {
		*got = a

		return domaintest.Collector(a.CollectorID(), "Jane", "jane@example.com"), nil
	})

	h, err := httpadapter.NewHandler(d)
	if err != nil {
		t.Fatal(err)
	}

	return h
}

func TestAuthenticate_ValidToken(t *testing.T) {
	want := domain.NewStaffActor(domaintest.CollectorID(), domain.UserID{})

	var got domain.Actor

	h := authHandler(t, sessionStub{actor: want}, &got)

	for _, header := range []string{"Bearer " + validToken, "bearer " + validToken} {
		req := httptest.NewRequest(http.MethodGet, "/collector", nil)
		req.Header.Set("Authorization", header)

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || got != want {
			t.Errorf("%q: status = %d, actor = %+v, want 200 and %+v", header, rec.Code, got, want)
		}
	}
}

func TestAuthenticate_Rejects(t *testing.T) {
	var got domain.Actor

	h := authHandler(t, sessionStub{actor: domain.NewStaffActor(domaintest.CollectorID(), domain.UserID{})}, &got)

	for _, header := range []string{"", "Bearer", "Bearer ", "Basic " + validToken, validToken, "Bearer unknown"} {
		req := httptest.NewRequest(http.MethodGet, "/collector", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		var p httpx.Problem
		_ = json.NewDecoder(rec.Body).Decode(&p)

		if rec.Code != http.StatusUnauthorized || p.Code != httpx.CodeUnauthenticated {
			t.Errorf("%q: status = %d, code = %q, want 401 unauthenticated", header, rec.Code, p.Code)
		}

		if rec.Header().Get("WWW-Authenticate") != "Bearer" {
			t.Errorf("%q: WWW-Authenticate = %q, want Bearer", header, rec.Header().Get("WWW-Authenticate"))
		}
	}

	if got != (domain.Actor{}) {
		t.Errorf("handler ran with actor %+v", got)
	}
}

func TestAuthenticate_ResolverError(t *testing.T) {
	var got domain.Actor

	h := authHandler(t, sessionStub{err: errors.New("db down")}, &got)

	req := httptest.NewRequest(http.MethodGet, "/collector", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestAuthenticate_PublicRoutesNeedNoSession(t *testing.T) {
	h := newHandler(t)

	for _, path := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, rec.Code)
		}
	}
}
