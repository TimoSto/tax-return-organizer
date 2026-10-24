package httpadapter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
)

func TestNewHandler(t *testing.T) {
	h := httpadapter.NewHandler(httpadapter.Deps{})

	for _, tt := range []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/healthz", http.StatusOK},
		{http.MethodGet, "/collectors", http.StatusMethodNotAllowed},
		{http.MethodGet, "/nope", http.StatusNotFound},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

		if rec.Code != tt.want {
			t.Errorf("%s %s = %d, want %d", tt.method, tt.path, rec.Code, tt.want)
		}

		if rec.Header().Get("X-Request-ID") == "" {
			t.Errorf("%s %s: missing X-Request-ID", tt.method, tt.path)
		}
	}
}
