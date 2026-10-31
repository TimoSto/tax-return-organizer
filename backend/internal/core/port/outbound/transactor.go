package outbound

import "context"

// Transactor is how a use case makes several repository calls atomic. The use case owns the transaction: it decides
// where it starts and ends, and repositories never open one themselves. A single repository call is atomic on its
// own (SavePersona writes the persona and its items all or nothing), so a use case needs InTx only when it writes
// through more than one call, for example a collector and its personas.
//
// R is the set of repositories fn receives. Each use case declares its own R with exactly the repositories it
// writes through, so a new repository reaches no use case that does not ask for it; the composition root builds
// R on the transaction.
//
// Transactor runs fn in a transaction. All repositories fn receives share it. The transaction commits if fn
// returns nil, otherwise it is rolled back and fn's error is returned.
type Transactor[R any] interface {
	InTx(ctx context.Context, fn func(ctx context.Context, repos R) error) error
}
