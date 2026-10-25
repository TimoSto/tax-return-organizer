package domain

import (
	"fmt"
	"slices"
	"strings"
)

var (
	// ErrItemTitleRequired is returned for a blank item title.
	ErrItemTitleRequired = invalid("item_title_required", "item title is required")
	// ErrInvalidRecurrence is returned for a recurrence that is not one of the known ones.
	ErrInvalidRecurrence = invalid("invalid_recurrence", "invalid recurrence")
	// ErrSpecRequired is returned for an item without a spec.
	ErrSpecRequired = invalid("item_spec_required", "item spec is required")
	// ErrSpecMismatch is returned when an update gives an item a spec of another kind.
	ErrSpecMismatch = invalid("item_spec_mismatch", "spec does not match item kind")
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
// kind of the item, is given by its spec. Instances are immutable and always valid; create them with
// NewItemDefinition or RehydrateItemDefinition.
type ItemDefinition struct {
	id         ItemID
	title      string
	recurrence Recurrence
	spec       ItemSpec
}

// NewItemDefinition validates the input and returns an item definition with the given id.
func NewItemDefinition(id ItemID, title string, recurrence Recurrence, spec ItemSpec) (ItemDefinition, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return ItemDefinition{}, ErrItemTitleRequired
	}

	if !recurrence.Valid() {
		return ItemDefinition{}, fmt.Errorf("%w: %q", ErrInvalidRecurrence, recurrence)
	}

	if spec == nil {
		return ItemDefinition{}, ErrSpecRequired
	}

	if err := spec.Validate(); err != nil {
		return ItemDefinition{}, err
	}

	return ItemDefinition{id: id, title: title, recurrence: recurrence, spec: cloneSpec(spec)}, nil
}

// RehydrateItemDefinition rebuilds an item definition that was stored earlier, e.g. when a repository loads it.
// It applies the same rules as NewItemDefinition, so corrupt stored data is reported instead of becoming an invalid
// item.
func RehydrateItemDefinition(id ItemID, title string, recurrence Recurrence, spec ItemSpec) (ItemDefinition, error) {
	return NewItemDefinition(id, title, recurrence, spec)
}

// ID returns the id of the item.
func (i ItemDefinition) ID() ItemID { return i.id }

// Title returns the title of the item.
func (i ItemDefinition) Title() string { return i.title }

// Recurrence returns how often the item is due within a tax year.
func (i ItemDefinition) Recurrence() Recurrence { return i.recurrence }

// Spec returns what the item asks for. The result is a copy: changing it does not change the item.
func (i ItemDefinition) Spec() ItemSpec { return cloneSpec(i.spec) }

// Kind tells whether the item is satisfied by uploaded documents or by an entered value.
// It is empty for the zero ItemDefinition, which is not a valid item.
func (i ItemDefinition) Kind() Kind {
	if i.spec == nil {
		return ""
	}

	return i.spec.Kind()
}

// cloneSpec copies the parts of a spec that reference memory, so specs are never shared between items and callers.
func cloneSpec(s ItemSpec) ItemSpec {
	if d, ok := s.(DocumentSpec); ok {
		d.MetadataFields = slices.Clone(d.MetadataFields)

		return d
	}

	return s
}
