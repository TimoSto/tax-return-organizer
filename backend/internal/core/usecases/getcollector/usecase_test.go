package getcollector_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	fakecollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

type failingRepo struct {
	ports.CollectorRepository
	err error
}

func (r failingRepo) GetCollector(context.Context, uuid.UUID) (*domain.Collector, error) {
	return nil, r.err
}

func TestExecute_Success(t *testing.T) {
	repo := fakecollector.NewRepository()
	want := &domain.Collector{ID: uuid.New(), Name: "Jane", Email: "jane@example.com"}

	if err := repo.SaveCollector(context.Background(), want); err != nil {
		t.Fatalf("seed collector: %v", err)
	}

	got, err := getcollector.New(repo).Execute(context.Background(), want.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *got != *want {
		t.Errorf("collector = %+v, want %+v", got, want)
	}
}

func TestExecute_NotFound(t *testing.T) {
	got, err := getcollector.New(fakecollector.NewRepository()).Execute(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrCollectorNotFound)
	}

	if got != nil {
		t.Errorf("collector = %+v, want nil", got)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := getcollector.New(failingRepo{err: boom}).Execute(context.Background(), uuid.New())
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}
