package backend

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake"
)

func TestWire_CreateCollector(t *testing.T) {
	h := wire(slog.New(slog.DiscardHandler), fake.NewCollectorRepository())

	req := httptest.NewRequest(http.MethodPost, "/collectors", strings.NewReader(`{"name":"Acme","email":"a@example.com"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestWire_Healthz(t *testing.T) {
	h := wire(slog.New(slog.DiscardHandler), fake.NewCollectorRepository())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("missing X-Request-ID header")
	}
}
