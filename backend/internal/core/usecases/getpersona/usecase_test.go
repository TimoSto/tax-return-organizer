package getpersona_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	fakecollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	fakepersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getpersona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

var (
	// testCollectorID is the collector the tests act for.
	testCollectorID = domaintest.CollectorID()
	// otherCollectorID is another collector, which owns data the tests must not see.
	otherCollectorID = domaintest.CollectorID()
	// unknownCollectorID is a collector that is never stored.
	unknownCollectorID = domaintest.CollectorID()
)

// failingStore hands out repositories whose GetPersona fails.
type failingStore struct {
	err error
}

func (s failingStore) For(domain.CollectorID) outbound.PersonaRepository { return failingRepo(s) }

type failingRepo struct {
	err error
}

func (r failingRepo) GetPersona(context.Context, domain.PersonaID) (*domain.Persona, error) {
	return nil, r.err
}

func (failingRepo) ListPersonas(context.Context) ([]*domain.Persona, error) { return nil, nil }

func (failingRepo) SavePersona(context.Context, *domain.Persona) error { return nil }

// seededStore returns a persona store for the collectors testCollectorID and otherCollectorID.
func seededStore(t *testing.T) *fakepersona.Store {
	t.Helper()

	repo := fakecollector.NewRepository()

	for _, id := range []domain.CollectorID{testCollectorID, otherCollectorID} {
		if err := repo.SaveCollector(context.Background(), domaintest.Collector(id, "K", "k@example.com")); err != nil {
			t.Fatalf("seed collector: %v", err)
		}
	}

	return fakepersona.NewStore(repo)
}

func TestExecute_Success(t *testing.T) {
	store := seededStore(t)
	id := domaintest.PersonaID()
	want := domaintest.Persona(id, testCollectorID, "Employed", "Has a job",
		domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{}))

	if err := store.For(want.CollectorID()).SavePersona(context.Background(), want); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	got, err := getpersona.New(store).Execute(context.Background(), domaintest.Staff(want.CollectorID()), want.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(want, got, domaintest.CmpOptions()); diff != "" {
		t.Errorf("persona mismatch (-want +got):\n%s", diff)
	}
}

func TestExecute_NotFound(t *testing.T) {
	got, err := getpersona.New(seededStore(t)).Execute(context.Background(), domaintest.Staff(testCollectorID), domain.PersonaID(uuid.New()))
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}

	if got != nil {
		t.Errorf("persona = %+v, want nil", got)
	}
}

func TestExecute_OtherCollector(t *testing.T) {
	store := seededStore(t)
	p := domaintest.Persona(domaintest.PersonaID(), testCollectorID, "Employed", "")

	if err := store.For(p.CollectorID()).SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	_, err := getpersona.New(store).Execute(context.Background(), domaintest.Staff(otherCollectorID), p.ID())
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}
}

// The store reports the unknown collector, the use case passes it on.
func TestExecute_UnknownCollector(t *testing.T) {
	got, err := getpersona.New(seededStore(t)).Execute(context.Background(), domaintest.Staff(unknownCollectorID), domaintest.PersonaID())
	if !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrCollectorNotFound)
	}

	if got != nil {
		t.Errorf("persona = %+v, want nil", got)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := getpersona.New(failingStore{err: boom}).Execute(context.Background(), domaintest.Staff(testCollectorID), domain.PersonaID(uuid.New()))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
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
