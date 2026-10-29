package outbound

import (
	"context"
	"errors"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// CollectorRepository persists collectors. It is deliberately not tenant-scoped: the collector is the tenant itself,
// so it is looked up and created by id. Everything a collector owns is reached through a store (see PersonaStore).
type CollectorRepository interface {
	// GetCollector returns domain.ErrCollectorNotFound if no collector with the given id exists.
	GetCollector(ctx context.Context, id domain.CollectorID) (*domain.Collector, error)
	// SaveCollector inserts a new collector. It returns domain.ErrCollectorIDTaken if the id is already in use.
	SaveCollector(ctx context.Context, collector *domain.Collector) error
}

// ErrTenantMismatch is returned when a repository bound to one collector is asked to save data of another. It is
// deliberately not a domain error: it is a programming error (a use case picked the wrong repository), so it
// surfaces as a 500 instead of being shown to the client.
var ErrTenantMismatch = errors.New("entity belongs to another collector")

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

// PersonaStore hands out persona repositories that are bound to one collector. Tenant isolation is structural: a
// use case can only reach the personas of the collector it asked For, there is no method that takes a collector id.
// A store is cheap and holds no per-collector state, For returns a small view on the shared storage.
type PersonaStore interface {
	For(collectorID domain.CollectorID) PersonaRepository
}

// PersonaRepository persists the personas of one collector (see PersonaStore.For) together with their item
// definitions, which are part of the persona. Personas, and the ids of their items, of other collectors are not
// visible through it.
//
// Referential integrity is the repository's job, not the caller's: every method returns domain.ErrCollectorNotFound
// if the bound collector does not exist, so use cases need no check of their own. Postgres gets this from the
// foreign key (writes) and from reading through the collectors table (reads), the in-memory fake emulates it.
type PersonaRepository interface {
	// GetPersona returns the persona with all its items in their stored order.
	// It returns domain.ErrCollectorNotFound if the collector does not exist, otherwise domain.ErrPersonaNotFound
	// if the collector has no persona with the given id, which includes personas that belong to another collector.
	GetPersona(ctx context.Context, id domain.PersonaID) (*domain.Persona, error)
	// ListPersonas returns all personas of the collector with their items, ordered by name and then id.
	// A collector without personas yields an empty list, an unknown collector domain.ErrCollectorNotFound.
	ListPersonas(ctx context.Context) ([]*domain.Persona, error)
	// SavePersona inserts the persona or, if its id exists already, replaces name, description and items.
	// Items missing from persona.Items are deleted, the others keep their id. Item ids are scoped to their
	// persona: another persona, of this or any other collector, may use the same item id.
	// It returns ErrTenantMismatch if the persona belongs to another collector than the repository,
	// domain.ErrCollectorNotFound if the collector does not exist (even if the id also belongs to another one),
	// or domain.ErrPersonaNotFound if the id belongs to another collector. Nothing is changed then: the save is
	// atomic on its own, without the caller wrapping it in a transaction.
	//
	// There is no concurrency control: SavePersona replaces the whole persona, so the last save wins. Two
	// people editing the same persona from stale copies silently overwrite each other, and a stale save
	// also deletes the items the other person added meanwhile. That is acceptable while personas are edited
	// rarely by one person per collector. Before that changes, or before client data refers to item
	// definitions (deleting one would then lose that data), add optimistic locking: a persona version that
	// the client sends back, saved with "UPDATE ... WHERE version = $n" and answered with 409 on a mismatch.
	SavePersona(ctx context.Context, persona *domain.Persona) error
}
