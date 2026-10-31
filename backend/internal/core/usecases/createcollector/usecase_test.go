package createcollector_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	fakecollector "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/collector"
	faketransactor "github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/fake/transactor"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/locale"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
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

func newTransactor(repo outbound.CollectorRepository, personas outbound.PersonaStore) *faketransactor.Transactor[createcollector.Repositories] {
	return faketransactor.New(createcollector.Repositories{Collectors: repo, Personas: personas})
}

func newUsecase(repo outbound.CollectorRepository) *createcollector.Usecase {
	return createcollector.New(newTransactor(repo, &personaRecorder{}), personatemplate.Predefined(), randomIDs{})
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

			if id != (domain.CollectorID{}) {
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

func TestExecute_TakenIDFails(t *testing.T) {
	repo := fakecollector.NewRepository()
	fixed := domaintest.CollectorID()
	if err := repo.SaveCollector(context.Background(), domaintest.Collector(fixed, "Taken", "taken@example.com")); err != nil {
		t.Fatalf("seed: %v", err)
	}

	tx := newTransactor(repo, &personaRecorder{})
	uc := createcollector.New(tx, personatemplate.Predefined(), fixedCollectorID{id: fixed})

	id, err := uc.Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, domain.ErrCollectorIDTaken) {
		t.Errorf("err = %v, want it to wrap %v", err, domain.ErrCollectorIDTaken)
	}

	if id != (domain.CollectorID{}) || tx.Calls != 1 {
		t.Errorf("id = %v, transactions = %d, want the zero id and one transaction (no retry)", id, tx.Calls)
	}
}

func TestExecute_AddsPredefinedPersonas(t *testing.T) {
	for locale, wantName := range map[string]string{"": "Employed", "en": "Employed", "de": "Angestellt"} {
		t.Run("locale "+locale, func(t *testing.T) {
			personas := &personaRecorder{}
			uc := createcollector.New(newTransactor(fakecollector.NewRepository(), personas), personatemplate.Predefined(), randomIDs{})

			id, err := uc.Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com", Locale: locale})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(personas.saved) != len(personatemplate.Predefined()) {
				t.Fatalf("saved %d personas, want %d", len(personas.saved), len(personatemplate.Predefined()))
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
	if !errors.Is(err, locale.ErrInvalidLocale) {
		t.Errorf("err = %v, want %v", err, locale.ErrInvalidLocale)
	}

	if id != (domain.CollectorID{}) {
		t.Errorf("id = %q, want empty", id)
	}
}

func TestExecute_PersonaFailureFails(t *testing.T) {
	boom := errors.New("boom")
	uc := createcollector.New(newTransactor(fakecollector.NewRepository(), &personaRecorder{err: boom}), personatemplate.Predefined(), randomIDs{})

	id, err := uc.Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}

	if id != (domain.CollectorID{}) {
		t.Errorf("id = %q, want empty", id)
	}
}

func TestExecute_WithIDGenerator(t *testing.T) {
	repo := fakecollector.NewRepository()
	want := domaintest.CollectorID()
	uc := createcollector.New(newTransactor(repo, &personaRecorder{}), personatemplate.Predefined(), fixedCollectorID{id: want})

	id, err := uc.Execute(context.Background(), inbound.CreateCollectorCommand{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != want {
		t.Errorf("id = %v, want %v", id, want)
	}
}

// randomIDs is an outbound.IDGenerator that draws random ids.
type randomIDs struct{}

func (randomIDs) NewCollectorID() (domain.CollectorID, error) {
	return domain.CollectorID(uuid.New()), nil
}

func (randomIDs) NewPersonaID() (domain.PersonaID, error) { return domain.PersonaID(uuid.New()), nil }

func (randomIDs) NewItemID() (domain.ItemID, error) { return domain.ItemID(uuid.New()), nil }

// fixedCollectorID draws the collector id id and random ids otherwise.
type fixedCollectorID struct {
	randomIDs
	id domain.CollectorID
}

func (f fixedCollectorID) NewCollectorID() (domain.CollectorID, error) { return f.id, nil }
