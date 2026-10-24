package createcollector_test

import (
	"context"
	"errors"
	"testing"

	fakecollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

// personaRecorder is a persona repository that only remembers what was saved.
type personaRecorder struct {
	ports.PersonaRepository
	saved []*domain.Persona
	err   error
}

func (r *personaRecorder) SavePersona(_ context.Context, p *domain.Persona) error {
	if r.err != nil {
		return r.err
	}

	r.saved = append(r.saved, p)

	return nil
}

func newUsecase(repo ports.CollectorRepository) *createcollector.Usecase {
	return createcollector.New(repo, &personaRecorder{})
}

type failingRepo struct {
	ports.CollectorRepository
	err error
}

func (r failingRepo) SaveCollector(context.Context, *domain.Collector) error { return r.err }

func TestExecute_Success(t *testing.T) {
	repo := fakecollector.NewRepository()

	id, err := newUsecase(repo).Execute(context.Background(), ports.CreateCollectorCmd{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetCollector(context.Background(), id)
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
		{"blank name", "  ", "jane@example.com", domain.ErrNameRequired},
		{"invalid email", "Jane", "not-an-email", domain.ErrInvalidEmail},
		{"empty email", "Jane", "", domain.ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := newUsecase(fakecollector.NewRepository()).Execute(context.Background(), ports.CreateCollectorCmd{Name: tt.cName, Email: tt.email})
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

	_, err := newUsecase(failingRepo{err: boom}).Execute(context.Background(), ports.CreateCollectorCmd{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}

type collidingRepo struct {
	ports.CollectorRepository
	collisions int
	calls      int
}

func (r *collidingRepo) SaveCollector(ctx context.Context, c *domain.Collector) error {
	r.calls++
	if r.calls <= r.collisions {
		return domain.ErrCollectorIDTaken
	}

	return r.CollectorRepository.SaveCollector(ctx, c)
}

func TestExecute_RetriesOnIDCollision(t *testing.T) {
	repo := &collidingRepo{CollectorRepository: fakecollector.NewRepository(), collisions: 2}

	id, err := newUsecase(repo).Execute(context.Background(), ports.CreateCollectorCmd{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.calls != 3 {
		t.Errorf("calls = %d, want 3", repo.calls)
	}

	if _, err := repo.GetCollector(context.Background(), id); err != nil {
		t.Errorf("collector not saved: %v", err)
	}
}

func TestExecute_GivesUpAfterTooManyCollisions(t *testing.T) {
	repo := &collidingRepo{CollectorRepository: fakecollector.NewRepository(), collisions: 100}

	id, err := newUsecase(repo).Execute(context.Background(), ports.CreateCollectorCmd{Name: "Jane", Email: "jane@example.com"})
	if err == nil {
		t.Fatal("expected an error")
	}

	if id != "" || repo.calls != 5 {
		t.Errorf("id = %q, calls = %d, want empty id and 5 calls", id, repo.calls)
	}
}

func TestExecute_AddsPredefinedPersonas(t *testing.T) {
	for locale, wantName := range map[string]string{"": "Employed", "en": "Employed", "de": "Angestellt"} {
		t.Run("locale "+locale, func(t *testing.T) {
			personas := &personaRecorder{}
			uc := createcollector.New(fakecollector.NewRepository(), personas)

			id, err := uc.Execute(context.Background(), ports.CreateCollectorCmd{Name: "Jane", Email: "jane@example.com", Locale: locale})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(personas.saved) != len(domain.PersonaTemplates()) {
				t.Fatalf("saved %d personas, want %d", len(personas.saved), len(domain.PersonaTemplates()))
			}

			if got := personas.saved[0]; got.Name != wantName || got.CollectorID != id {
				t.Errorf("first persona = %q of %q, want %q of %q", got.Name, got.CollectorID, wantName, id)
			}
		})
	}
}

func TestExecute_InvalidLocale(t *testing.T) {
	repo := fakecollector.NewRepository()

	id, err := newUsecase(repo).Execute(context.Background(), ports.CreateCollectorCmd{Name: "Jane", Email: "jane@example.com", Locale: "fr"})
	if !errors.Is(err, domain.ErrInvalidLocale) {
		t.Errorf("err = %v, want %v", err, domain.ErrInvalidLocale)
	}

	if id != "" {
		t.Errorf("id = %q, want empty", id)
	}
}

func TestExecute_PersonaFailureKeepsCollector(t *testing.T) {
	repo := fakecollector.NewRepository()
	uc := createcollector.New(repo, &personaRecorder{err: errors.New("boom")})

	id, err := uc.Execute(context.Background(), ports.CreateCollectorCmd{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.GetCollector(context.Background(), id); err != nil {
		t.Errorf("collector not saved: %v", err)
	}
}

func TestExecute_WithIDGenerator(t *testing.T) {
	repo := fakecollector.NewRepository()
	uc := createcollector.New(repo, &personaRecorder{}, createcollector.WithIDGenerator(func() (domain.CollectorID, error) { return "aaaa", nil }))

	id, err := uc.Execute(context.Background(), ports.CreateCollectorCmd{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "aaaa" {
		t.Errorf("id = %q, want aaaa", id)
	}
}
