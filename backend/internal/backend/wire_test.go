package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	fakecollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	fakepersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	faketransactor "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/transactor"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

// newFakeRepositories returns in-memory repositories. The persona repository accepts personas of the stored
// collectors, and of collector k7m2, which is stored up front.
func newFakeRepositories() port.Repositories {
	collectors := fakecollector.NewRepository()
	if err := collectors.SaveCollector(context.Background(), &domain.Collector{ID: "k7m2", Name: "K", Email: "k@example.com"}); err != nil {
		panic(err)
	}

	return port.Repositories{Collectors: collectors, Personas: fakepersona.NewRepositoryFor(collectors)}
}

func wireFake(repos port.Repositories) http.Handler {
	return wire(repos, faketransactor.New(repos))
}

func TestWire_CreateCollector(t *testing.T) {
	h := wireFake(newFakeRepositories())

	req := httptest.NewRequest(http.MethodPost, "/collectors", strings.NewReader(`{"name":"Acme","email":"a@example.com"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestWire_Healthz(t *testing.T) {
	h := wireFake(newFakeRepositories())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("missing X-Request-ID header")
	}
}

func TestWire_GetCollector(t *testing.T) {
	h := wireFake(newFakeRepositories())

	create := httptest.NewRecorder()
	h.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/collectors", strings.NewReader(`{"name":"Acme","email":"a@example.com"}`)))

	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/collectors/"+created.ID, nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestWire_GetCollector_NotFound(t *testing.T) {
	h := wireFake(newFakeRepositories())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/collectors/zzzz", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestWire_GetPersona(t *testing.T) {
	repos := newFakeRepositories()
	p := &domain.Persona{ID: uuid.New(), CollectorID: "k7m2", Name: "Employed"}

	if err := repos.Personas.SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	h := wireFake(repos)

	for _, tc := range []struct {
		name, path string
		want       int
	}{
		{"found", "/collectors/k7m2/personas/" + p.ID.String(), http.StatusOK},
		{"other collector", "/collectors/bbbb/personas/" + p.ID.String(), http.StatusNotFound},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

		if rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}

func TestWire_ListPersonas(t *testing.T) {
	repos := newFakeRepositories()
	p := &domain.Persona{ID: uuid.New(), CollectorID: "k7m2", Name: "Employed"}

	if err := repos.Personas.SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	h := wireFake(repos)

	for _, tc := range []struct {
		name, path string
		want       int
	}{
		{"found", "/collectors/k7m2/personas", http.StatusOK},
		{"unknown collector", "/collectors/bbbb/personas", http.StatusNotFound},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

		if rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}
