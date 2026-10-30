package domain

import "sync"

// ErrorKind classifies a domain error so adapters can map it without knowing every sentinel.
type ErrorKind int

const (
	// ErrorInvalid means the input is wrong.
	ErrorInvalid ErrorKind = iota + 1
	// ErrorNotFound means a referenced entity does not exist.
	ErrorNotFound
	// ErrorConflict means the request clashes with existing state.
	ErrorConflict
	// ErrorForbidden means the actor may not do this.
	ErrorForbidden
)

// Error is a domain error: a business rule that was violated, with a client-safe English message and an ErrorKind.
// How a client is told about it (status, stable code) is up to the inbound adapter, which maps the sentinels.
// Sentinels are *Error values, so errors.Is matches by identity.
type Error struct {
	Kind ErrorKind
	msg  string
}

func (e *Error) Error() string { return e.msg }

var (
	registryMu sync.Mutex
	registry   []*Error
)

// NewError returns a classified error for a sentinel. Declare sentinels as package level variables; this is also
// how sentinels outside the domain package are made. Every error made here is listed by Errors.
func NewError(kind ErrorKind, msg string) *Error {
	e := &Error{Kind: kind, msg: msg}

	registryMu.Lock()
	defer registryMu.Unlock()

	registry = append(registry, e)

	return e
}

// Errors returns all errors made with NewError so far, i.e. all sentinels of the packages loaded. It exists so that
// tests can check that adapters handle every sentinel; production code has no use for it. Only a test binary that
// links the whole production graph (internal/backend) sees them all.
func Errors() []*Error {
	registryMu.Lock()
	defer registryMu.Unlock()

	return append([]*Error(nil), registry...)
}

func invalid(msg string) *Error  { return NewError(ErrorInvalid, msg) }
func notFound(msg string) *Error { return NewError(ErrorNotFound, msg) }
func conflict(msg string) *Error { return NewError(ErrorConflict, msg) }

func forbidden(msg string) *Error { return NewError(ErrorForbidden, msg) }
