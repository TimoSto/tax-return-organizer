package persona_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

const testCollectorID domain.CollectorID = "k7m2"

var testID = uuid.MustParse("3f2b8c1e-5d4a-4e6f-9a7b-1c2d3e4f5a6b")

func doGet(t *testing.T, get ports.GetPersonaPort, method, collectorID, id string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/collectors/"+collectorID+"/personas/"+id, nil)
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := persona.Get(slog.New(slog.DiscardHandler), get)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestGetPersona_OK(t *testing.T) {
	var (
		gotCollector domain.CollectorID
		gotID        uuid.UUID
	)

	rec := doGet(t, func(_ context.Context, collectorID domain.CollectorID, id uuid.UUID) (*domain.Persona, error) {
		gotCollector, gotID = collectorID, id

		return &domain.Persona{ID: testID, CollectorID: testCollectorID, Name: "Employed", Description: "Has a job"}, nil
	}, http.MethodGet, testCollectorID.String(), testID.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if gotCollector != testCollectorID || gotID != testID {
		t.Errorf("port got (%v, %v), want (%v, %v)", gotCollector, gotID, testCollectorID, testID)
	}

	want := `{"id":"` + testID.String() + `","collectorId":"` + testCollectorID.String() +
		`","name":"Employed","description":"Has a job","items":[]}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestGetPersona_Errors(t *testing.T) {
	tests := []struct {
		name        string
		collectorID string
		id          string
		portErr     error
		want        int
	}{
		{"invalid collector id", "not-an-id", testID.String(), nil, http.StatusBadRequest},
		{"invalid persona id", testCollectorID.String(), "not-a-uuid", nil, http.StatusBadRequest},
		{"not found", testCollectorID.String(), testID.String(), domain.ErrPersonaNotFound, http.StatusNotFound},
		{"internal", testCollectorID.String(), testID.String(), errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doGet(t, func(context.Context, domain.CollectorID, uuid.UUID) (*domain.Persona, error) {
				return nil, tt.portErr
			}, http.MethodGet, tt.collectorID, tt.id)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestGetPersona_MethodNotAllowed(t *testing.T) {
	rec := doGet(t, nil, http.MethodPost, testCollectorID.String(), testID.String())

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestGetPersona_Items(t *testing.T) {
	docID, dataID := uuid.New(), uuid.New()

	rec := doGet(t, func(context.Context, domain.CollectorID, uuid.UUID) (*domain.Persona, error) {
		return &domain.Persona{
			ID: testID, CollectorID: testCollectorID, Name: "Employed",
			Items: []domain.ItemDefinition{
				{
					ID: docID, PersonaID: testID, Title: "Payslip", Kind: domain.KindDocument,
					Recurrence: domain.RecurrenceMonthly,
					DocumentSpec: &domain.DocumentSpec{MetadataFields: []domain.MetadataField{
						{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
					}},
				},
				{
					ID: dataID, PersonaID: testID, Title: "Commuting days", Kind: domain.KindData,
					Recurrence: domain.RecurrenceAnnually,
					ValueSpec:  &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days"},
				},
			},
		}, nil
	}, http.MethodGet, testCollectorID.String(), testID.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	want := `"items":[` +
		`{"id":"` + docID.String() + `","title":"Payslip","kind":"document","recurrence":"monthly",` +
		`"documentSpec":{"metadataFields":[{"name":"Employer","spec":{"type":"text","unit":"","required":true}}]},"valueSpec":null},` +
		`{"id":"` + dataID.String() + `","title":"Commuting days","kind":"data","recurrence":"annually",` +
		`"documentSpec":null,"valueSpec":{"type":"number","unit":"days","required":false}}]`
	if got := rec.Body.String(); !strings.Contains(got, want) {
		t.Errorf("body = %s, want it to contain %s", got, want)
	}
}
