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

type ChecklistTemplate struct {
	ID         int
	TaxYearID  int
	Title      string
	Recurrence Recurrence
}

func NewChecklistTemplate(taxYearID int, title string, recurrence Recurrence) (*ChecklistTemplate, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("checklist template title must not be empty")
	}
	if !recurrence.valid() {
		return nil, fmt.Errorf("invalid recurrence %q", recurrence)
	}
	return &ChecklistTemplate{TaxYearID: taxYearID, Title: title, Recurrence: recurrence}, nil
}

// GenerateItems expands a template into its checklist items: a single
// item for a 'once' template, or one item per month for a 'monthly' one.
func (t *ChecklistTemplate) GenerateItems() []*ChecklistItem {
	if t.Recurrence == RecurrenceMonthly {
		items := make([]*ChecklistItem, 0, 12)
		for month := 1; month <= 12; month++ {
			m := month
			items = append(items, &ChecklistItem{TemplateID: t.ID, Month: &m})
		}
		return items
	}
	return []*ChecklistItem{{TemplateID: t.ID}}
}

type ChecklistItem struct {
	ID         int
	TemplateID int
	Month      *int // nil for items generated from a 'once' template
	DocumentID *uuid.UUID
	Done       bool
}

func (i *ChecklistItem) MarkDone(documentID uuid.UUID) {
	i.DocumentID = &documentID
	i.Done = true
}

func (i *ChecklistItem) MarkUndone() {
	i.DocumentID = nil
	i.Done = false
}
