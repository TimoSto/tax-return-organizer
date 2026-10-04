package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Requirement is what a checklist template declares an item needs: a
// document or a typed value. It is sealed — only this package implements it.
//
// Accepts reports whether v satisfies the requirement: ErrWrongFulfillmentKind
// if v is the wrong kind, ErrInvalidValue if it's the right kind but breaks a
// rule (e.g. a year out of range). Turning raw input (form fields, JSON) into
// a Value is the inbound adapter's job, not the domain's.
type Requirement interface {
	Kind() ValueKind
	Accepts(v Value) error
	isRequirement()
}

type DocumentRequirement struct{}

type TextRequirement struct{}

// NumberRequirement asks for a number, with an optional display unit
// (e.g. "days", "EUR").
type NumberRequirement struct{ Unit string }

type YearRequirement struct{}

type BoolRequirement struct{}

func NewNumberRequirement(unit string) NumberRequirement {
	return NumberRequirement{Unit: strings.TrimSpace(unit)}
}

func (DocumentRequirement) Kind() ValueKind { return KindDocument }
func (TextRequirement) Kind() ValueKind     { return KindText }
func (NumberRequirement) Kind() ValueKind   { return KindNumber }
func (YearRequirement) Kind() ValueKind     { return KindYear }
func (BoolRequirement) Kind() ValueKind     { return KindBool }

func (DocumentRequirement) isRequirement() {}
func (TextRequirement) isRequirement()     {}
func (NumberRequirement) isRequirement()   {}
func (YearRequirement) isRequirement()     {}
func (BoolRequirement) isRequirement()     {}

// mismatch is the error every Accepts returns when v isn't the concrete value
// type the requirement expects. That covers a nil value, a value of another
// kind, and a pointer to the right type (only value types are accepted). It
// formats v with %T, never calling v.Kind(), which would panic on a nil pointer.
func mismatch(r Requirement, v Value) error {
	if v == nil {
		return fmt.Errorf("%w: value is nil", ErrInvalidValue)
	}
	return fmt.Errorf("%w: item requires a %s, got %T", ErrWrongFulfillmentKind, r.Kind(), v)
}

func (r DocumentRequirement) Accepts(v Value) error {
	dv, ok := v.(DocumentValue)
	if !ok {
		return mismatch(r, v)
	}
	if dv.DocumentID == uuid.Nil {
		return fmt.Errorf("%w: document id must not be empty", ErrInvalidValue)
	}
	return nil
}

func (r TextRequirement) Accepts(v Value) error {
	tv, ok := v.(TextValue)
	if !ok {
		return mismatch(r, v)
	}
	if strings.TrimSpace(tv.Text) == "" {
		return fmt.Errorf("%w: text must not be empty", ErrInvalidValue)
	}
	return nil
}

func (r NumberRequirement) Accepts(v Value) error {
	if _, ok := v.(NumberValue); !ok {
		return mismatch(r, v)
	}
	return nil
}

func (r YearRequirement) Accepts(v Value) error {
	yv, ok := v.(YearValue)
	if !ok {
		return mismatch(r, v)
	}
	if yv.Year < minYear || yv.Year > maxYear {
		return fmt.Errorf("%w: year %d out of range", ErrInvalidValue, yv.Year)
	}
	return nil
}

func (r BoolRequirement) Accepts(v Value) error {
	if _, ok := v.(BoolValue); !ok {
		return mismatch(r, v)
	}
	return nil
}
