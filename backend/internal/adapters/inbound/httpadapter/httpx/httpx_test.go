package httpx_test

import (
	"errors"
	"fmt"
	"log/slog"
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
		{domain.ErrItemIDTaken, http.StatusBadRequest},
		{domain.ErrUnknownCollector, http.StatusBadRequest},
		{domain.ErrInvalidLocale, http.StatusBadRequest},
		{fmt.Errorf("wrapped: %w", domain.ErrSpecMismatch), http.StatusBadRequest},
		{domain.ErrCollectorNotFound, http.StatusNotFound},
		{domain.ErrPersonaNotFound, http.StatusNotFound},
		{domain.ErrItemNotFound, http.StatusNotFound},
		{errors.New("database down"), http.StatusInternalServerError},
	} {
		t.Run(tc.err.Error(), func(t *testing.T) {
			rec := httptest.NewRecorder()

			httpx.WriteError(slog.New(slog.DiscardHandler), rec, httptest.NewRequest(http.MethodGet, "/", nil), tc.err)

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
