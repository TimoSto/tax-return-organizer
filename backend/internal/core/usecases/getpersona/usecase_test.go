package getpersona_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getpersona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

// testTaxPracticeID is the tax practice the tests act for.
var testTaxPracticeID = domaintest.TaxPracticeID()

// personaStub is a persona store whose repositories return the personas it holds, whatever tax practice they are
// bound to: tenant isolation and the unknown tax practice are the store's job and covered by the repository contract
// tests. It remembers the tax practice of the last For.
type personaStub struct {
	personas      []*domain.Persona
	taxPracticeID domain.TaxPracticeID
}

func (s *personaStub) For(id domain.TaxPracticeID) outbound.PersonaRepository {
	s.taxPracticeID = id

	return s
}

func (s *personaStub) GetPersona(_ context.Context, id domain.PersonaID) (*domain.Persona, error) {
	for _, p := range s.personas {
		if p.ID() == id {
			return p, nil
		}
	}

	return nil, domain.ErrPersonaNotFound
}

func (s *personaStub) ListPersonas(context.Context) ([]*domain.Persona, error) {
	return s.personas, nil
}

func (*personaStub) SavePersona(context.Context, *domain.Persona) error {
	return errors.New("personaStub: unexpected SavePersona")
}

// failingStore hands out repositories whose GetPersona fails.
type failingStore struct {
	err error
}

func (s failingStore) For(domain.TaxPracticeID) outbound.PersonaRepository { return failingRepo(s) }

type failingRepo struct {
	err error
}

func (r failingRepo) GetPersona(context.Context, domain.PersonaID) (*domain.Persona, error) {
	return nil, r.err
}

func (failingRepo) ListPersonas(context.Context) ([]*domain.Persona, error) { return nil, nil }

func (failingRepo) SavePersona(context.Context, *domain.Persona) error { return nil }

func TestExecute_Success(t *testing.T) {
	want := domaintest.Persona(domaintest.PersonaID(), testTaxPracticeID, "Employed", "Has a job",
		domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{}))
	store := &personaStub{personas: []*domain.Persona{want}}

	got, err := getpersona.New(store).Execute(context.Background(), domaintest.Staff(want.TaxPracticeID()), want.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(want, got, domaintest.CmpOptions()); diff != "" {
		t.Errorf("persona mismatch (-want +got):\n%s", diff)
	}

	if store.taxPracticeID != testTaxPracticeID {
		t.Errorf("read from tax practice %s, want the actor's tax practice %s", store.taxPracticeID, testTaxPracticeID)
	}
}

func TestExecute_NotFound(t *testing.T) {
	got, err := getpersona.New(&personaStub{}).Execute(context.Background(), domaintest.Staff(testTaxPracticeID), domain.PersonaID(uuid.New()))
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}

	if got != nil {
		t.Errorf("persona = %+v, want nil", got)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := getpersona.New(failingStore{err: boom}).Execute(context.Background(), domaintest.Staff(testTaxPracticeID), domain.PersonaID(uuid.New()))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}

// untouchedStore fails the test if the use case asks it for a repository.
type untouchedStore struct {
	t *testing.T
}

func (s untouchedStore) For(domain.TaxPracticeID) outbound.PersonaRepository {
	s.t.Helper()
	s.t.Fatal("store was used, want the actor rejected before")

	return nil
}

// Only staff may see personas. The actor is rejected before the store is asked, so a client does not even learn
// whether its tax practice exists.
func TestExecute_Forbidden(t *testing.T) {
	for name, actor := range map[string]domain.Actor{
		"client": domaintest.Client(testTaxPracticeID),
		"zero":   {},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := getpersona.New(untouchedStore{t}).Execute(context.Background(), actor, domaintest.PersonaID())
			if !errors.Is(err, domain.ErrForbidden) {
				t.Errorf("err = %v, want %v", err, domain.ErrForbidden)
			}

			if got != nil {
				t.Errorf("got %+v, want nil", got)
			}
		})
	}
}
