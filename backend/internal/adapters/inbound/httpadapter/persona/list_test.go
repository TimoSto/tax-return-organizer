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
)

type listFunc func(context.Context, domain.CollectorID) ([]*domain.Persona, error)

func (f listFunc) Execute(ctx context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error) {
	return f(ctx, collectorID)
}

func doList(t *testing.T, list listFunc, method, collectorID string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/collectors/"+collectorID+"/personas", nil)
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := persona.List(list)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestListPersonas_OK(t *testing.T) {
	var gotCollector domain.CollectorID

	rec := doList(t, func(_ context.Context, collectorID domain.CollectorID) ([]*domain.Persona, error) {
		gotCollector = collectorID

		return []*domain.Persona{{ID: testID, CollectorID: testCollectorID, Name: "Employed", Description: "Has a job"}}, nil
	}, http.MethodGet, testCollectorID.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if gotCollector != testCollectorID {
		t.Errorf("port got %v, want %v", gotCollector, testCollectorID)
	}

	want := `{"personas":[{"id":"` + testID.String() + `","collectorId":"` + testCollectorID.String() +
		`","name":"Employed","description":"Has a job","items":[]}]}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestListPersonas_EmptyIsNotNull(t *testing.T) {
	rec := doList(t, func(context.Context, domain.CollectorID) ([]*domain.Persona, error) {
		return nil, nil
	}, http.MethodGet, testCollectorID.String())

	if got, want := strings.TrimSpace(rec.Body.String()), `{"personas":[]}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestListPersonas_Errors(t *testing.T) {
	for _, tc := range []struct {
		name        string
		collectorID string
		err         error
		want        int
	}{
		{"invalid collector id", "!", nil, http.StatusBadRequest},
		{"unknown collector", testCollectorID.String(), domain.ErrCollectorNotFound, http.StatusNotFound},
		{"other error", testCollectorID.String(), errors.New("boom"), http.StatusInternalServerError},
	} {
		rec := doList(t, func(context.Context, domain.CollectorID) ([]*domain.Persona, error) {
			return nil, tc.err
		}, http.MethodGet, tc.collectorID)

		if rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}

func TestListPersonas_MethodNotAllowed(t *testing.T) {
	rec := doList(t, nil, http.MethodPost, testCollectorID.String())

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
