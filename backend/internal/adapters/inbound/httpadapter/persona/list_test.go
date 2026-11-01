package persona_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

type listFunc func(context.Context, domain.Actor) ([]*domain.Persona, error)

func (f listFunc) Execute(ctx context.Context, actor domain.Actor) ([]*domain.Persona, error) {
	return f(ctx, actor)
}

func doList(t *testing.T, list listFunc, method string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/personas", nil)
	req = req.WithContext(httpx.WithActor(req.Context(), domaintest.Staff(testTaxPracticeID)))
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := persona.List(list)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestListPersonas_OK(t *testing.T) {
	var gotTaxPractice domain.TaxPracticeID

	rec := doList(t, func(_ context.Context, actor domain.Actor) ([]*domain.Persona, error) {
		gotTaxPractice = actor.TaxPracticeID()

		return []*domain.Persona{domaintest.Persona(testID, testTaxPracticeID, "Employed", "Has a job")}, nil
	}, http.MethodGet)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if gotTaxPractice != testTaxPracticeID {
		t.Errorf("use case got tax practice %v, want %v", gotTaxPractice, testTaxPracticeID)
	}

	want := `{"personas":[{"id":"` + testID.String() + `","taxPracticeId":"` + testTaxPracticeID.String() +
		`","name":"Employed","description":"Has a job","items":[]}]}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestListPersonas_EmptyIsNotNull(t *testing.T) {
	rec := doList(t, func(context.Context, domain.Actor) ([]*domain.Persona, error) {
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
		{"unknown tax practice", domain.ErrTaxPracticeNotFound, http.StatusNotFound},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden},
		{"other error", errors.New("boom"), http.StatusInternalServerError},
	} {
		rec := doList(t, func(context.Context, domain.Actor) ([]*domain.Persona, error) {
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
