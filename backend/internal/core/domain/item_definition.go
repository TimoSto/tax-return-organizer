package domain

import (
	"fmt"
	"strings"
)

var (
	// ErrItemTitleRequired is returned for a blank item title.
	ErrItemTitleRequired = invalid("item title is required")
	// ErrInvalidRecurrence is returned for a recurrence that is not one of the known ones.
	ErrInvalidRecurrence = invalid("invalid recurrence")
	// ErrSpecRequired is returned for an item without a spec.
	ErrSpecRequired = invalid("item spec is required")
	// ErrSpecMismatch is returned when an update gives an item a spec of another kind.
	ErrSpecMismatch = invalid("spec does not match item kind")
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

// ItemDefinition is something a persona asks the client to provide. What exactly is asked for, and with it the
// kind of the item, is given by its Spec.
type ItemDefinition struct {
	ID         ItemID
	Title      string
	Recurrence Recurrence
	Spec       ItemSpec
}

// Kind tells whether the item is satisfied by uploaded documents or by an entered value.
func (i ItemDefinition) Kind() Kind { return i.Spec.Kind() }

// NewItemDefinition validates the input and returns an item definition with the given id.
func NewItemDefinition(id ItemID, title string, recurrence Recurrence, spec ItemSpec) (*ItemDefinition, error) {
	return buildItemDefinition(id, title, recurrence, spec)
}

// buildItemDefinition validates the input and returns an item definition with the given id.
func buildItemDefinition(id ItemID, title string, recurrence Recurrence, spec ItemSpec) (*ItemDefinition, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrItemTitleRequired
	}

	if !recurrence.Valid() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidRecurrence, recurrence)
	}

	if spec == nil {
		return nil, ErrSpecRequired
	}

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	return &ItemDefinition{ID: id, Title: title, Recurrence: recurrence, Spec: spec}, nil
}
