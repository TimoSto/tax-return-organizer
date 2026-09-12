package port

import "errors"

// ErrNotFound is returned by repository lookups that find nothing.
// Adapters should wrap it (fmt.Errorf("...: %w", ErrNotFound)) so callers
// can check with errors.Is(err, port.ErrNotFound).
var ErrNotFound = errors.New("not found")
