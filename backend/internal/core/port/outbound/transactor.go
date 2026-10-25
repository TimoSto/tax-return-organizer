package outbound

import "context"

// Repositories groups the repositories that can take part in one transaction.
// Resources owned by a collector are added as stores, never as plain repositories.
type Repositories struct {
	Collectors CollectorRepository
	Personas   PersonaStore
}

// Transactor is how a use case makes several repository calls atomic. The use case owns the transaction: it decides
// where it starts and ends, and repositories never open one themselves. A single repository call is atomic on its
// own (SavePersona writes the persona and its items all or nothing), so a use case needs InTx only when it writes
// through more than one call, for example a collector and its personas.
//
// Transactor runs fn in a transaction. All repositories fn receives share it. The transaction commits if fn
// returns nil, otherwise it is rolled back and fn's error is returned.
type Transactor interface {
	InTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}
