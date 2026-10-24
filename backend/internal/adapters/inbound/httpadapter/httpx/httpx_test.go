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
		{domain.ErrInvalidKind, http.StatusBadRequest},
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
