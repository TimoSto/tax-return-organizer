package createcollector_test

import (
	"context"
	"errors"
	"testing"

	fakecollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	faketransactor "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/transactor"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/idgen"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/templates"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
)

// personaRecorder is a persona store whose repositories only remember what was saved.
type personaRecorder struct {
	saved []*domain.Persona
	err   error
}

func (r *personaRecorder) For(domain.CollectorID) outbound.PersonaRepository { return r }

func (r *personaRecorder) GetPersona(context.Context, domain.PersonaID) (*domain.Persona, error) {
	return nil, domain.ErrPersonaNotFound
}

func (r *personaRecorder) ListPersonas(context.Context) ([]*domain.Persona, error) { return nil, nil }

func (r *personaRecorder) SavePersona(_ context.Context, p *domain.Persona) error {
	if r.err != nil {
		return r.err
	}

	r.saved = append(r.saved, p)

	return nil
}

func newTransactor(repo outbound.CollectorRepository, personas outbound.PersonaStore) *faketransactor.Transactor {
	return faketransactor.New(outbound.Repositories{Collectors: repo, Personas: personas})
}

func newUsecase(repo outbound.CollectorRepository) *createcollector.Usecase {
	return createcollector.New(newTransactor(repo, &personaRecorder{}), templates.New(), idgen.Random{})
}

type failingRepo struct {
	outbound.CollectorRepository
	err error
}

func (r failingRepo) SaveCollector(context.Context, *domain.Collector) error { return r.err }

func TestExecute_Success(t *testing.T) {
	repo := fakecollector.NewRepository()

	id, err := newUsecase(repo).Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetCollector(context.Background(), id)
	if err != nil {
		t.Fatalf("collector not saved: %v", err)
	}

	if got.Name() != "Jane" || got.Email() != "jane@example.com" {
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
			id, err := newUsecase(fakecollector.NewRepository()).Execute(context.Background(), inbound.CreateCollectorCommand{Name: tt.cName, Email: tt.email})
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

	_, err := newUsecase(failingRepo{err: boom}).Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}

type collidingRepo struct {
	outbound.CollectorRepository
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

	id, err := newUsecase(repo).Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
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

	id, err := newUsecase(repo).Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
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
			uc := createcollector.New(newTransactor(fakecollector.NewRepository(), personas), templates.New(), idgen.Random{})

			id, err := uc.Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com", Locale: locale})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(personas.saved) != len(templates.New().PersonaTemplates()) {
				t.Fatalf("saved %d personas, want %d", len(personas.saved), len(templates.New().PersonaTemplates()))
			}

			if got := personas.saved[0]; got.Name() != wantName || got.CollectorID() != id {
				t.Errorf("first persona = %q of %q, want %q of %q", got.Name(), got.CollectorID(), wantName, id)
			}
		})
	}
}

func TestExecute_InvalidLocale(t *testing.T) {
	repo := fakecollector.NewRepository()

	id, err := newUsecase(repo).Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com", Locale: "fr"})
	if !errors.Is(err, personatemplate.ErrInvalidLocale) {
		t.Errorf("err = %v, want %v", err, personatemplate.ErrInvalidLocale)
	}

	if id != "" {
		t.Errorf("id = %q, want empty", id)
	}
}

func TestExecute_PersonaFailureFails(t *testing.T) {
	boom := errors.New("boom")
	uc := createcollector.New(newTransactor(fakecollector.NewRepository(), &personaRecorder{err: boom}), templates.New(), idgen.Random{})

	id, err := uc.Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}

	if id != "" {
		t.Errorf("id = %q, want empty", id)
	}
}

func TestExecute_RetriesInFreshTransaction(t *testing.T) {
	repo := &collidingRepo{CollectorRepository: fakecollector.NewRepository(), collisions: 2}
	tx := newTransactor(repo, &personaRecorder{})

	if _, err := createcollector.New(tx, templates.New(), idgen.Random{}).Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tx.Calls != 3 {
		t.Errorf("transactions = %d, want one per attempt (3)", tx.Calls)
	}
}

func TestExecute_WithIDGenerator(t *testing.T) {
	repo := fakecollector.NewRepository()
	uc := createcollector.New(newTransactor(repo, &personaRecorder{}), templates.New(), fixedCollectorID{})

	id, err := uc.Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "aaaa" {
		t.Errorf("id = %q, want aaaa", id)
	}
}

// fixedCollectorID draws the collector id aaaa and random ids otherwise.
type fixedCollectorID struct{ idgen.Random }

func (fixedCollectorID) NewCollectorID() (domain.CollectorID, error) { return "aaaa", nil }
