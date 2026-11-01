package taxpractice_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/taxpractice"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

var testID = domaintest.TaxPracticeID()

type createFunc func(context.Context, inbound.CreateTaxPracticeCommand) (domain.TaxPracticeID, error)

func (f createFunc) Execute(ctx context.Context, cmd inbound.CreateTaxPracticeCommand) (domain.TaxPracticeID, error) {
	return f(ctx, cmd)
}

func do(t *testing.T, create createFunc, method, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/tax-practices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := taxpractice.Create(create)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestCreateTaxPractice_Created(t *testing.T) {
	var got inbound.CreateTaxPracticeCommand

	rec := do(t, func(_ context.Context, cmd inbound.CreateTaxPracticeCommand) (domain.TaxPracticeID, error) {
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
}

func TestCreateTaxPractice_Errors(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		portErr error
		want    int
		code    string
	}{
		{"malformed json", `{`, nil, http.StatusBadRequest, "invalid_request_body"},
		{"unknown field", `{"name":"a","email":"a@b.de","x":1}`, nil, http.StatusBadRequest, "invalid_request_body"},
		{"name required", `{}`, domain.ErrNameRequired, http.StatusBadRequest, "tax_practice_name_required"},
		{"invalid email", `{}`, errors.Join(domain.ErrInvalidEmail, errors.New("x")), http.StatusBadRequest, "invalid_email"},
		{"internal", `{}`, errors.New("db down"), http.StatusInternalServerError, "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, func(context.Context, inbound.CreateTaxPracticeCommand) (domain.TaxPracticeID, error) {
				return domain.TaxPracticeID{}, tt.portErr
			}, http.MethodPost, tt.body)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}

			var p httpx.Problem
			if err := json.NewDecoder(rec.Body).Decode(&p); err != nil || p.Code != tt.code {
				t.Errorf("code = %q (%v), want %q", p.Code, err, tt.code)
			}
		})
	}
}

func TestCreateTaxPractice_MethodNotAllowed(t *testing.T) {
	rec := do(t, nil, http.MethodGet, "")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
