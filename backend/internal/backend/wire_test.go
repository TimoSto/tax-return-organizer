package backend

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/health"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/postgrestest"
)

// One container is shared by all tests in this package.
func TestMain(m *testing.M) {
	os.Exit(postgrestest.Run(m))
}

// unknownCollectorID is a collector that is never stored.
var unknownCollectorID = domaintest.CollectorID()

// collectorSessions takes the token as a collector id and resolves it to an employee of that collector, so a test
// can act for any collector.
type collectorSessions struct{}

func (collectorSessions) ResolveSession(_ context.Context, token string) (domain.Actor, bool, error) {
	id, err := domain.ParseCollectorID(token)
	if err != nil {
		return domain.Actor{}, false, nil //nolint:nilerr // an unparsable token is unknown, not a failure
	}

	return domain.NewStaffActor(id, domain.UserID{}), true, nil
}

// get sends a GET request with a session for collectorID.
func get(h http.Handler, path string, collectorID domain.CollectorID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+collectorID.String())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

type pinger struct{ err error }

func (p pinger) Ping(context.Context) error { return p.err }

// newHandler wires the use cases on the Postgres repositories as Run does. It stores a new collector and returns
// its id: every test works in its own collector, so the tests share the database without seeing each other's data.
func newHandler(t *testing.T, ready health.Pinger) (http.Handler, Repositories, domain.CollectorID) {
	t.Helper()

	pool := postgrestest.RequirePool(t)
	repos := postgresRepositories(pool)
	id := domaintest.CollectorID()

	if err := repos.Collectors.SaveCollector(context.Background(), domaintest.Collector(id, "K", "k@example.com")); err != nil {
		t.Fatalf("seed collector: %v", err)
	}

	h, err := wire(repos, NewTransactor(pool), ready, collectorSessions{})
	if err != nil {
		t.Fatalf("wire: %v", err)
	}

	return h, repos, id
}

func TestWire_CreateCollector(t *testing.T) {
	h, _, _ := newHandler(t, pinger{})

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
			h, _, _ := newHandler(t, pinger{err: tc.err})

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestWire_Healthz(t *testing.T) {
	h, _, _ := newHandler(t, pinger{})

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
	h, _, collectorID := newHandler(t, pinger{})

	if rec := get(h, "/collector", collectorID); rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestWire_GetCollector_NotFound(t *testing.T) {
	h, _, _ := newHandler(t, pinger{})

	if rec := get(h, "/collector", unknownCollectorID); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestWire_Unauthenticated(t *testing.T) {
	h, _, _ := newHandler(t, pinger{})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/collector", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWire_GetPersona(t *testing.T) {
	h, repos, collectorID := newHandler(t, pinger{})
	p := domaintest.Persona(domaintest.PersonaID(), collectorID, "Employed", "")

	if err := repos.Personas.For(p.CollectorID()).SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	for _, tc := range []struct {
		name  string
		actor domain.CollectorID
		want  int
	}{
		{"found", collectorID, http.StatusOK},
		{"other collector", unknownCollectorID, http.StatusNotFound},
	} {
		if rec := get(h, "/personas/"+p.ID().String(), tc.actor); rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}

func TestWire_ListPersonas(t *testing.T) {
	h, repos, collectorID := newHandler(t, pinger{})
	p := domaintest.Persona(domaintest.PersonaID(), collectorID, "Employed", "")

	if err := repos.Personas.For(p.CollectorID()).SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	for _, tc := range []struct {
		name  string
		actor domain.CollectorID
		want  int
	}{
		{"found", collectorID, http.StatusOK},
		{"unknown collector", unknownCollectorID, http.StatusNotFound},
	} {
		if rec := get(h, "/personas", tc.actor); rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}
