package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewChecklistTemplate(t *testing.T) {
	collectorID := uuid.New()
	docReq := domain.DocumentRequirement{}

	t.Run("valid monthly document", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Bank statement", domain.RecurrenceMonthly, docReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tmpl.CategoryID != 1 {
			t.Errorf("CategoryID = %d, want 1", tmpl.CategoryID)
		}
		if tmpl.Requirement != docReq {
			t.Errorf("Requirement = %v, want %v", tmpl.Requirement, docReq)
		}
	})

	t.Run("valid once number", func(t *testing.T) {
		req := domain.NewNumberRequirement(" days ")
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Homeoffice days", domain.RecurrenceOnce, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, ok := tmpl.Requirement.(domain.NumberRequirement)
		if !ok || got.Unit != "days" {
			t.Errorf("Requirement = %#v, want NumberRequirement{Unit: \"days\"}", tmpl.Requirement)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		if _, err := domain.NewChecklistTemplate(collectorID, 1, "  ", domain.RecurrenceOnce, docReq); !errors.Is(err, domain.ErrEmptyChecklistTitle) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyChecklistTitle)
		}
	})

	t.Run("invalid recurrence", func(t *testing.T) {
		if _, err := domain.NewChecklistTemplate(collectorID, 1, "Bank statement", domain.Recurrence("yearly"), docReq); !errors.Is(err, domain.ErrInvalidRecurrence) {
			t.Errorf("err = %v, want %v", err, domain.ErrInvalidRecurrence)
		}
	})

	t.Run("missing requirement", func(t *testing.T) {
		if _, err := domain.NewChecklistTemplate(collectorID, 1, "Bank statement", domain.RecurrenceOnce, nil); !errors.Is(err, domain.ErrMissingRequirement) {
			t.Errorf("err = %v, want %v", err, domain.ErrMissingRequirement)
		}
	})
}

func TestChecklistTemplate_GenerateItem(t *testing.T) {
	collectorID := uuid.New()
	docReq := domain.DocumentRequirement{}

	t.Run("once happy path", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Insurance premium notice", domain.RecurrenceOnce, docReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tmpl.ID = 42

		item, err := tmpl.GenerateItem(7, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.TemplateID != 42 || item.ClientTaxYearID != 7 {
			t.Errorf("item = %+v, want TemplateID=42 ClientTaxYearID=7", item)
		}
		if item.Month != nil {
			t.Errorf("Month = %v, want nil for a 'once' item", item.Month)
		}
	})

	t.Run("monthly happy path", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Payroll statement", domain.RecurrenceMonthly, docReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tmpl.ID = 42
		month := 1

		item, err := tmpl.GenerateItem(7, &month)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.Month == nil || *item.Month != 1 {
			t.Errorf("Month = %v, want 1", item.Month)
		}
	})

	t.Run("copies the requirement from the template", func(t *testing.T) {
		req := domain.NewNumberRequirement("days")
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Homeoffice days", domain.RecurrenceOnce, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		item, err := tmpl.GenerateItem(7, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.Requirement != domain.Requirement(req) {
			t.Errorf("Requirement = %#v, want %#v", item.Requirement, req)
		}
		if item.Done() {
			t.Error("a freshly generated item must not be done")
		}
	})

	t.Run("once with a month errors", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Insurance premium notice", domain.RecurrenceOnce, docReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		month := 1
		if _, err := tmpl.GenerateItem(7, &month); !errors.Is(err, domain.ErrRecurrenceMismatch) {
			t.Errorf("err = %v, want %v", err, domain.ErrRecurrenceMismatch)
		}
	})

	t.Run("monthly without a month errors", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Payroll statement", domain.RecurrenceMonthly, docReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := tmpl.GenerateItem(7, nil); !errors.Is(err, domain.ErrRecurrenceMismatch) {
			t.Errorf("err = %v, want %v", err, domain.ErrRecurrenceMismatch)
		}
	})

	t.Run("monthly with an out-of-range month errors", func(t *testing.T) {
		tmpl, err := domain.NewChecklistTemplate(collectorID, 1, "Payroll statement", domain.RecurrenceMonthly, docReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		month := 13
		if _, err := tmpl.GenerateItem(7, &month); !errors.Is(err, domain.ErrInvalidMonth) {
			t.Errorf("err = %v, want %v", err, domain.ErrInvalidMonth)
		}
	})
}

func TestChecklistItem_Satisfy(t *testing.T) {
	t.Run("document item with document", func(t *testing.T) {
		item := &domain.ChecklistItem{Requirement: domain.DocumentRequirement{}}
		docID := uuid.New()
		if err := item.Satisfy(domain.DocumentValue{DocumentID: docID}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !item.Done() || item.Value != domain.Value(domain.DocumentValue{DocumentID: docID}) {
			t.Errorf("item = %+v, want done with DocumentValue %v", item, docID)
		}
	})

	t.Run("document item rejects a number", func(t *testing.T) {
		item := &domain.ChecklistItem{Requirement: domain.DocumentRequirement{}}
		if err := item.Satisfy(domain.NumberValue{Number: 14}); !errors.Is(err, domain.ErrWrongFulfillmentKind) {
			t.Errorf("err = %v, want %v", err, domain.ErrWrongFulfillmentKind)
		}
		if item.Done() {
			t.Error("item must not be done after a rejected value")
		}
	})

	t.Run("number item with number", func(t *testing.T) {
		item := &domain.ChecklistItem{Requirement: domain.NewNumberRequirement("days")}
		if err := item.Satisfy(domain.NumberValue{Number: 14}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got, ok := item.Value.(domain.NumberValue); !ok || got.Number != 14 {
			t.Errorf("Value = %#v, want NumberValue{14}", item.Value)
		}
	})

	t.Run("number item rejects a document", func(t *testing.T) {
		item := &domain.ChecklistItem{Requirement: domain.NewNumberRequirement("days")}
		if err := item.Satisfy(domain.DocumentValue{DocumentID: uuid.New()}); !errors.Is(err, domain.ErrWrongFulfillmentKind) {
			t.Errorf("err = %v, want %v", err, domain.ErrWrongFulfillmentKind)
		}
	})

	t.Run("invalid value leaves the item untouched", func(t *testing.T) {
		item := &domain.ChecklistItem{Requirement: domain.YearRequirement{}}
		if err := item.Satisfy(domain.YearValue{Year: 1800}); !errors.Is(err, domain.ErrInvalidValue) {
			t.Errorf("err = %v, want %v", err, domain.ErrInvalidValue)
		}
		if item.Done() || item.Value != nil {
			t.Errorf("item = %+v, want untouched after a rejected value", item)
		}
	})

	t.Run("pointer value is rejected without storing it", func(t *testing.T) {
		item := &domain.ChecklistItem{Requirement: domain.TextRequirement{}}
		if err := item.Satisfy(&domain.TextValue{Text: "x"}); !errors.Is(err, domain.ErrWrongFulfillmentKind) {
			t.Errorf("err = %v, want %v", err, domain.ErrWrongFulfillmentKind)
		}
		if item.Done() {
			t.Error("item must not be done after a rejected value")
		}
	})

	t.Run("item without a requirement", func(t *testing.T) {
		item := &domain.ChecklistItem{}
		if err := item.Satisfy(domain.BoolValue{Bool: true}); !errors.Is(err, domain.ErrMissingRequirement) {
			t.Errorf("err = %v, want %v", err, domain.ErrMissingRequirement)
		}
	})

	t.Run("reset", func(t *testing.T) {
		item := &domain.ChecklistItem{Requirement: domain.BoolRequirement{}}
		_ = item.Satisfy(domain.BoolValue{Bool: true})
		item.Reset()
		if item.Done() || item.Value != nil {
			t.Errorf("item = %+v, want cleared state after Reset", item)
		}
	})
}
