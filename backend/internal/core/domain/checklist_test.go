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

	t.Run("valid", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Bank statement", domain.RecurrenceMonthly, domain.Deadline{DaysAfterPeriodEnd: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tmpl.CategoryID != 1 {
			t.Errorf("CategoryID = %d, want 1", tmpl.CategoryID)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		if _, err := domain.NewChecklistTemplate(collectorID, 1, "  ", domain.RecurrenceOnce, domain.Deadline{}); !errors.Is(err, domain.ErrEmptyChecklistTitle) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyChecklistTitle)
		}
	})

	t.Run("invalid recurrence", func(t *testing.T) {
		if _, err := domain.NewChecklistTemplate(collectorID, 1, "Bank statement", domain.Recurrence("yearly"), domain.Deadline{}); !errors.Is(err, domain.ErrInvalidRecurrence) {
			t.Errorf("err = %v, want %v", err, domain.ErrInvalidRecurrence)
		}
	})
}

func TestChecklistTemplate_GenerateItems(t *testing.T) {
	collectorID := uuid.New()

	t.Run("once", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Insurance premium notice", domain.RecurrenceOnce, domain.Deadline{DaysAfterPeriodEnd: 90})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tmpl.ID = 42

		items := tmpl.GenerateItems(7, 2024)
		if len(items) != 1 {
			t.Fatalf("len(items) = %d, want 1", len(items))
		}
		item := items[0]
		if item.TemplateID != 42 || item.ClientTaxYearID != 7 {
			t.Errorf("item = %+v, want TemplateID=42 ClientTaxYearID=7", item)
		}
		if item.Month != nil {
			t.Errorf("Month = %v, want nil for a 'once' item", item.Month)
		}
		wantDue := time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC)
		if !item.DueDate.Equal(wantDue) {
			t.Errorf("DueDate = %v, want %v", item.DueDate, wantDue)
		}
	})

	t.Run("monthly", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Payroll statement", domain.RecurrenceMonthly, domain.Deadline{DaysAfterPeriodEnd: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tmpl.ID = 42

		items := tmpl.GenerateItems(7, 2024)
		if len(items) != 12 {
			t.Fatalf("len(items) = %d, want 12", len(items))
		}

		jan := items[0]
		if jan.Month == nil || *jan.Month != 1 {
			t.Fatalf("items[0].Month = %v, want 1", jan.Month)
		}
		wantDue := time.Date(2024, time.February, 5, 0, 0, 0, 0, time.UTC)
		if !jan.DueDate.Equal(wantDue) {
			t.Errorf("items[0].DueDate = %v, want %v", jan.DueDate, wantDue)
		}
	})
}

func TestDeadline_DueDate(t *testing.T) {
	t.Run("once (month 0) uses end of tax year", func(t *testing.T) {
		d := domain.Deadline{DaysAfterPeriodEnd: 0}
		got := d.DueDate(2024, 0)
		want := time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("DueDate = %v, want %v", got, want)
		}
	})

	t.Run("monthly uses end of covered month", func(t *testing.T) {
		d := domain.Deadline{DaysAfterPeriodEnd: 5}
		got := d.DueDate(2024, 2) // February 2024 is a leap year, ends the 29th
		want := time.Date(2024, time.March, 5, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("DueDate = %v, want %v", got, want)
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
