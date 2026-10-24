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
)

type failingRepo struct {
	port.PersonaRepository
	err error
}

func (r failingRepo) GetPersona(context.Context, domain.CollectorID, uuid.UUID) (*domain.Persona, error) {
	return nil, r.err
}

func TestExecute_Success(t *testing.T) {
	repo := fakepersona.NewRepository("k7m2")
	id := uuid.New()
	want := &domain.Persona{
		ID: id, CollectorID: "k7m2", Name: "Employed", Description: "Has a job",
		Items: []domain.ItemDefinition{{
			ID: uuid.New(), Title: "Payslip",
			Recurrence: domain.RecurrenceMonthly, Spec: domain.DocumentSpec{},
		}},
	}

	if err := repo.SavePersona(context.Background(), want); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	got, err := getpersona.New(repo).Execute(tenant.WithCollector(context.Background(), want.CollectorID), want.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("persona mismatch (-want +got):\n%s", diff)
	}
}

func TestExecute_NotFound(t *testing.T) {
	got, err := getpersona.New(fakepersona.NewRepository()).Execute(tenant.WithCollector(context.Background(), "k7m2"), uuid.New())
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}

	if got != nil {
		t.Errorf("persona = %+v, want nil", got)
	}
}

func TestExecute_OtherCollector(t *testing.T) {
	repo := fakepersona.NewRepository("k7m2")
	p := &domain.Persona{ID: uuid.New(), CollectorID: "k7m2", Name: "Employed"}

	if err := repo.SavePersona(context.Background(), p); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	_, err := getpersona.New(repo).Execute(tenant.WithCollector(context.Background(), "bbbb"), p.ID)
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := getpersona.New(failingRepo{err: boom}).Execute(tenant.WithCollector(context.Background(), "k7m2"), uuid.New())
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}

func TestExecute_NoTenant(t *testing.T) {
	_, err := getpersona.New(fakepersona.NewRepository()).Execute(context.Background(), uuid.New())
	if !errors.Is(err, tenant.ErrNoTenant) {
		t.Errorf("err = %v, want %v", err, tenant.ErrNoTenant)
	}
}
