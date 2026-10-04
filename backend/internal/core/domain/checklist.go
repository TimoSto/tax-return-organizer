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
// (e.g. "bank statement") within a Category, how often it's required, and
// when it's due.
type ChecklistTemplate struct {
	ID          int
	CollectorID uuid.UUID
	CategoryID  int
	Title       string
	Recurrence  Recurrence
	Deadline    Deadline
}

func NewChecklistTemplate(collectorID uuid.UUID, categoryID int, title string, recurrence Recurrence, deadline Deadline) (*ChecklistTemplate, error) {
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
		Deadline:    deadline,
	}, nil
}

// GenerateItems expands a template into one client's checklist items for a
// given tax year: a single item for a 'once' template, or one item per
// month for a 'monthly' one. Each item's due date is derived from the
// template's Deadline rule.
func (t *ChecklistTemplate) GenerateItems(clientTaxYearID, year int) []*ChecklistItem {
	if t.Recurrence == RecurrenceMonthly {
		items := make([]*ChecklistItem, 0, 12)
		for month := 1; month <= 12; month++ {
			m := month
			items = append(items, &ChecklistItem{
				TemplateID:      t.ID,
				ClientTaxYearID: clientTaxYearID,
				Month:           &m,
				DueDate:         t.Deadline.DueDate(year, month),
			})
		}
		return items
	}
	return []*ChecklistItem{{
		TemplateID:      t.ID,
		ClientTaxYearID: clientTaxYearID,
		DueDate:         t.Deadline.DueDate(year, 0),
	}}
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
