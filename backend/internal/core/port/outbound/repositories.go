package outbound

import (
	"context"
	"errors"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// TaxPracticeRepository persists tax practices. It is deliberately not tenant-scoped: the tax practice is the tenant itself,
// so it is looked up and created by id. Everything a tax practice owns is reached through a store (see PersonaStore).
type TaxPracticeRepository interface {
	// GetTaxPractice returns domain.ErrTaxPracticeNotFound if no tax practice with the given id exists.
	GetTaxPractice(ctx context.Context, id domain.TaxPracticeID) (*domain.TaxPractice, error)
	// SaveTaxPractice inserts a new tax practice. It returns domain.ErrTaxPracticeIDTaken if the id is already in use.
	SaveTaxPractice(ctx context.Context, taxPractice *domain.TaxPractice) error
}

// ErrTenantMismatch is returned when a repository bound to one tax practice is asked to save data of another. It is
// deliberately not a domain error: it is a programming error (a use case picked the wrong repository), so it
// surfaces as a 500 instead of being shown to the client.
var ErrTenantMismatch = errors.New("entity belongs to another tax practice")

// ErrCorruptData is matched (errors.Is) by the errors CorruptData returns.
var ErrCorruptData = errors.New("corrupt stored data")

type corruptDataError struct{ what, cause string }

func (e *corruptDataError) Error() string { return "corrupt stored data: " + e.what + ": " + e.cause }

func (e *corruptDataError) Is(target error) bool { return target == ErrCorruptData }

// CorruptData reports that a stored entity (what) breaks a domain rule. It is for repositories rebuilding entities
// from stored rows. The cause is kept as text only and is deliberately not wrapped: a domain error inside it would
// otherwise be shown to the client as their mistake (400), but the data was stored by us, so it must surface as a 500.
func CorruptData(what string, cause error) error {
	return &corruptDataError{what: what, cause: cause.Error()}
}

// PersonaStore hands out persona repositories that are bound to one tax practice. Tenant isolation is structural: a
// use case can only reach the personas of the tax practice it asked For, there is no method that takes a tax practice id.
// A store is cheap and holds no per-tax-practice state, For returns a small view on the shared storage.
type PersonaStore interface {
	For(taxPracticeID domain.TaxPracticeID) PersonaRepository
}

// PersonaRepository persists the personas of one tax practice (see PersonaStore.For) together with their item
// definitions, which are part of the persona. Personas, and the ids of their items, of other tax practices are not
// visible through it.
//
// Referential integrity is the repository's job, not the caller's: every method returns domain.ErrTaxPracticeNotFound
// if the bound tax practice does not exist, so use cases need no check of their own. Postgres gets this from the
// foreign key (writes) and from reading through the tax practices table (reads).
type PersonaRepository interface {
	// GetPersona returns the persona with all its items in their stored order.
	// It returns domain.ErrTaxPracticeNotFound if the tax practice does not exist, otherwise domain.ErrPersonaNotFound
	// if the tax practice has no persona with the given id, which includes personas that belong to another tax practice.
	GetPersona(ctx context.Context, id domain.PersonaID) (*domain.Persona, error)
	// ListPersonas returns all personas of the tax practice, in no particular order, with their items in their stored
	// order. Sorting them for a reader is the use case's job.
	// A tax practice without personas yields an empty list, an unknown taxPractice domain.ErrTaxPracticeNotFound.
	ListPersonas(ctx context.Context) ([]*domain.Persona, error)
	// SavePersona inserts the persona or, if its id exists already, replaces name, description and items.
	// Items missing from persona.Items are deleted, the others keep their id. Item ids are scoped to their
	// persona: another persona, of this or any other tax practice, may use the same item id.
	// It returns ErrTenantMismatch if the persona belongs to another tax practice than the repository,
	// domain.ErrTaxPracticeNotFound if the tax practice does not exist (even if the id also belongs to another one),
	// or domain.ErrPersonaNotFound if the id belongs to another tax practice. Nothing is changed then: the save is
	// atomic on its own, without the caller wrapping it in a transaction.
	//
	// There is no concurrency control on purpose: SavePersona replaces the whole persona, so the last save
	// wins, and a stale save also deletes the items another employee added meanwhile. Concurrent edits of one
	// persona are rare and accepted. Client data never depends on a persona: a tax year copies the item
	// definitions when it is created and refers to them only softly, so no save can lose client data (see
	// "Personas have no optimistic locking" in docs/ARCHITECTURE.md).
	SavePersona(ctx context.Context, persona *domain.Persona) error
}
