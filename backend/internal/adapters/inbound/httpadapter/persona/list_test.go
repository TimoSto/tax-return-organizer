package persona_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/tenant"
)

type listFunc func(context.Context) ([]*domain.Persona, error)

func (f listFunc) Execute(ctx context.Context) ([]*domain.Persona, error) {
	return f(ctx)
}

func doList(t *testing.T, list listFunc, method string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/collectors/"+testCollectorID.String()+"/personas", nil)
	req = req.WithContext(tenant.WithCollector(req.Context(), testCollectorID))
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := persona.List(list)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestListPersonas_OK(t *testing.T) {
	rec := doList(t, func(context.Context) ([]*domain.Persona, error) {
		return []*domain.Persona{domaintest.Persona(testID, testCollectorID, "Employed", "Has a job")}, nil
	}, http.MethodGet)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	want := `{"personas":[{"id":"` + testID.String() + `","collectorId":"` + testCollectorID.String() +
		`","name":"Employed","description":"Has a job","items":[]}]}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestListPersonas_EmptyIsNotNull(t *testing.T) {
	rec := doList(t, func(context.Context) ([]*domain.Persona, error) {
		return nil, nil
	}, http.MethodGet)

	if got, want := strings.TrimSpace(rec.Body.String()), `{"personas":[]}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestListPersonas_Errors(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{"unknown collector", domain.ErrCollectorNotFound, http.StatusNotFound},
		{"other error", errors.New("boom"), http.StatusInternalServerError},
	} {
		rec := doList(t, func(context.Context) ([]*domain.Persona, error) {
			return nil, tc.err
		}, http.MethodGet)

		if rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}

func TestListPersonas_MethodNotAllowed(t *testing.T) {
	rec := doList(t, nil, http.MethodPost)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
