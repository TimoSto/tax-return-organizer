package listpersonas_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/listpersonas"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

// testTaxPracticeID is the tax practice the tests act for.
var testTaxPracticeID = domaintest.TaxPracticeID()

// personaStub is a persona store whose repositories list the personas it holds, whatever tax practice they are bound
// to: tenant isolation and the unknown tax practice are the store's job and covered by the repository contract tests.
// It remembers the tax practice of the last For.
type personaStub struct {
	personas      []*domain.Persona
	taxPracticeID domain.TaxPracticeID
}

func (s *personaStub) For(id domain.TaxPracticeID) outbound.PersonaRepository {
	s.taxPracticeID = id

	return s
}

func (*personaStub) GetPersona(context.Context, domain.PersonaID) (*domain.Persona, error) {
	return nil, errors.New("personaStub: unexpected GetPersona")
}

func (s *personaStub) ListPersonas(context.Context) ([]*domain.Persona, error) {
	return s.personas, nil
}

func (*personaStub) SavePersona(context.Context, *domain.Persona) error {
	return errors.New("personaStub: unexpected SavePersona")
}

// failingStore hands out repositories whose ListPersonas fails.
type failingStore struct {
	err error
}

func (s failingStore) For(domain.TaxPracticeID) outbound.PersonaRepository { return failingRepo(s) }

type failingRepo struct {
	err error
}

func (failingRepo) GetPersona(context.Context, domain.PersonaID) (*domain.Persona, error) {
	return nil, nil
}

func (r failingRepo) ListPersonas(context.Context) ([]*domain.Persona, error) { return nil, r.err }

func (failingRepo) SavePersona(context.Context, *domain.Persona) error { return nil }

func TestExecute_Success(t *testing.T) {
	want := []*domain.Persona{
		domaintest.Persona(domaintest.PersonaID(), testTaxPracticeID, "Employed", "Has a job"),
		domaintest.Persona(domaintest.PersonaID(), testTaxPracticeID, "Letting", ""),
	}
	store := &personaStub{personas: want}

	got, err := listpersonas.New(store).Execute(context.Background(), domaintest.Staff(testTaxPracticeID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(want, got, domaintest.CmpOptions()); diff != "" {
		t.Errorf("personas mismatch (-want +got):\n%s", diff)
	}

	if store.taxPracticeID != testTaxPracticeID {
		t.Errorf("listed tax practice %s, want the actor's tax practice %s", store.taxPracticeID, testTaxPracticeID)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := listpersonas.New(failingStore{err: boom}).Execute(context.Background(), domaintest.Staff(testTaxPracticeID))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want wrapped %v", err, boom)
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
			got, err := listpersonas.New(untouchedStore{t}).Execute(context.Background(), actor)
			if !errors.Is(err, domain.ErrForbidden) {
				t.Errorf("err = %v, want %v", err, domain.ErrForbidden)
			}

			if got != nil {
				t.Errorf("got %+v, want nil", got)
			}
		})
	}
}
