package tenant

import "errors"

// ErrNoTenant means a use case ran without a collector in its context. It is a programming error and deliberately
// not a domain error, so it surfaces as a 500 instead of being shown to the client.
var ErrNoTenant = errors.New("no collector in context")
