package listpersonas_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	fakecollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	fakepersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/listpersonas"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

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

// seededStore returns a persona store for the collectors k7m2 and bbbb.
func seededStore(t *testing.T) *fakepersona.Store {
	t.Helper()

	repo := fakecollector.NewRepository()

	for _, id := range []domain.CollectorID{"k7m2", "bbbb"} {
		if err := repo.SaveCollector(context.Background(), domaintest.Collector(id, "K", "k@example.com")); err != nil {
			t.Fatalf("seed collector: %v", err)
		}
	}

	return fakepersona.NewStore(repo)
}

func TestExecute_Success(t *testing.T) {
	store := seededStore(t)
	ctx := context.Background()
	employed := domaintest.Persona(domaintest.PersonaID(), "k7m2", "Employed", "Has a job")
	letting := domaintest.Persona(domaintest.PersonaID(), "k7m2", "Letting", "")
	other := domaintest.Persona(domaintest.PersonaID(), "bbbb", "Other", "")

	for _, p := range []*domain.Persona{letting, other, employed} {
		if err := store.For(p.CollectorID()).SavePersona(ctx, p); err != nil {
			t.Fatalf("seed persona: %v", err)
		}
	}

	got, err := listpersonas.New(store).Execute(ctx, domaintest.Staff("k7m2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff([]*domain.Persona{employed, letting}, got, domaintest.CmpOptions()); diff != "" {
		t.Errorf("personas mismatch (-want +got):\n%s", diff)
	}
}

// The store reports the unknown collector, the use case passes it on.
func TestExecute_UnknownCollector(t *testing.T) {
	got, err := listpersonas.New(seededStore(t)).Execute(context.Background(), domaintest.Staff("zzzz"))
	if !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrCollectorNotFound)
	}

	if got != nil {
		t.Errorf("personas = %+v, want nil", got)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := listpersonas.New(failingStore{err: boom}).Execute(context.Background(), domaintest.Staff("k7m2"))
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
		"client": domaintest.Client("k7m2"),
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
