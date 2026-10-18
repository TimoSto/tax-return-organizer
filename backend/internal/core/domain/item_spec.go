package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidValueType is returned for a value type that is not one of the known ones.
	ErrInvalidValueType = errors.New("invalid value type")
	// ErrMetadataFieldNameRequired is returned for a metadata field with a blank name.
	ErrMetadataFieldNameRequired = errors.New("metadata field name is required")
	// ErrDuplicateMetadataField is returned when two metadata fields of one document spec share a name.
	ErrDuplicateMetadataField = errors.New("duplicate metadata field name")
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
	Name string
	Spec ValueSpec
}

// DocumentSpec describes what is asked for a document item.
type DocumentSpec struct {
	MetadataFields []MetadataField
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

		if err := f.Spec.Validate(); err != nil {
			return fmt.Errorf("metadata field %q: %w", name, err)
		}
	}

	return nil
}
