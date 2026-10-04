package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Recurrence string

const (
	RecurrenceOnce    Recurrence = "once"
	RecurrenceMonthly Recurrence = "monthly"
)

func (r Recurrence) valid() bool {
	return r == RecurrenceOnce || r == RecurrenceMonthly
}

// ChecklistTemplate is defined by a Collector and reused across that
// collector's clients: it describes one required piece of information
// (e.g. "bank statement") within a Category, what satisfies it (a document or
// a typed value, see Requirement), and how often it's required.
type ChecklistTemplate struct {
	ID          int
	CollectorID uuid.UUID
	CategoryID  int
	Title       string
	Recurrence  Recurrence
	Requirement Requirement
}

func NewChecklistTemplate(collectorID uuid.UUID, categoryID int, title string, recurrence Recurrence, requirement Requirement) (*ChecklistTemplate, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyChecklistTitle
	}
	if !recurrence.valid() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidRecurrence, recurrence)
	}
	if requirement == nil {
		return nil, ErrMissingRequirement
	}
	return &ChecklistTemplate{
		CollectorID: collectorID,
		CategoryID:  categoryID,
		Title:       title,
		Recurrence:  recurrence,
		Requirement: requirement,
	}, nil
}

// GenerateItem builds one client's checklist item for this template and tax
// year (and, for a 'monthly' template, a given month). month must be nil
// for a 'once' template and 1-12 for a 'monthly' one; a mismatch returns
// ErrRecurrenceMismatch or ErrInvalidMonth.
func (t *ChecklistTemplate) GenerateItem(clientTaxYearID int, month *int) (*ChecklistItem, error) {
	if t.Recurrence == RecurrenceOnce && month != nil {
		return nil, ErrRecurrenceMismatch
	}
	if t.Recurrence == RecurrenceMonthly {
		if month == nil {
			return nil, ErrRecurrenceMismatch
		}
		if *month < 1 || *month > 12 {
			return nil, ErrInvalidMonth
		}
	}
	return &ChecklistItem{
		TemplateID:      t.ID,
		ClientTaxYearID: clientTaxYearID,
		Month:           month,
		Requirement:     t.Requirement,
	}, nil
}

// ChecklistItem is one client's instance of a template's requirement for a
// given tax year (and, for a 'monthly' template, a given month). Requirement
// is copied from the template when the item is generated, so the item alone
// says what it needs and later template edits don't invalidate it. Value is
// nil while the item is outstanding.
type ChecklistItem struct {
	ID              int
	TemplateID      int
	ClientTaxYearID int
	Month           *int // nil for items generated from a 'once' template

	Requirement Requirement
	Value       Value
}

// Done reports whether the item has been satisfied.
func (i *ChecklistItem) Done() bool {
	return i.Value != nil
}

// Satisfy records v as what fulfills the item. It errors with
// ErrMissingRequirement for an item without a requirement, and otherwise with
// whatever the requirement's Accepts returns (ErrWrongFulfillmentKind,
// ErrInvalidValue); a rejected value leaves the item untouched.
func (i *ChecklistItem) Satisfy(v Value) error {
	if i.Requirement == nil {
		return ErrMissingRequirement
	}
	if err := i.Requirement.Accepts(v); err != nil {
		return err
	}
	i.Value = v
	return nil
}

// Reset clears whatever satisfied the item, marking it not done again.
func (i *ChecklistItem) Reset() {
	i.Value = nil
}
