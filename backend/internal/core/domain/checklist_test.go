package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewChecklistTemplate(t *testing.T) {
	collectorID := uuid.New()

	t.Run("valid monthly", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Bank statement", domain.RecurrenceMonthly)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tmpl.CategoryID != 1 {
			t.Errorf("CategoryID = %d, want 1", tmpl.CategoryID)
		}
	})

	t.Run("valid once", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Insurance premium notice", domain.RecurrenceOnce)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tmpl.CategoryID != 1 {
			t.Errorf("CategoryID = %d, want 1", tmpl.CategoryID)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		if _, err := domain.NewChecklistTemplate(collectorID, 1, "  ", domain.RecurrenceOnce); !errors.Is(err, domain.ErrEmptyChecklistTitle) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyChecklistTitle)
		}
	})

	t.Run("invalid recurrence", func(t *testing.T) {
		if _, err := domain.NewChecklistTemplate(collectorID, 1, "Bank statement", domain.Recurrence("yearly")); !errors.Is(err, domain.ErrInvalidRecurrence) {
			t.Errorf("err = %v, want %v", err, domain.ErrInvalidRecurrence)
		}
	})
}

func TestChecklistTemplate_GenerateItem(t *testing.T) {
	collectorID := uuid.New()
	wantDue := time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC)

	t.Run("once happy path", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Insurance premium notice", domain.RecurrenceOnce)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tmpl.ID = 42

		item, err := tmpl.GenerateItem(7, nil, wantDue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.TemplateID != 42 || item.ClientTaxYearID != 7 {
			t.Errorf("item = %+v, want TemplateID=42 ClientTaxYearID=7", item)
		}
		if item.Month != nil {
			t.Errorf("Month = %v, want nil for a 'once' item", item.Month)
		}
		if !item.DueDate.Equal(wantDue) {
			t.Errorf("DueDate = %v, want %v", item.DueDate, wantDue)
		}
	})

	t.Run("monthly happy path", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Payroll statement", domain.RecurrenceMonthly)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tmpl.ID = 42
		month := 1

		item, err := tmpl.GenerateItem(7, &month, wantDue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.Month == nil || *item.Month != 1 {
			t.Errorf("Month = %v, want 1", item.Month)
		}
		if !item.DueDate.Equal(wantDue) {
			t.Errorf("DueDate = %v, want %v", item.DueDate, wantDue)
		}
	})

	t.Run("once with a month errors", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Insurance premium notice", domain.RecurrenceOnce)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		month := 1
		if _, err := tmpl.GenerateItem(7, &month, wantDue); !errors.Is(err, domain.ErrRecurrenceMismatch) {
			t.Errorf("err = %v, want %v", err, domain.ErrRecurrenceMismatch)
		}
	})

	t.Run("monthly without a month errors", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Payroll statement", domain.RecurrenceMonthly)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := tmpl.GenerateItem(7, nil, wantDue); !errors.Is(err, domain.ErrRecurrenceMismatch) {
			t.Errorf("err = %v, want %v", err, domain.ErrRecurrenceMismatch)
		}
	})

	t.Run("monthly with an out-of-range month errors", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Payroll statement", domain.RecurrenceMonthly)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		month := 13
		if _, err := tmpl.GenerateItem(7, &month, wantDue); !errors.Is(err, domain.ErrInvalidMonth) {
			t.Errorf("err = %v, want %v", err, domain.ErrInvalidMonth)
		}
	})

	t.Run("zero due date errors", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Insurance premium notice", domain.RecurrenceOnce)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := tmpl.GenerateItem(7, nil, time.Time{}); !errors.Is(err, domain.ErrMissingDeadlineDate) {
			t.Errorf("err = %v, want %v", err, domain.ErrMissingDeadlineDate)
		}
	})
}

func TestChecklistItem_Satisfy(t *testing.T) {
	t.Run("with document", func(t *testing.T) {
		item := &domain.ChecklistItem{}
		docID := uuid.New()
		if err := item.SatisfyWithDocument(docID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !item.Done || item.DocumentID == nil || *item.DocumentID != docID {
			t.Errorf("item = %+v, want Done=true DocumentID=%v", item, docID)
		}

		if err := item.SatisfyWithValue("14"); !errors.Is(err, domain.ErrChecklistItemAlreadySatisfiedWithDocument) {
			t.Errorf("err = %v, want %v", err, domain.ErrChecklistItemAlreadySatisfiedWithDocument)
		}
	})

	t.Run("with structured value", func(t *testing.T) {
		item := &domain.ChecklistItem{}
		if err := item.SatisfyWithValue("14"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !item.Done || item.StructuredValue == nil || *item.StructuredValue != "14" {
			t.Errorf("item = %+v, want Done=true StructuredValue=14", item)
		}

		if err := item.SatisfyWithDocument(uuid.New()); !errors.Is(err, domain.ErrChecklistItemAlreadySatisfiedWithValue) {
			t.Errorf("err = %v, want %v", err, domain.ErrChecklistItemAlreadySatisfiedWithValue)
		}
	})

	t.Run("reset", func(t *testing.T) {
		item := &domain.ChecklistItem{}
		_ = item.SatisfyWithValue("14")
		item.Reset()
		if item.Done || item.DocumentID != nil || item.StructuredValue != nil {
			t.Errorf("item = %+v, want cleared state after Reset", item)
		}
	})
}

func TestChecklistItem_IsOverdue(t *testing.T) {
	due := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)

	t.Run("overdue when unsatisfied and past due", func(t *testing.T) {
		item := &domain.ChecklistItem{DueDate: due}
		if !item.IsOverdue(due.AddDate(0, 0, 1)) {
			t.Error("expected item to be overdue")
		}
	})

	t.Run("not overdue when done", func(t *testing.T) {
		item := &domain.ChecklistItem{DueDate: due, Done: true}
		if item.IsOverdue(due.AddDate(0, 0, 1)) {
			t.Error("expected a done item to never be overdue")
		}
	})

	t.Run("not overdue before due date", func(t *testing.T) {
		item := &domain.ChecklistItem{DueDate: due}
		if item.IsOverdue(due.AddDate(0, 0, -1)) {
			t.Error("expected item not to be overdue before its due date")
		}
	})
}
