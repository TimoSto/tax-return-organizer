package listpersonas_test

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
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/listpersonas"
)

type failingRepo struct {
	port.PersonaRepository
	err error
}

func (r failingRepo) ListPersonas(context.Context, domain.CollectorID) ([]*domain.Persona, error) {
	return nil, r.err
}

func TestExecute_Success(t *testing.T) {
	repo := fakepersona.NewRepository("k7m2", "bbbb")
	ctx := context.Background()
	employed := &domain.Persona{ID: uuid.New(), CollectorID: "k7m2", Name: "Employed", Description: "Has a job"}
	letting := &domain.Persona{ID: uuid.New(), CollectorID: "k7m2", Name: "Letting"}
	other := &domain.Persona{ID: uuid.New(), CollectorID: "bbbb", Name: "Other"}

	for _, p := range []*domain.Persona{letting, other, employed} {
		if err := repo.SavePersona(ctx, p); err != nil {
			t.Fatalf("seed persona: %v", err)
		}
	}

	got, err := listpersonas.New(repo).Execute(tenant.WithCollector(ctx, "k7m2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff([]*domain.Persona{employed, letting}, got); diff != "" {
		t.Errorf("personas mismatch (-want +got):\n%s", diff)
	}
}

func TestExecute_UnknownCollector(t *testing.T) {
	got, err := listpersonas.New(fakepersona.NewRepository()).Execute(tenant.WithCollector(context.Background(), "k7m2"))
	if !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrCollectorNotFound)
	}

	if got != nil {
		t.Errorf("personas = %+v, want nil", got)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := listpersonas.New(failingRepo{err: boom}).Execute(tenant.WithCollector(context.Background(), "k7m2"))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want wrapped %v", err, boom)
	}
}

func TestExecute_NoTenant(t *testing.T) {
	_, err := listpersonas.New(fakepersona.NewRepository()).Execute(context.Background())
	if !errors.Is(err, tenant.ErrNoTenant) {
		t.Errorf("err = %v, want %v", err, tenant.ErrNoTenant)
	}
}
