package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/auth"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

var testID = domaintest.TaxPracticeID()

// do sends a request with the given actor on the context; withActor false sends none.
func do(t *testing.T, actor domain.Actor, withActor bool, method string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/auth/session", nil)
	if withActor {
		req = req.WithContext(httpx.WithActor(context.Background(), actor))
	}

	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := auth.Session()
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestSession_OK(t *testing.T) {
	tests := []struct {
		name  string
		actor domain.Actor
		role  string
	}{
		{"staff", domaintest.Staff(testID), "staff"},
		{"client", domaintest.Client(testID), "client"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, tt.actor, true, http.MethodGet)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}

			want := `{"role":"` + tt.role + `","userId":"` + tt.actor.UserID().String() +
				`","taxPracticeId":"` + testID.String() + `"}`
			if got := strings.TrimSpace(rec.Body.String()); got != want {
				t.Errorf("body = %s, want %s", got, want)
			}
		})
	}
}

func TestSession_Errors(t *testing.T) {
	tests := []struct {
		name      string
		actor     domain.Actor
		withActor bool
		want      int
	}{
		{"no actor", domain.Actor{}, false, http.StatusUnauthorized},
		{"actor without role", domain.Actor{}, true, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, tt.actor, tt.withActor, http.MethodGet)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestSession_MethodNotAllowed(t *testing.T) {
	rec := do(t, domaintest.Staff(testID), true, http.MethodPost)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
