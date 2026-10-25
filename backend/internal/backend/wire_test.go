package backend

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/health"
	fakecollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	fakepersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	faketransactor "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/transactor"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

// newFakeRepositories returns in-memory repositories. The persona repository accepts personas of the stored
// collectors, and of collector k7m2, which is stored up front.
func newFakeRepositories() outbound.Repositories {
	collectors := fakecollector.NewRepository()
	if err := collectors.SaveCollector(context.Background(), domaintest.Collector("k7m2", "K", "k@example.com")); err != nil {
		panic(err)
	}

	return outbound.Repositories{Collectors: collectors, Personas: fakepersona.NewStore()}
}

type pinger struct{ err error }

func (p pinger) Ping(context.Context) error { return p.err }

func wireFake(repos outbound.Repositories) http.Handler {
	return wireFakeReady(repos, pinger{})
}

func wireFakeReady(repos outbound.Repositories, ready health.Pinger) http.Handler {
	h, err := wire(repos, faketransactor.New(repos), ready)
	if err != nil {
		panic(err)
	}

	return h
}

func TestWire_CreateCollector(t *testing.T) {
	h := wireFake(newFakeRepositories())

	req := httptest.NewRequest(http.MethodPost, "/collectors", strings.NewReader(`{"name":"Acme","email":"a@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestWire_Readyz(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{"database up", nil, http.StatusOK},
		{"database down", errors.New("down"), http.StatusServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := wireFakeReady(newFakeRepositories(), pinger{err: tc.err})

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
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
	createReq := httptest.NewRequest(http.MethodPost, "/collectors", strings.NewReader(`{"name":"Acme","email":"a@example.com"}`))
	createReq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(create, createReq)

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
	p := domaintest.Persona(domaintest.PersonaID(), "k7m2", "Employed", "")

	if err := repos.Personas.For(p.CollectorID()).SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	h := wireFake(repos)

	for _, tc := range []struct {
		name, path string
		want       int
	}{
		{"found", "/collectors/k7m2/personas/" + p.ID().String(), http.StatusOK},
		{"other collector", "/collectors/bbbb/personas/" + p.ID().String(), http.StatusNotFound},
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
	p := domaintest.Persona(domaintest.PersonaID(), "k7m2", "Employed", "")

	if err := repos.Personas.For(p.CollectorID()).SavePersona(context.Background(), p); err != nil {
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
