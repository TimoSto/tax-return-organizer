package ports

import "context"

// returns the collector id if successful
type CreateCollectorPort func(ctx context.Context, name string, email string) (string, error)
