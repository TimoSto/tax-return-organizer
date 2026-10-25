package collector_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/tenant"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

type getFunc func(context.Context) (*domain.Collector, error)

func (f getFunc) Execute(ctx context.Context) (*domain.Collector, error) {
	return f(ctx)
}

func doGet(t *testing.T, get getFunc, method string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/collectors/"+testID.String(), nil)
	req = req.WithContext(tenant.With(req.Context(), testID))
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := collector.Get(get)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestGetCollector_OK(t *testing.T) {
	rec := doGet(t, func(context.Context) (*domain.Collector, error) {
		return domaintest.Collector(testID, "Jane", "jane@example.com"), nil
	}, http.MethodGet)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	want := `{"id":"` + testID.String() + `","name":"Jane","email":"jane@example.com"}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestGetCollector_Errors(t *testing.T) {
	tests := []struct {
		name    string
		portErr error
		want    int
	}{
		{"not found", domain.ErrCollectorNotFound, http.StatusNotFound},
		{"internal", errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doGet(t, func(context.Context) (*domain.Collector, error) {
				return nil, tt.portErr
			}, http.MethodGet)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestGetCollector_MethodNotAllowed(t *testing.T) {
	rec := doGet(t, nil, http.MethodPost)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
