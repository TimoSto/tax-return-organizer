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

// unknownTaxPracticeID is a tax practice that is never stored.
var unknownTaxPracticeID = domaintest.TaxPracticeID()

// taxPracticeSessions takes the token as a tax practice id and resolves it to an employee of that tax practice, so a test
// can act for any tax practice.
//
// TODO(auth): delete once wire builds ResolveSession; the tests then log in, or seed a session, per tax practice.
type taxPracticeSessions struct{}

func (taxPracticeSessions) ResolveSession(_ context.Context, token string) (domain.Actor, bool, error) {
	id, err := domain.ParseTaxPracticeID(token)
	if err != nil {
		return domain.Actor{}, false, nil //nolint:nilerr // an unparsable token is unknown, not a failure
	}

	return domain.NewStaffActor(id, domain.UserID{}), true, nil
}

// get sends a GET request with a session for taxPracticeID.
func get(h http.Handler, path string, taxPracticeID domain.TaxPracticeID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+taxPracticeID.String())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

type pinger struct{ err error }

func (p pinger) Ping(context.Context) error { return p.err }

// newHandler wires the use cases on the Postgres repositories as Run does. It stores a new tax practice and returns
// its id: every test works in its own tax practice, so the tests share the database without seeing each other's data.
func newHandler(t *testing.T, ready health.Pinger) (http.Handler, Repositories, domain.TaxPracticeID) {
	t.Helper()

	pool := postgrestest.RequirePool(t)
	repos := postgresRepositories(pool)
	id := domaintest.TaxPracticeID()

	if err := repos.TaxPractices.SaveTaxPractice(context.Background(), domaintest.TaxPractice(id, "K", "k@example.com")); err != nil {
		t.Fatalf("seed tax practice: %v", err)
	}

	h, err := wire(repos, NewTransactor(pool), ready, taxPracticeSessions{})
	if err != nil {
		t.Fatalf("wire: %v", err)
	}

	return h, repos, id
}

func TestWire_CreateTaxPractice(t *testing.T) {
	h, _, _ := newHandler(t, pinger{})

	req := httptest.NewRequest(http.MethodPost, "/tax-practices", strings.NewReader(`{"name":"Acme","email":"a@example.com"}`))
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

func TestWire_GetTaxPractice(t *testing.T) {
	h, _, taxPracticeID := newHandler(t, pinger{})

	if rec := get(h, "/tax-practice", taxPracticeID); rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestWire_GetTaxPractice_NotFound(t *testing.T) {
	h, _, _ := newHandler(t, pinger{})

	if rec := get(h, "/tax-practice", unknownTaxPracticeID); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestWire_Unauthenticated(t *testing.T) {
	h, _, _ := newHandler(t, pinger{})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tax-practice", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWire_GetPersona(t *testing.T) {
	h, repos, taxPracticeID := newHandler(t, pinger{})
	p := domaintest.Persona(domaintest.PersonaID(), taxPracticeID, "Employed", "")

	if err := repos.Personas.For(p.TaxPracticeID()).SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	for _, tc := range []struct {
		name  string
		actor domain.TaxPracticeID
		want  int
	}{
		{"found", taxPracticeID, http.StatusOK},
		{"other tax practice", unknownTaxPracticeID, http.StatusNotFound},
	} {
		if rec := get(h, "/personas/"+p.ID().String(), tc.actor); rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}

func TestWire_ListPersonas(t *testing.T) {
	h, repos, taxPracticeID := newHandler(t, pinger{})
	p := domaintest.Persona(domaintest.PersonaID(), taxPracticeID, "Employed", "")

	if err := repos.Personas.For(p.TaxPracticeID()).SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	for _, tc := range []struct {
		name  string
		actor domain.TaxPracticeID
		want  int
	}{
		{"found", taxPracticeID, http.StatusOK},
		{"unknown tax practice", unknownTaxPracticeID, http.StatusNotFound},
	} {
		if rec := get(h, "/personas", tc.actor); rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}
