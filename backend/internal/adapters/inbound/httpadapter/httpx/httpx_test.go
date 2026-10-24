package httpx_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
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
		{domain.ErrItemIDTaken, http.StatusConflict},
		{domain.ErrUnknownCollector, http.StatusConflict},
		{domain.ErrInvalidLocale, http.StatusBadRequest},
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

			var body httpx.ErrorResponse
			if e := json.NewDecoder(rec.Body).Decode(&body); e != nil {
				t.Fatal(e)
			}

			if body.Error != domain.ErrInvalidEmail.Error() {
				t.Errorf("error = %q, want %q", body.Error, domain.ErrInvalidEmail.Error())
			}
		})
	}
}

func TestWriteError_InternalErrorIsGeneric(t *testing.T) {
	rec := httptest.NewRecorder()

	httpx.WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), errors.New("database down"))

	var body httpx.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	if body.Error != "internal error" {
		t.Errorf("error = %q, want generic message", body.Error)
	}
}

func TestParsePathID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /c/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := httpx.ParsePathID(w, r, "id", domain.ParseCollectorID, "collector id")
		if ok {
			_, _ = w.Write([]byte(id.String()))
		}
	})

	for _, tc := range []struct {
		path, body string
		want       int
	}{
		{"/c/k7m2", "k7m2", http.StatusOK},
		{"/c/!", `{"error":"invalid collector id"}` + "\n", http.StatusBadRequest},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

		if rec.Code != tc.want || rec.Body.String() != tc.body {
			t.Errorf("%s = %d %q, want %d %q", tc.path, rec.Code, rec.Body.String(), tc.want, tc.body)
		}
	}
}
