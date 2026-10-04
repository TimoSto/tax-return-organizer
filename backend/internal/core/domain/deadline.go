package domain

import "time"

// Deadline describes when a checklist item generated from a template becomes
// due, relative to the period it covers rather than as an absolute date: a
// ChecklistTemplate is reused across tax years, so only a rule survives from
// one year to the next, not a fixed date.
type Deadline struct {
	// DaysAfterPeriodEnd is how many days after the covered period ends the
	// item is due. The period is the tax year itself for a 'once' template,
	// or the covered month for a 'monthly' template.
	DaysAfterPeriodEnd int
}

// DueDate computes the absolute due date for a given tax year and, for a
// 'monthly' template, the covered month (1-12). Pass month 0 for the 'once'
// case, where the period is the whole tax year (ending December 31).
func (d Deadline) DueDate(year, month int) time.Time {
	var periodEnd time.Time
	if month == 0 {
		periodEnd = time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	} else {
		// Day 0 of the following month is the last day of `month`.
		periodEnd = time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC)
	}
	return periodEnd.AddDate(0, 0, d.DaysAfterPeriodEnd)
}
