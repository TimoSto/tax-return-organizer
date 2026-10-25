package domain

// ErrorKind classifies a domain error so adapters can map it without knowing every sentinel.
type ErrorKind int

const (
	// ErrorInvalid means the input is wrong.
	ErrorInvalid ErrorKind = iota + 1
	// ErrorNotFound means a referenced entity does not exist.
	ErrorNotFound
	// ErrorConflict means the request clashes with existing state.
	ErrorConflict
)

// Error is a domain error with a client-safe message, a stable machine-readable code and an ErrorKind.
// The code is part of the API contract: clients switch on it (and translate it), so never change one.
// Sentinels are *Error values, so errors.Is matches by identity.
type Error struct {
	Kind ErrorKind
	// Code is a stable snake_case identifier such as "collector_not_found".
	Code string
	msg  string
}

func (e *Error) Error() string { return e.msg }

// NewError returns a classified error for a sentinel that is declared outside the domain package.
func NewError(kind ErrorKind, code, msg string) *Error {
	return &Error{Kind: kind, Code: code, msg: msg}
}

func invalid(code, msg string) *Error  { return NewError(ErrorInvalid, code, msg) }
func notFound(code, msg string) *Error { return NewError(ErrorNotFound, code, msg) }
func conflict(code, msg string) *Error { return NewError(ErrorConflict, code, msg) }
