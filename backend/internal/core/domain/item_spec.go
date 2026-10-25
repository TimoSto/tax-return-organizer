package domain

import (
	"fmt"
	"strings"
)

var (
	// ErrInvalidValueType is returned for a value type that is not one of the known ones.
	ErrInvalidValueType = invalid("invalid_value_type", "invalid value type")
	// ErrMetadataFieldNameRequired is returned for a metadata field with a blank name.
	ErrMetadataFieldNameRequired = invalid("metadata_field_name_required", "metadata field name is required")
	// ErrDuplicateMetadataField is returned when two metadata fields of one document spec share a name.
	ErrDuplicateMetadataField = invalid("duplicate_metadata_field", "duplicate metadata field name")
)

// ValueType is the kind of value a data item or metadata field holds.
type ValueType string

// The known value types.
const (
	ValueTypeText   ValueType = "text"
	ValueTypeNumber ValueType = "number"
	ValueTypeDate   ValueType = "date"
	ValueTypeBool   ValueType = "bool"
)

// Valid reports whether t is a known value type.
func (t ValueType) Valid() bool {
	switch t {
	case ValueTypeText, ValueTypeNumber, ValueTypeDate, ValueTypeBool:
		return true
	default:
		return false
	}
}

// ItemSpec says what an item asks for. It is a sum type that only DocumentSpec (a document item) and ValueSpec
// (a data item) implement; use a type switch to tell them apart.
type ItemSpec interface {
	// Kind is the kind of item the spec belongs to.
	Kind() Kind
	// Validate checks that the spec is well-formed.
	Validate() error

	isItemSpec()
}

// ValueSpec describes a single value the client has to provide. As an ItemSpec it makes a data item.
type ValueSpec struct {
	Type ValueType
	// Unit is an optional display label (e.g. "EUR", "days"). It is never interpreted.
	Unit     string
	Required bool
}

// Validate checks that the spec has a known type.
func (s ValueSpec) Validate() error {
	if !s.Type.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidValueType, s.Type)
	}

	return nil
}

// Kind is KindData.
func (ValueSpec) Kind() Kind { return KindData }

func (ValueSpec) isItemSpec() {}

// MetadataField is a value the client enters for an uploaded document.
type MetadataField struct {
	Name      string
	ValueSpec ValueSpec
}

// DocumentSpec describes what is asked for a document item. As an ItemSpec it makes a document item.
type DocumentSpec struct {
	// MetadataFields may be nil or empty if the document has none.
	MetadataFields []MetadataField
}

// Kind is KindDocument.
func (DocumentSpec) Kind() Kind { return KindDocument }

func (DocumentSpec) isItemSpec() {}

// Validate checks that all metadata fields have a unique, non-blank name and a valid spec.
func (s DocumentSpec) Validate() error {
	seen := make(map[string]bool, len(s.MetadataFields))

	for _, f := range s.MetadataFields {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			return ErrMetadataFieldNameRequired
		}

		if seen[name] {
			return fmt.Errorf("%w: %q", ErrDuplicateMetadataField, name)
		}

		seen[name] = true

		if err := f.ValueSpec.Validate(); err != nil {
			return fmt.Errorf("metadata field %q: %w", name, err)
		}
	}

	return nil
}
