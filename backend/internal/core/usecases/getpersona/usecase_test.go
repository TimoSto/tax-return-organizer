package getpersona_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	fakepersona "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/tenant"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getpersona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

// failingStore hands out repositories whose GetPersona fails.
type failingStore struct {
	err error
}

func (s failingStore) For(domain.CollectorID) port.PersonaRepository { return failingRepo(s) }

type failingRepo struct {
	err error
}

func (r failingRepo) GetPersona(context.Context, domain.PersonaID) (*domain.Persona, error) {
	return nil, r.err
}

func (failingRepo) ListPersonas(context.Context) ([]*domain.Persona, error) { return nil, nil }

func (failingRepo) SavePersona(context.Context, *domain.Persona) error { return nil }

func TestExecute_Success(t *testing.T) {
	store := fakepersona.NewStore()
	id := domaintest.PersonaID()
	want := domaintest.Persona(id, "k7m2", "Employed", "Has a job",
		domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{}))

	if err := store.For(want.CollectorID()).SavePersona(context.Background(), want); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	got, err := getpersona.New(store).Execute(tenant.With(context.Background(), want.CollectorID()), want.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(want, got, domaintest.CmpOptions()); diff != "" {
		t.Errorf("persona mismatch (-want +got):\n%s", diff)
	}
}

func TestExecute_NotFound(t *testing.T) {
	got, err := getpersona.New(fakepersona.NewStore()).Execute(tenant.With(context.Background(), "k7m2"), domain.PersonaID(uuid.New()))
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}

	if got != nil {
		t.Errorf("persona = %+v, want nil", got)
	}
}

func TestExecute_OtherCollector(t *testing.T) {
	store := fakepersona.NewStore()
	p := domaintest.Persona(domaintest.PersonaID(), "k7m2", "Employed", "")

	if err := store.For(p.CollectorID()).SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	_, err := getpersona.New(store).Execute(tenant.With(context.Background(), "bbbb"), p.ID())
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := getpersona.New(failingStore{err: boom}).Execute(tenant.With(context.Background(), "k7m2"), domain.PersonaID(uuid.New()))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}

func TestExecute_NoTenant(t *testing.T) {
	_, err := getpersona.New(fakepersona.NewStore()).Execute(context.Background(), domain.PersonaID(uuid.New()))
	if !errors.Is(err, tenant.ErrNoTenant) {
		t.Errorf("err = %v, want %v", err, tenant.ErrNoTenant)
	}
}
