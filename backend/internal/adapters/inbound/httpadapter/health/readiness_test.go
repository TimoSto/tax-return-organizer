package health_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/health"
)

type pinger func(context.Context) error

func (p pinger) Ping(ctx context.Context) error { return p(ctx) }

func TestReadiness(t *testing.T) {
	for _, tc := range []struct {
		name     string
		ping     error
		wantCode int
		wantBody string
	}{
		{"database reachable", nil, http.StatusOK, `{"status":"ok"}`},
		{"database down", errors.New("connection refused"), http.StatusServiceUnavailable, `{"status":"unavailable"}`},
		{"draining", health.ErrDraining, http.StatusServiceUnavailable, `{"status":"unavailable"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			route := health.Readiness(pinger(func(context.Context) error { return tc.ping }))
			mux := http.NewServeMux()
			mux.Handle(route.Pattern, route.Handler)

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantCode)
			}

			if got := strings.TrimSpace(rec.Body.String()); got != tc.wantBody {
				t.Errorf("body = %s, want %s", got, tc.wantBody)
			}
		})
	}
}

func TestReadiness_PingIsBounded(t *testing.T) {
	route := health.Readiness(pinger(func(ctx context.Context) error {
		if _, ok := ctx.Deadline(); !ok {
			t.Error("ping context has no deadline")
		}

		return nil
	}))

	route.Handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/readyz", nil))
}
