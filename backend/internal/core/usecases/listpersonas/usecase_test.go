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

// testCollectorID is the collector the tests act for.
var testCollectorID = domaintest.CollectorID()

// personaStub is a persona store whose repositories list the personas it holds, whatever collector they are bound
// to: tenant isolation, ordering and the unknown collector are the store's job and covered by the repository
// contract tests. It remembers the collector of the last For.
type personaStub struct {
	personas    []*domain.Persona
	collectorID domain.CollectorID
}

func (s *personaStub) For(id domain.CollectorID) outbound.PersonaRepository {
	s.collectorID = id

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

func (s failingStore) For(domain.CollectorID) outbound.PersonaRepository { return failingRepo(s) }

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
		domaintest.Persona(domaintest.PersonaID(), testCollectorID, "Employed", "Has a job"),
		domaintest.Persona(domaintest.PersonaID(), testCollectorID, "Letting", ""),
	}
	store := &personaStub{personas: want}

	got, err := listpersonas.New(store).Execute(context.Background(), domaintest.Staff(testCollectorID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(want, got, domaintest.CmpOptions()); diff != "" {
		t.Errorf("personas mismatch (-want +got):\n%s", diff)
	}

	if store.collectorID != testCollectorID {
		t.Errorf("listed collector %s, want the actor's collector %s", store.collectorID, testCollectorID)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := listpersonas.New(failingStore{err: boom}).Execute(context.Background(), domaintest.Staff(testCollectorID))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want wrapped %v", err, boom)
	}
}

// untouchedStore fails the test if the use case asks it for a repository.
type untouchedStore struct {
	t *testing.T
}

func (s untouchedStore) For(domain.CollectorID) outbound.PersonaRepository {
	s.t.Helper()
	s.t.Fatal("store was used, want the actor rejected before")

	return nil
}

// Only staff may see personas. The actor is rejected before the store is asked, so a client does not even learn
// whether its collector exists.
func TestExecute_Forbidden(t *testing.T) {
	for name, actor := range map[string]domain.Actor{
		"client": domaintest.Client(testCollectorID),
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
