package domain

import (
	"fmt"
	"slices"
	"strings"
)

var (
	// ErrItemTitleRequired is returned for a blank item title.
	ErrItemTitleRequired = invalid("item title is required")
	// ErrItemTitleTooLong is returned for an item title longer than MaxItemTitleLength.
	ErrItemTitleTooLong = invalid("item title is too long")
	// ErrInvalidRecurrence is returned for a recurrence that is not one of the known ones.
	ErrInvalidRecurrence = invalid("invalid recurrence")
	// ErrSpecRequired is returned for an item without a spec.
	ErrSpecRequired = invalid("item spec is required")
	// ErrSpecMismatch is returned when an update gives an item a spec of another kind.
	ErrSpecMismatch = invalid("spec does not match item kind")
	// ErrInstanceLabelNotRequired is returned for a repeatable item whose first instance field, the label, is not
	// required.
	ErrInstanceLabelNotRequired = invalid("the first instance field must be required")
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
//
// An item with instance fields is repeatable: the client provides it once per instance (e.g. a bank statement per
// bank account). The instances are set per tax year by the collector, who fills the instance fields for each of
// them. The first instance field labels the instance, so it must be required.
type ItemDefinition struct {
	id             ItemID
	title          string
	recurrence     Recurrence
	spec           ItemSpec
	instanceFields []MetadataField
}

// NewItemDefinition validates the input and returns an item definition with the given id. instanceFields may be
// nil or empty for an item that is not repeatable.
func NewItemDefinition(
	id ItemID, title string, recurrence Recurrence, spec ItemSpec, instanceFields []MetadataField,
) (ItemDefinition, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return ItemDefinition{}, ErrItemTitleRequired
	}

	if tooLong(title, MaxItemTitleLength) {
		return ItemDefinition{}, ErrItemTitleTooLong
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

	if err := validateInstanceFields(instanceFields); err != nil {
		return ItemDefinition{}, err
	}

	return ItemDefinition{
		id: id, title: title, recurrence: recurrence, spec: cloneSpec(spec), instanceFields: cloneFields(instanceFields),
	}, nil
}

// RehydrateItemDefinition rebuilds an item definition that was stored earlier, e.g. when a repository loads it.
// It applies the same rules as NewItemDefinition, so corrupt stored data is reported instead of becoming an invalid
// item.
//
// It currently just calls NewItemDefinition. It is kept as its own entry point for the upcoming UpdateItem use case,
// where rebuilding a stored item and creating a new one may follow different rules.
func RehydrateItemDefinition(
	id ItemID, title string, recurrence Recurrence, spec ItemSpec, instanceFields []MetadataField,
) (ItemDefinition, error) {
	return NewItemDefinition(id, title, recurrence, spec, instanceFields)
}

// ID returns the id of the item.
func (i ItemDefinition) ID() ItemID { return i.id }

// Title returns the title of the item.
func (i ItemDefinition) Title() string { return i.title }

// Recurrence returns how often the item is due within a tax year.
func (i ItemDefinition) Recurrence() Recurrence { return i.recurrence }

// Spec returns what the item asks for. The result is a copy: changing it does not change the item.
func (i ItemDefinition) Spec() ItemSpec { return cloneSpec(i.spec) }

// InstanceFields returns the fields the collector fills per instance of a repeatable item, the label first. It is
// nil for an item that is not repeatable. The result is a copy: changing it does not change the item.
func (i ItemDefinition) InstanceFields() []MetadataField { return cloneFields(i.instanceFields) }

// Repeatable reports whether the client provides the item once per instance.
func (i ItemDefinition) Repeatable() bool { return len(i.instanceFields) > 0 }

// Kind tells whether the item is satisfied by uploaded documents or by an entered value.
// It is empty for the zero ItemDefinition, which is not a valid item.
func (i ItemDefinition) Kind() Kind {
	if i.spec == nil {
		return ""
	}

	return i.spec.Kind()
}

// validateInstanceFields checks the instance fields like metadata fields, and that the label is required.
func validateInstanceFields(fields []MetadataField) error {
	if err := validateFields(fields); err != nil {
		return fmt.Errorf("instance fields: %w", err)
	}

	if len(fields) > 0 && !fields[0].ValueSpec.Required {
		return fmt.Errorf("%w: %q", ErrInstanceLabelNotRequired, strings.TrimSpace(fields[0].Name))
	}

	return nil
}

// cloneFields copies fields and normalizes an empty list to nil, so a single item has one representation.
func cloneFields(fields []MetadataField) []MetadataField {
	if len(fields) == 0 {
		return nil
	}

	return slices.Clone(fields)
}

// cloneSpec copies the parts of a spec that reference memory, so specs are never shared between items and callers.
func cloneSpec(s ItemSpec) ItemSpec {
	if d, ok := s.(DocumentSpec); ok {
		d.MetadataFields = slices.Clone(d.MetadataFields)

		return d
	}

	return s
}
