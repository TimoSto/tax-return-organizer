package outbound

import "context"

// Repositories groups the repositories that can take part in one transaction.
// Resources owned by a collector are added as stores, never as plain repositories.
type Repositories struct {
	Collectors CollectorRepository
	Personas   PersonaStore
}

// Transactor runs fn in a transaction. All repositories fn receives share it. The transaction commits if fn
// returns nil, otherwise it is rolled back and fn's error is returned.
type Transactor interface {
	InTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}
