package collector_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

func doGet(t *testing.T, get ports.GetCollectorPort, method, id string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/collectors/"+id, nil)
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := collector.Get(slog.New(slog.DiscardHandler), get)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestGetCollector_OK(t *testing.T) {
	var gotID uuid.UUID

	rec := doGet(t, func(_ context.Context, id uuid.UUID) (*domain.Collector, error) {
		gotID = id

		return &domain.Collector{ID: testID, Name: "Jane", Email: "jane@example.com"}, nil
	}, http.MethodGet, testID.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if gotID != testID {
		t.Errorf("port got id %v, want %v", gotID, testID)
	}

	want := `{"id":"` + testID.String() + `","name":"Jane","email":"jane@example.com"}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestGetCollector_Errors(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		portErr error
		want    int
	}{
		{"invalid id", "not-a-uuid", nil, http.StatusBadRequest},
		{"not found", testID.String(), domain.ErrCollectorNotFound, http.StatusNotFound},
		{"internal", testID.String(), errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doGet(t, func(context.Context, uuid.UUID) (*domain.Collector, error) {
				return nil, tt.portErr
			}, http.MethodGet, tt.id)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestGetCollector_MethodNotAllowed(t *testing.T) {
	rec := doGet(t, nil, http.MethodPost, testID.String())

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
