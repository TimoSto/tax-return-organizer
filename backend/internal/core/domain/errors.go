package domain

import "errors"

// Sentinel validation errors returned by the domain constructors and
// mutators below. Callers can match on them with errors.Is, even when the
// returned error wraps one with additional context (e.g. via %w).
var (
	ErrEmptyCategoryName   = errors.New("category name must not be empty")
	ErrEmptyChecklistTitle = errors.New("checklist template title must not be empty")
	ErrInvalidRecurrence   = errors.New("invalid recurrence")

	ErrRecurrenceMismatch = errors.New("checklist template recurrence does not match the requested item generation")
	ErrInvalidMonth       = errors.New("month must be between 1 and 12")

	ErrMissingRequirement   = errors.New("checklist template requires a requirement")
	ErrWrongFulfillmentKind = errors.New("checklist item cannot be satisfied this way")
	ErrInvalidValue         = errors.New("invalid value")

	ErrTaxYearOutOfRange = errors.New("tax year out of plausible range")

	ErrEmptyDocumentFilename = errors.New("document filename must not be empty")
	ErrEmptyDocumentMimeType = errors.New("document mime type must not be empty")
	ErrEmptyDocumentContent  = errors.New("document content must not be empty")

	ErrEmptyCollectorName = errors.New("collector name must not be empty")
	ErrEmptyClientName    = errors.New("client name must not be empty")
)
