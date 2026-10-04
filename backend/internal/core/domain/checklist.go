package domain

import (
	"fmt"
	"strings"
	"time"

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
// (e.g. "bank statement") within a Category and how often it's required.
// It carries no deadline of its own — due dates are never derived from a
// rule on the template, since even a 'monthly' requirement's due date can
// shift year to year (e.g. around holidays). Instead the collector enters
// an absolute TemplateYearDeadline for each tax year (and, for 'monthly',
// each month of it) when opening that year.
type ChecklistTemplate struct {
	ID          int
	CollectorID uuid.UUID
	CategoryID  int
	Title       string
	Recurrence  Recurrence
}

func NewChecklistTemplate(collectorID uuid.UUID, categoryID int, title string, recurrence Recurrence) (*ChecklistTemplate, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyChecklistTitle
	}
	if !recurrence.valid() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidRecurrence, recurrence)
	}
	return &ChecklistTemplate{
		CollectorID: collectorID,
		CategoryID:  categoryID,
		Title:       title,
		Recurrence:  recurrence,
	}, nil
}

// GenerateItem builds one client's checklist item for this template and tax
// year, using a due date and (for 'monthly' templates) a month already
// resolved by the caller from this template's TemplateYearDeadline for that
// (year[, month]). month must be nil for a 'once' template and 1-12 for a
// 'monthly' one; a mismatch returns ErrRecurrenceMismatch or ErrInvalidMonth.
// A zero dueDate returns ErrMissingDeadlineDate.
func (t *ChecklistTemplate) GenerateItem(clientTaxYearID int, month *int, dueDate time.Time) (*ChecklistItem, error) {
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
	if dueDate.IsZero() {
		return nil, ErrMissingDeadlineDate
	}
	return &ChecklistItem{
		TemplateID:      t.ID,
		ClientTaxYearID: clientTaxYearID,
		Month:           month,
		DueDate:         dueDate,
	}, nil
}

// ChecklistItem is one client's instance of a template's requirement for a
// given tax year (and, for a 'monthly' template, a given month). It's
// satisfied either by a linked Document or by a structured value entered
// directly — never both.
type ChecklistItem struct {
	ID              int
	TemplateID      int
	ClientTaxYearID int
	Month           *int // nil for items generated from a 'once' template
	DueDate         time.Time

	DocumentID      *uuid.UUID
	StructuredValue *string

	Done bool
}

// SatisfyWithDocument links the item to a document that fulfills it.
// It errors if the item is already satisfied with a structured value.
func (i *ChecklistItem) SatisfyWithDocument(documentID uuid.UUID) error {
	if i.StructuredValue != nil {
		return ErrChecklistItemAlreadySatisfiedWithValue
	}
	i.DocumentID = &documentID
	i.Done = true
	return nil
}

// SatisfyWithValue records a structured value (e.g. a count or amount) that
// fulfills the item directly, in place of uploading a file. It errors if the
// item is already satisfied with a document.
func (i *ChecklistItem) SatisfyWithValue(value string) error {
	if i.DocumentID != nil {
		return ErrChecklistItemAlreadySatisfiedWithDocument
	}
	i.StructuredValue = &value
	i.Done = true
	return nil
}

// Reset clears whatever satisfied the item, marking it not done again.
func (i *ChecklistItem) Reset() {
	i.DocumentID = nil
	i.StructuredValue = nil
	i.Done = false
}

// IsOverdue reports whether the item is still unsatisfied past its due date.
func (i *ChecklistItem) IsOverdue(now time.Time) bool {
	return !i.Done && now.After(i.DueDate)
}
