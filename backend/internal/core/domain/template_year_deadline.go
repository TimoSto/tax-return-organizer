package domain

import (
	"fmt"
	"time"
)

// TemplateYearDeadline is a collector's fixed, absolute due date for a
// ChecklistTemplate in a given tax year — the only source of a checklist
// item's due date (ChecklistTemplate itself carries no deadline rule, since
// due dates can shift irregularly year to year regardless of recurrence,
// e.g. a statutory filing deadline or a monthly cutoff moved by a holiday).
// Month is nil for a 'once' template's single yearly deadline, or 1-12 for
// one of a 'monthly' template's per-month deadlines — the collector enters
// one of these per (template, year[, month]) when opening that year.
type TemplateYearDeadline struct {
	ID         int
	TemplateID int
	Year       int
	Month      *int
	DueDate    time.Time
}

func NewTemplateYearDeadline(templateID, year int, month *int, dueDate time.Time) (*TemplateYearDeadline, error) {
	if year < 1900 || year > 2200 {
		return nil, fmt.Errorf("%w: %d", ErrTaxYearOutOfRange, year)
	}
	if month != nil && (*month < 1 || *month > 12) {
		return nil, ErrInvalidMonth
	}
	if dueDate.IsZero() {
		return nil, ErrMissingDeadlineDate
	}
	return &TemplateYearDeadline{TemplateID: templateID, Year: year, Month: month, DueDate: dueDate}, nil
}
