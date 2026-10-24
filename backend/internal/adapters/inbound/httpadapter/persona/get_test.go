package persona_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/tenant"
)

const testCollectorID domain.CollectorID = "k7m2"

var testID = domain.PersonaID(uuid.MustParse("3f2b8c1e-5d4a-4e6f-9a7b-1c2d3e4f5a6b"))

type getFunc func(context.Context, domain.PersonaID) (*domain.Persona, error)

func (f getFunc) Execute(ctx context.Context, id domain.PersonaID) (*domain.Persona, error) {
	return f(ctx, id)
}

func doGet(t *testing.T, get getFunc, method, id string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/collectors/"+testCollectorID.String()+"/personas/"+id, nil)
	req = req.WithContext(tenant.WithCollector(req.Context(), testCollectorID))
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	route := persona.Get(get)
	mux.Handle(route.Pattern, route.Handler)
	mux.ServeHTTP(rec, req)

	return rec
}

func TestGetPersona_OK(t *testing.T) {
	var gotID domain.PersonaID

	rec := doGet(t, func(_ context.Context, id domain.PersonaID) (*domain.Persona, error) {
		gotID = id

		return domaintest.Persona(testID, testCollectorID, "Employed", "Has a job"), nil
	}, http.MethodGet, testID.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if gotID != testID {
		t.Errorf("port got %v, want %v", gotID, testID)
	}

	want := `{"id":"` + testID.String() + `","collectorId":"` + testCollectorID.String() +
		`","name":"Employed","description":"Has a job","items":[]}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestGetPersona_Errors(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		portErr error
		want    int
	}{
		{"invalid persona id", "not-a-uuid", nil, http.StatusBadRequest},
		{"not found", testID.String(), domain.ErrPersonaNotFound, http.StatusNotFound},
		{"internal", testID.String(), errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doGet(t, func(context.Context, domain.PersonaID) (*domain.Persona, error) {
				return nil, tt.portErr
			}, http.MethodGet, tt.id)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestGetPersona_MethodNotAllowed(t *testing.T) {
	rec := doGet(t, nil, http.MethodPost, testID.String())

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestGetPersona_Items(t *testing.T) {
	docID, dataID := domaintest.ItemID(), domaintest.ItemID()

	rec := doGet(t, func(context.Context, domain.PersonaID) (*domain.Persona, error) {
		return domaintest.Persona(testID, testCollectorID, "Employed", "",
			domaintest.Item(docID, "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{MetadataFields: []domain.MetadataField{
				{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
			}}),
			domaintest.Item(dataID, "Commuting days", domain.RecurrenceAnnually,
				domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days"}),
		), nil
	}, http.MethodGet, testID.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	want := `"items":[` +
		`{"id":"` + docID.String() + `","title":"Payslip","kind":"document","recurrence":"monthly",` +
		`"documentSpec":{"metadataFields":[{"name":"Employer","valueSpec":{"type":"text","unit":"","required":true}}]},"valueSpec":null},` +
		`{"id":"` + dataID.String() + `","title":"Commuting days","kind":"data","recurrence":"annually",` +
		`"documentSpec":null,"valueSpec":{"type":"number","unit":"days","required":false}}]`
	if got := rec.Body.String(); !strings.Contains(got, want) {
		t.Errorf("body = %s, want it to contain %s", got, want)
	}
}
