package domain

import (
	"fmt"
	"strings"
)

var (
	// ErrInvalidValueType is returned for a value type that is not one of the known ones.
	ErrInvalidValueType = invalid("invalid value type")
	// ErrMetadataFieldNameRequired is returned for a metadata field with a blank name.
	ErrMetadataFieldNameRequired = invalid("metadata field name is required")
	// ErrDuplicateMetadataField is returned when two metadata fields of one document spec share a name.
	ErrDuplicateMetadataField = invalid("duplicate metadata field name")
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

// ValueSpec describes a single value the client has to provide.
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

// MetadataField is a value the client enters for an uploaded document.
type MetadataField struct {
	Name      string
	ValueSpec ValueSpec
}

// DocumentSpec describes what is asked for a document item.
type DocumentSpec struct {
	// MetadataFields is nil, never empty, if the document has none: item definitions and repositories
	// normalise it, so values compare equal after a round trip.
	MetadataFields []MetadataField
}

// normalized returns a copy whose empty MetadataFields is nil.
func (s DocumentSpec) normalized() DocumentSpec {
	if len(s.MetadataFields) == 0 {
		s.MetadataFields = nil
	}

	return s
}

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
