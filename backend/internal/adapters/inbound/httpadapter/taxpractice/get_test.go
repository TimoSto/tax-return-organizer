package taxpractice_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/taxpractice"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

type getFunc func(context.Context, domain.Actor) (*domain.TaxPractice, error)

func (f getFunc) Execute(ctx context.Context, actor domain.Actor) (*domain.TaxPractice, error) {
	return f(ctx, actor)
}

func doGet(t *testing.T, get getFunc, method string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/tax-practice", nil)
	req = req.WithContext(httpx.WithActor(req.Context(), domaintest.Staff(testID)))
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := taxpractice.Get(get)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestGetTaxPractice_OK(t *testing.T) {
	var gotID domain.TaxPracticeID

	rec := doGet(t, func(_ context.Context, actor domain.Actor) (*domain.TaxPractice, error) {
		gotID = actor.TaxPracticeID()

		return domaintest.TaxPractice(testID, "Jane", "jane@example.com"), nil
	}, http.MethodGet)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if gotID != testID {
		t.Errorf("use case got tax practice %v, want %v", gotID, testID)
	}

	want := `{"id":"` + testID.String() + `","name":"Jane","email":"jane@example.com"}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestGetTaxPractice_Errors(t *testing.T) {
	tests := []struct {
		name    string
		portErr error
		want    int
	}{
		{"not found", domain.ErrTaxPracticeNotFound, http.StatusNotFound},
		{"internal", errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doGet(t, func(context.Context, domain.Actor) (*domain.TaxPractice, error) {
				return nil, tt.portErr
			}, http.MethodGet)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestGetTaxPractice_MethodNotAllowed(t *testing.T) {
	rec := doGet(t, nil, http.MethodPost)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
