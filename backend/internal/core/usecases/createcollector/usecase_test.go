package createcollector_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

type failingRepo struct {
	ports.CollectorRepository
	err error
}

func (r failingRepo) SaveCollector(context.Context, *domain.Collector) error { return r.err }

func TestExecute_Success(t *testing.T) {
	repo := fake.NewCollectorRepository()

	id, err := createcollector.New(repo).Execute(context.Background(), "Jane", "jane@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("id %q is not a uuid: %v", id, err)
	}

	got, err := repo.GetCollector(context.Background(), uid)
	if err != nil {
		t.Fatalf("collector not saved: %v", err)
	}

	if got.Name != "Jane" || got.Email != "jane@example.com" {
		t.Errorf("unexpected collector: %+v", got)
	}
}

func TestExecute_Validation(t *testing.T) {
	tests := []struct {
		name    string
		cName   string
		email   string
		wantErr error
	}{
		{"blank name", "  ", "jane@example.com", createcollector.ErrNameRequired},
		{"invalid email", "Jane", "not-an-email", domain.ErrInvalidEmail},
		{"empty email", "Jane", "", domain.ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := createcollector.New(fake.NewCollectorRepository()).Execute(context.Background(), tt.cName, tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}

			if id != "" {
				t.Errorf("id = %q, want empty", id)
			}
		})
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := createcollector.New(failingRepo{err: boom}).Execute(context.Background(), "Jane", "jane@example.com")
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}
