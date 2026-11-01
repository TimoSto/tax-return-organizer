package createtaxpractice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/locale"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/createtaxpractice"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

// taxPracticeRecorder is a tax practice repository that only remembers what was saved.
type taxPracticeRecorder struct {
	saved []*domain.TaxPractice
}

func (r *taxPracticeRecorder) GetTaxPractice(context.Context, domain.TaxPracticeID) (*domain.TaxPractice, error) {
	return nil, errors.New("taxPracticeRecorder: unexpected GetTaxPractice")
}

func (r *taxPracticeRecorder) SaveTaxPractice(_ context.Context, c *domain.TaxPractice) error {
	r.saved = append(r.saved, c)

	return nil
}

// personaRecorder is a persona store whose repositories only remember what was saved.
type personaRecorder struct {
	saved []*domain.Persona
	err   error
}

func (r *personaRecorder) For(domain.TaxPracticeID) outbound.PersonaRepository { return r }

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

// transactor runs the callback on fixed repositories without a transaction: rollback is covered by the Postgres
// transactor tests.
type transactor struct {
	repos createtaxpractice.Repositories
	// calls counts the InTx calls.
	calls int
}

func (t *transactor) InTx(ctx context.Context, fn func(ctx context.Context, repos createtaxpractice.Repositories) error) error {
	t.calls++

	return fn(ctx, t.repos)
}

func newTransactor(repo outbound.TaxPracticeRepository, personas outbound.PersonaStore) *transactor {
	return &transactor{repos: createtaxpractice.Repositories{TaxPractices: repo, Personas: personas}}
}

func newUsecase(repo outbound.TaxPracticeRepository) *createtaxpractice.Usecase {
	return createtaxpractice.New(newTransactor(repo, &personaRecorder{}), personatemplate.Predefined(), randomIDs{})
}

type failingRepo struct {
	outbound.TaxPracticeRepository
	err error
}

func (r failingRepo) SaveTaxPractice(context.Context, *domain.TaxPractice) error { return r.err }

func TestExecute_Success(t *testing.T) {
	repo := &taxPracticeRecorder{}

	id, err := newUsecase(repo).Execute(context.Background(), inbound.CreateTaxPracticeCommand{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("saved %d tax practices, want 1", len(repo.saved))
	}

	if got := repo.saved[0]; got.ID() != id || got.Name() != "Jane" || got.Email() != "jane@example.com" {
		t.Errorf("unexpected tax practice: %+v", got)
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
			id, err := newUsecase(&taxPracticeRecorder{}).Execute(context.Background(), inbound.CreateTaxPracticeCommand{Name: tt.cName, Email: tt.email})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}

			if id != (domain.TaxPracticeID{}) {
				t.Errorf("id = %q, want empty", id)
			}
		})
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := newUsecase(failingRepo{err: boom}).Execute(context.Background(), inbound.CreateTaxPracticeCommand{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}

func TestExecute_TakenIDFails(t *testing.T) {
	tx := newTransactor(failingRepo{err: domain.ErrTaxPracticeIDTaken}, &personaRecorder{})
	uc := createtaxpractice.New(tx, personatemplate.Predefined(), randomIDs{})

	id, err := uc.Execute(context.Background(), inbound.CreateTaxPracticeCommand{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, domain.ErrTaxPracticeIDTaken) {
		t.Errorf("err = %v, want it to wrap %v", err, domain.ErrTaxPracticeIDTaken)
	}

	if id != (domain.TaxPracticeID{}) || tx.calls != 1 {
		t.Errorf("id = %v, transactions = %d, want the zero id and one transaction (no retry)", id, tx.calls)
	}
}

func TestExecute_AddsPredefinedPersonas(t *testing.T) {
	for locale, wantName := range map[string]string{"": "Employed", "en": "Employed", "de": "Angestellt"} {
		t.Run("locale "+locale, func(t *testing.T) {
			personas := &personaRecorder{}
			uc := createtaxpractice.New(newTransactor(&taxPracticeRecorder{}, personas), personatemplate.Predefined(), randomIDs{})

			id, err := uc.Execute(context.Background(), inbound.CreateTaxPracticeCommand{Name: "Jane", Email: "jane@example.com", Locale: locale})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(personas.saved) != len(personatemplate.Predefined()) {
				t.Fatalf("saved %d personas, want %d", len(personas.saved), len(personatemplate.Predefined()))
			}

			if got := personas.saved[0]; got.Name() != wantName || got.TaxPracticeID() != id {
				t.Errorf("first persona = %q of %q, want %q of %q", got.Name(), got.TaxPracticeID(), wantName, id)
			}
		})
	}
}

func TestExecute_InvalidLocale(t *testing.T) {
	repo := &taxPracticeRecorder{}

	id, err := newUsecase(repo).Execute(context.Background(), inbound.CreateTaxPracticeCommand{Name: "Jane", Email: "jane@example.com", Locale: "fr"})
	if !errors.Is(err, locale.ErrInvalidLocale) {
		t.Errorf("err = %v, want %v", err, locale.ErrInvalidLocale)
	}

	if id != (domain.TaxPracticeID{}) {
		t.Errorf("id = %q, want empty", id)
	}
}

func TestExecute_PersonaFailureFails(t *testing.T) {
	boom := errors.New("boom")
	uc := createtaxpractice.New(newTransactor(&taxPracticeRecorder{}, &personaRecorder{err: boom}), personatemplate.Predefined(), randomIDs{})

	id, err := uc.Execute(context.Background(), inbound.CreateTaxPracticeCommand{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}

	if id != (domain.TaxPracticeID{}) {
		t.Errorf("id = %q, want empty", id)
	}
}

func TestExecute_WithIDGenerator(t *testing.T) {
	repo := &taxPracticeRecorder{}
	want := domaintest.TaxPracticeID()
	uc := createtaxpractice.New(newTransactor(repo, &personaRecorder{}), personatemplate.Predefined(), fixedTaxPracticeID{id: want})

	id, err := uc.Execute(context.Background(), inbound.CreateTaxPracticeCommand{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != want {
		t.Errorf("id = %v, want %v", id, want)
	}
}

// randomIDs is an outbound.IDGenerator that draws random ids.
type randomIDs struct{}

func (randomIDs) NewTaxPracticeID() (domain.TaxPracticeID, error) {
	return domain.TaxPracticeID(uuid.New()), nil
}

func (randomIDs) NewPersonaID() (domain.PersonaID, error) { return domain.PersonaID(uuid.New()), nil }

func (randomIDs) NewItemID() (domain.ItemID, error) { return domain.ItemID(uuid.New()), nil }

// fixedTaxPracticeID draws the tax practice id id and random ids otherwise.
type fixedTaxPracticeID struct {
	randomIDs
	id domain.TaxPracticeID
}

func (f fixedTaxPracticeID) NewTaxPracticeID() (domain.TaxPracticeID, error) { return f.id, nil }
