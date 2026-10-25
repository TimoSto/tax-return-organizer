package httpx_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/logging"
)

func TestWriteError_Status(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
	}{
		{domain.ErrNameRequired, http.StatusBadRequest},
		{domain.ErrInvalidEmail, http.StatusBadRequest},
		{domain.ErrInvalidCollectorID, http.StatusBadRequest},
		{domain.ErrPersonaNameRequired, http.StatusBadRequest},
		{domain.ErrItemTitleRequired, http.StatusBadRequest},
		{domain.ErrInvalidRecurrence, http.StatusBadRequest},
		{domain.ErrSpecMismatch, http.StatusBadRequest},
		{domain.ErrInvalidValueType, http.StatusBadRequest},
		{domain.ErrMetadataFieldNameRequired, http.StatusBadRequest},
		{domain.ErrDuplicateMetadataField, http.StatusBadRequest},
		{domain.ErrNameTooLong, http.StatusBadRequest},
		{domain.ErrPersonaNameTooLong, http.StatusBadRequest},
		{domain.ErrPersonaDescriptionTooLong, http.StatusBadRequest},
		{domain.ErrItemTitleTooLong, http.StatusBadRequest},
		{domain.ErrMetadataFieldNameTooLong, http.StatusBadRequest},
		{domain.ErrTooManyMetadataFields, http.StatusBadRequest},
		{domain.ErrUnitTooLong, http.StatusBadRequest},
		{domain.ErrTooManyItems, http.StatusConflict},
		// Server-side faults: a duplicate id comes from our own generator, stored data from our own writes.
		{fmt.Errorf("add item: %w", domain.ErrDuplicateItemID), http.StatusInternalServerError},
		{outbound.CorruptData("collector aaaa", domain.ErrInvalidEmail), http.StatusInternalServerError},
		{fmt.Errorf("get collector: %w", outbound.CorruptData("collector aaaa", domain.ErrNameRequired)), http.StatusInternalServerError},
		{domain.ErrItemIDTaken, http.StatusConflict},
		{personatemplate.ErrInvalidLocale, http.StatusBadRequest},
		{domain.ErrCollectorIDTaken, http.StatusConflict},
		{fmt.Errorf("wrapped: %w", domain.ErrSpecMismatch), http.StatusBadRequest},
		{domain.ErrCollectorNotFound, http.StatusNotFound},
		{domain.ErrPersonaNotFound, http.StatusNotFound},
		{domain.ErrItemNotFound, http.StatusNotFound},
		{errors.New("database down"), http.StatusInternalServerError},
	} {
		t.Run(tc.err.Error(), func(t *testing.T) {
			rec := httptest.NewRecorder()

			httpx.WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), tc.err)

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestWriteError_BodyHasOnlyDomainMessage(t *testing.T) {
	for name, err := range map[string]error{
		"wrapped": fmt.Errorf("save persona: %w", domain.ErrInvalidEmail),
		"joined":  errors.Join(domain.ErrInvalidEmail, errors.New("secret detail")),
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			httpx.WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), err)

			var body httpx.Problem
			if e := json.NewDecoder(rec.Body).Decode(&body); e != nil {
				t.Fatal(e)
			}

			if body.Detail != domain.ErrInvalidEmail.Error() {
				t.Errorf("detail = %q, want %q", body.Detail, domain.ErrInvalidEmail.Error())
			}
		})
	}
}

func TestWriteError_InternalErrorIsGeneric(t *testing.T) {
	rec := httptest.NewRecorder()

	httpx.WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), errors.New("database down"))

	var body httpx.Problem
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	if body.Detail != "internal error" {
		t.Errorf("error = %q, want generic message", body.Detail)
	}
}

func TestTenant(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /collectors/{"+httpx.CollectorIDParam+"}", func(w http.ResponseWriter, r *http.Request) {
		if id, ok := httpx.Tenant(w, r); ok {
			_, _ = w.Write([]byte(id.String()))
		}
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/collectors/k7m2", nil))

	if rec.Code != http.StatusOK || rec.Body.String() != "k7m2" {
		t.Errorf("valid id: status = %d, body = %q, want 200 and k7m2", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/collectors/!", nil))

	var p httpx.Problem
	_ = json.NewDecoder(rec.Body).Decode(&p)

	if rec.Code != http.StatusBadRequest || p.Code != "invalid_collector_id" {
		t.Errorf("invalid id: status = %d, code = %q, want 400 invalid_collector_id", rec.Code, p.Code)
	}
}

func TestParsePathID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /c/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := httpx.ParsePathID(w, r, "id", domain.ParseCollectorID)
		if ok {
			_, _ = w.Write([]byte(id.String()))
		}
	})

	for _, tc := range []struct {
		path, body string
		want       int
	}{
		{"/c/k7m2", "k7m2", http.StatusOK},
		{"/c/!", "", http.StatusBadRequest},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

		if tc.want == http.StatusBadRequest {
			var p httpx.Problem
			_ = json.NewDecoder(rec.Body).Decode(&p)
			tc.body = ""

			if p.Code != "invalid_collector_id" {
				t.Errorf("%s code = %q", tc.path, p.Code)
			}

			rec.Body.Reset()
		}

		if rec.Code != tc.want || rec.Body.String() != tc.body {
			t.Errorf("%s = %d %q, want %d %q", tc.path, rec.Code, rec.Body.String(), tc.want, tc.body)
		}
	}
}

func TestWriteError_ProblemShape(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(logging.WithRequestID(req.Context(), "req-1"))

	httpx.WriteError(rec, req, fmt.Errorf("get collector: %w", domain.ErrCollectorNotFound))

	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("content type = %q", ct)
	}

	want := httpx.Problem{
		Type: "/problems/collector-not-found", Title: "Not Found", Status: 404,
		Detail: "collector not found", Code: "collector_not_found", RequestID: "req-1",
	}

	var got httpx.Problem
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("problem = %+v, want %+v", got, want)
	}
}

func TestDomainErrorCodes_AreUniqueSnakeCase(t *testing.T) {
	seen := map[string]error{}

	for _, err := range []*domain.Error{
		domain.ErrNameRequired, domain.ErrInvalidEmail, domain.ErrInvalidCollectorID, domain.ErrInvalidPersonaID,
		domain.ErrPersonaNameRequired, domain.ErrItemTitleRequired, domain.ErrInvalidRecurrence, domain.ErrSpecRequired,
		domain.ErrSpecMismatch, domain.ErrInvalidValueType, domain.ErrMetadataFieldNameRequired,
		domain.ErrDuplicateMetadataField, domain.ErrNameTooLong, domain.ErrPersonaNameTooLong,
		domain.ErrPersonaDescriptionTooLong, domain.ErrTooManyItems, domain.ErrItemTitleTooLong,
		domain.ErrMetadataFieldNameTooLong, domain.ErrTooManyMetadataFields, domain.ErrUnitTooLong, domain.ErrItemIDTaken, domain.ErrCollectorIDTaken,
		domain.ErrCollectorNotFound, domain.ErrPersonaNotFound, domain.ErrItemNotFound, personatemplate.ErrInvalidLocale,
	} {
		if !regexp.MustCompile(`^[a-z]+(_[a-z]+)*$`).MatchString(err.Code) {
			t.Errorf("%v: code %q is not snake_case", err, err.Code)
		}

		if prev, dup := seen[err.Code]; dup {
			t.Errorf("code %q used by %v and %v", err.Code, prev, err)
		}

		seen[err.Code] = err
	}
}

func TestDecodeJSON(t *testing.T) {
	type dst struct {
		A string `json:"a"`
	}

	for _, tc := range []struct {
		name, contentType, body string
		wantStatus              int
		wantCode                string
	}{
		{"ok", "application/json", `{"a":"x"}`, 0, ""},
		{"ok with charset", "application/json; charset=utf-8", `{"a":"x"}`, 0, ""},
		{"missing content type", "", `{"a":"x"}`, 415, "unsupported_media_type"},
		{"wrong content type", "text/plain", `{"a":"x"}`, 415, "unsupported_media_type"},
		{"syntax error", "application/json", `{`, 400, "invalid_request_body"},
		{"unknown field", "application/json", `{"b":1}`, 400, "invalid_request_body"},
		{"trailing object", "application/json", `{"a":"x"}{"a":"y"}`, 400, "invalid_request_body"},
		{"trailing garbage", "application/json", `{"a":"x"} nope`, 400, "invalid_request_body"},
		{"too large", "application/json", `{"a":"` + strings.Repeat("x", httpx.MaxBodyBytes) + `"}`, 413, "request_body_too_large"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)

			ok := httpx.DecodeJSON(rec, req, &dst{})

			if ok != (tc.wantStatus == 0) {
				t.Fatalf("ok = %v", ok)
			}

			if ok {
				return
			}

			var p httpx.Problem
			if err := json.NewDecoder(rec.Body).Decode(&p); err != nil {
				t.Fatal(err)
			}

			if rec.Code != tc.wantStatus || p.Code != tc.wantCode || p.Status != tc.wantStatus {
				t.Errorf("got %d %+v, want %d %s", rec.Code, p, tc.wantStatus, tc.wantCode)
			}
		})
	}
}
