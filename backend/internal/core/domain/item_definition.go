package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	// ErrItemTitleRequired is returned for a blank item title.
	ErrItemTitleRequired = errors.New("item title is required")
	// ErrInvalidKind is returned for an item kind that is not one of the known ones.
	ErrInvalidKind = errors.New("invalid item kind")
	// ErrInvalidRecurrence is returned for a recurrence that is not one of the known ones.
	ErrInvalidRecurrence = errors.New("invalid recurrence")
	// ErrSpecMismatch is returned when an item does not carry exactly the spec that its kind requires.
	ErrSpecMismatch = errors.New("spec does not match item kind")
)

// Kind tells whether an item is satisfied by uploaded documents or by an entered value.
type Kind string

// The known item kinds.
const (
	KindDocument Kind = "document"
	KindData     Kind = "data"
)

// Valid reports whether k is a known kind.
func (k Kind) Valid() bool { return k == KindDocument || k == KindData }

// Recurrence tells how often an item has to be provided within a tax year.
type Recurrence string

// The known recurrences.
const (
	RecurrenceMonthly      Recurrence = "monthly"
	RecurrenceQuarterly    Recurrence = "quarterly"
	RecurrenceSemiannually Recurrence = "semiannually"
	RecurrenceAnnually     Recurrence = "annually"
)

// Valid reports whether r is a known recurrence.
func (r Recurrence) Valid() bool { return r.Periods() > 0 }

// Periods returns how many times per year the item is due, or 0 for an unknown recurrence.
func (r Recurrence) Periods() int {
	switch r {
	case RecurrenceMonthly:
		return 12
	case RecurrenceQuarterly:
		return 4
	case RecurrenceSemiannually:
		return 2
	case RecurrenceAnnually:
		return 1
	default:
		return 0
	}
}

// ItemDefinition is something a persona asks the client to provide.
// A document item carries a DocumentSpec, a data item a ValueSpec, never both.
type ItemDefinition struct {
	ID           uuid.UUID
	PersonaID    uuid.UUID
	Title        string
	Kind         Kind
	Recurrence   Recurrence
	DocumentSpec *DocumentSpec
	ValueSpec    *ValueSpec
}

// NewItemDefinition validates the input and returns an item definition with a fresh id.
// The spec matching kind must be set, the other one must be nil.
func NewItemDefinition(
	personaID uuid.UUID,
	title string,
	kind Kind,
	recurrence Recurrence,
	documentSpec *DocumentSpec,
	dataSpec *ValueSpec,
) (*ItemDefinition, error) {
	return buildItemDefinition(uuid.New(), personaID, title, kind, recurrence, documentSpec, dataSpec)
}

// buildItemDefinition validates the input and returns an item definition with the given id.
func buildItemDefinition(
	id, personaID uuid.UUID,
	title string,
	kind Kind,
	recurrence Recurrence,
	documentSpec *DocumentSpec,
	dataSpec *ValueSpec,
) (*ItemDefinition, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrItemTitleRequired
	}

	if !kind.Valid() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidKind, kind)
	}

	if !recurrence.Valid() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidRecurrence, recurrence)
	}

	switch kind {
	case KindDocument:
		if documentSpec == nil || dataSpec != nil {
			return nil, fmt.Errorf("%w: a document item needs a document spec and no data spec", ErrSpecMismatch)
		}

		if err := documentSpec.Validate(); err != nil {
			return nil, err
		}
	case KindData:
		if dataSpec == nil || documentSpec != nil {
			return nil, fmt.Errorf("%w: a data item needs a data spec and no document spec", ErrSpecMismatch)
		}

		if err := dataSpec.Validate(); err != nil {
			return nil, err
		}
	}

	return &ItemDefinition{
		ID:           id,
		PersonaID:    personaID,
		Title:        title,
		Kind:         kind,
		Recurrence:   recurrence,
		DocumentSpec: documentSpec,
		ValueSpec:    dataSpec,
	}, nil
}
