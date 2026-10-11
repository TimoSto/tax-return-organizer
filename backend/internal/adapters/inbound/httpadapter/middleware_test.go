package httpadapter_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
)

func TestMiddleware_RecoversPanic(t *testing.T) {
	var logs bytes.Buffer

	log := slog.New(slog.NewTextHandler(&logs, nil))
	h := httpadapter.Middleware(log, http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("missing X-Request-ID")
	}

	if !strings.Contains(logs.String(), "panic") || !strings.Contains(logs.String(), "status=500") {
		t.Errorf("logs missing panic or request entry: %s", logs.String())
	}
}

func TestMiddleware_LogsRequest(t *testing.T) {
	var logs bytes.Buffer

	log := slog.New(slog.NewTextHandler(&logs, nil))
	h := httpadapter.Middleware(log, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/tea", nil))

	for _, want := range []string{"method=GET", "path=/tea", "status=418"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("logs missing %q: %s", want, logs.String())
		}
	}
}
