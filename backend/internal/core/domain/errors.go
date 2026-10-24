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

// Error is a domain error with a client-safe message and an ErrorKind.
// Sentinels are *Error values, so errors.Is matches by identity.
type Error struct {
	Kind ErrorKind
	msg  string
}

func (e *Error) Error() string { return e.msg }

func invalid(msg string) *Error  { return &Error{Kind: ErrorInvalid, msg: msg} }
func notFound(msg string) *Error { return &Error{Kind: ErrorNotFound, msg: msg} }
func conflict(msg string) *Error { return &Error{Kind: ErrorConflict, msg: msg} }
