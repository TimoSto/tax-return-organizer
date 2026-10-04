package domain

import "github.com/google/uuid"

// ValueKind names a concrete Value / Requirement type. Adapters use it as the
// discriminator when persisting or serialising them.
type ValueKind string

const (
	KindDocument ValueKind = "document"
	KindText     ValueKind = "text"
	KindNumber   ValueKind = "number"
	KindYear     ValueKind = "year"
	KindBool     ValueKind = "bool"
)

// Value is what satisfies a checklist item: a reference to an uploaded
// document or a typed structured value. It is sealed — only this package
// implements it — so a type switch over the concrete types below is exhaustive.
type Value interface {
	Kind() ValueKind
	isValue()
}

type DocumentValue struct{ DocumentID uuid.UUID }

type TextValue struct{ Text string }

// NumberValue is a whole number, e.g. a count of homeoffice days.
type NumberValue struct{ Number int64 }

type YearValue struct{ Year int }

type BoolValue struct{ Bool bool }

func (DocumentValue) Kind() ValueKind { return KindDocument }
func (TextValue) Kind() ValueKind     { return KindText }
func (NumberValue) Kind() ValueKind   { return KindNumber }
func (YearValue) Kind() ValueKind     { return KindYear }
func (BoolValue) Kind() ValueKind     { return KindBool }

func (DocumentValue) isValue() {}
func (TextValue) isValue()     {}
func (NumberValue) isValue()   {}
func (YearValue) isValue()     {}
func (BoolValue) isValue()     {}
