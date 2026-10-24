package httpadapter_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
	"github.com/TimoSto/tax-return-organizer/backend/internal/logging"
)

// captureLogs makes the default logger write to logs until the test ends. Tests using it must not run in parallel.
func captureLogs(t *testing.T, logs *bytes.Buffer) {
	t.Helper()

	prev := slog.Default()

	slog.SetDefault(slog.New(logging.NewHandler(slog.NewTextHandler(logs, nil))))
	t.Cleanup(func() { slog.SetDefault(prev) })
}

func TestMiddleware_RecoversPanic(t *testing.T) {
	var logs bytes.Buffer

	captureLogs(t, &logs)

	h := httpadapter.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))

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

	captureLogs(t, &logs)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tea/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })

	httpadapter.Middleware(mux).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/tea/42", nil))

	for _, want := range []string{"method=GET", `route="GET /tea/{id}"`, "status=418", "request_id="} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("logs missing %q: %s", want, logs.String())
		}
	}

	if strings.Contains(logs.String(), "/tea/42") {
		t.Errorf("logs contain the raw path: %s", logs.String())
	}
}

func TestMiddleware_LogsUnmatchedRoute(t *testing.T) {
	var logs bytes.Buffer

	captureLogs(t, &logs)

	httpadapter.Middleware(http.NewServeMux()).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/nope", nil))

	if !strings.Contains(logs.String(), "route=unmatched") {
		t.Errorf("logs missing route=unmatched: %s", logs.String())
	}
}

func TestMiddleware_HonoursIncomingRequestID(t *testing.T) {
	var logs bytes.Buffer

	captureLogs(t, &logs)

	var inHandler string

	h := httpadapter.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		inHandler, _ = logging.RequestIDFrom(r.Context())
		slog.InfoContext(r.Context(), "from handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Request-ID", "client-id.1")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "client-id.1" {
		t.Errorf("response id = %q, want client-id.1", got)
	}

	if inHandler != "client-id.1" {
		t.Errorf("context id = %q, want client-id.1", inHandler)
	}

	if n := strings.Count(logs.String(), "request_id=client-id.1"); n != 2 {
		t.Errorf("request id logged %d times, want 2: %s", n, logs.String())
	}
}

func TestMiddleware_ReplacesInvalidRequestID(t *testing.T) {
	h := httpadapter.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	for _, bad := range []string{"a b", "line\nbreak", strings.Repeat("a", 65)} {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header["X-Request-Id"] = []string{bad}

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if got := rec.Header().Get("X-Request-ID"); got == bad || got == "" {
			t.Errorf("id %q was not replaced, got %q", bad, got)
		}
	}
}
