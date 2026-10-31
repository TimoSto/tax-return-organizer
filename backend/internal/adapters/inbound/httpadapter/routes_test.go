package httpadapter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
)

type stubDeps struct{}

func (stubDeps) Execute(context.Context, inbound.CreateCollectorCommand) (domain.CollectorID, error) {
	return domain.CollectorID{}, nil
}

type stubGetCollector struct{}

func (stubGetCollector) Execute(context.Context, domain.Actor) (*domain.Collector, error) {
	return nil, nil
}

type stubGetPersona struct{}

func (stubGetPersona) Execute(context.Context, domain.Actor, domain.PersonaID) (*domain.Persona, error) {
	return nil, nil
}

type stubListPersonas struct{}

func (stubListPersonas) Execute(context.Context, domain.Actor) ([]*domain.Persona, error) {
	return nil, nil
}

type stubPinger struct{}

func (stubPinger) Ping(context.Context) error { return nil }

func fullDeps() httpadapter.Deps {
	return httpadapter.Deps{
		CreateCollector: stubDeps{},
		GetCollector:    stubGetCollector{},
		GetPersona:      stubGetPersona{},
		ListPersonas:    stubListPersonas{},
		Readiness:       stubPinger{},
	}
}

func newHandler(t *testing.T) http.Handler {
	t.Helper()

	h, err := httpadapter.NewHandler(fullDeps())
	if err != nil {
		t.Fatal(err)
	}

	return h
}

func TestNewHandler_RejectsMissingDeps(t *testing.T) {
	if _, err := httpadapter.NewHandler(httpadapter.Deps{}); err == nil ||
		!strings.Contains(err.Error(), "CreateCollector, GetCollector, GetPersona, ListPersonas, Readiness") {
		t.Errorf("err = %v, want all five fields named", err)
	}

	d := fullDeps()
	d.GetPersona = nil

	if _, err := httpadapter.NewHandler(d); err == nil || !strings.Contains(err.Error(), "GetPersona") || strings.Contains(err.Error(), "ListPersonas") {
		t.Errorf("err = %v, want only GetPersona named", err)
	}
}

// TestNewHandler_ChecksEveryDep sets each field of fullDeps to nil in turn, so a new field of Deps is covered without
// a test change (fullDeps must set it, or TestNewHandler fails).
func TestNewHandler_ChecksEveryDep(t *testing.T) {
	typ := reflect.TypeFor[httpadapter.Deps]()

	for i := range typ.NumField() {
		name := typ.Field(i).Name

		d := fullDeps()
		v := reflect.ValueOf(&d).Elem().Field(i)
		v.Set(reflect.Zero(v.Type()))

		if _, err := httpadapter.NewHandler(d); err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s = nil: err = %v, want it named", name, err)
		}
	}
}

func TestNewHandler(t *testing.T) {
	h := newHandler(t)

	for _, tt := range []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/healthz", http.StatusOK},
		{http.MethodGet, "/readyz", http.StatusOK},
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

func TestNewHandler_InvalidCollectorID(t *testing.T) {
	h := newHandler(t)

	for _, path := range []string{"/collectors/!", "/collectors/!/personas", "/collectors/!/personas/3f2b8c1e-5d4a-4e6f-9a7b-1c2d3e4f5a6b"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusBadRequest {
			t.Errorf("GET %s = %d, want 400", path, rec.Code)
		}
	}
}
