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
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

const testID domain.CollectorID = "k7m2"

func do(t *testing.T, create ports.CreateCollectorPort, method, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/collectors", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := collector.Create(create)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestCreateCollector_Created(t *testing.T) {
	var got ports.CreateCollectorCmd

	rec := do(t, func(_ context.Context, cmd ports.CreateCollectorCmd) (domain.CollectorID, error) {
		got = cmd

		return testID, nil
	}, http.MethodPost, `{"name":"Jane","email":"jane@example.com"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}

	if got.Name != "Jane" || got.Email != "jane@example.com" {
		t.Errorf("port got %+v", got)
	}

	if got := strings.TrimSpace(rec.Body.String()); got != `{"id":"`+testID.String()+`"}` {
		t.Errorf("body = %s", got)
	}

	if loc := rec.Header().Get("Location"); loc != "/collectors/"+testID.String() {
		t.Errorf("location = %q", loc)
	}
}

func TestCreateCollector_Errors(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		portErr error
		want    int
	}{
		{"malformed json", `{`, nil, http.StatusBadRequest},
		{"unknown field", `{"name":"a","email":"a@b.de","x":1}`, nil, http.StatusBadRequest},
		{"name required", `{}`, domain.ErrNameRequired, http.StatusBadRequest},
		{"invalid email", `{}`, errors.Join(domain.ErrInvalidEmail, errors.New("x")), http.StatusBadRequest},
		{"internal", `{}`, errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, func(context.Context, ports.CreateCollectorCmd) (domain.CollectorID, error) {
				return "", tt.portErr
			}, http.MethodPost, tt.body)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestCreateCollector_MethodNotAllowed(t *testing.T) {
	rec := do(t, nil, http.MethodGet, "")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
