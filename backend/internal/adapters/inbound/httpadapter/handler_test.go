package httpadapter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
)

func do(t *testing.T, create func(context.Context, string, string) (string, error), method, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/collectors", strings.NewReader(body))
	rec := httptest.NewRecorder()
	httpadapter.NewHandler(create).ServeHTTP(rec, req)

	return rec
}

func TestCreateCollector_Created(t *testing.T) {
	var gotName, gotEmail string

	rec := do(t, func(_ context.Context, name, email string) (string, error) {
		gotName, gotEmail = name, email

		return "abc", nil
	}, http.MethodPost, `{"name":"Jane","email":"jane@example.com"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}

	if gotName != "Jane" || gotEmail != "jane@example.com" {
		t.Errorf("port got %q, %q", gotName, gotEmail)
	}

	if got := strings.TrimSpace(rec.Body.String()); got != `{"id":"abc"}` {
		t.Errorf("body = %s", got)
	}

	if loc := rec.Header().Get("Location"); loc != "/collectors/abc" {
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
		{"name required", `{}`, createcollector.ErrNameRequired, http.StatusBadRequest},
		{"invalid email", `{}`, errors.Join(domain.ErrInvalidEmail, errors.New("x")), http.StatusBadRequest},
		{"internal", `{}`, errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, func(context.Context, string, string) (string, error) {
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
