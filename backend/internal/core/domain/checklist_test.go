package domain

import "testing"

func TestNewChecklistTemplate_RejectsInvalidRecurrence(t *testing.T) {
	if _, err := NewChecklistTemplate(1, "Payroll", "weekly"); err == nil {
		t.Fatal("expected error for invalid recurrence, got nil")
	}
}

func TestNewChecklistTemplate_RejectsEmptyTitle(t *testing.T) {
	if _, err := NewChecklistTemplate(1, "  ", RecurrenceOnce); err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
}

func TestGenerateItems_Once(t *testing.T) {
	tpl, err := NewChecklistTemplate(1, "Annual bank statement", RecurrenceOnce)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tpl.ID = 42

	items := tpl.GenerateItems()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Month != nil {
		t.Fatalf("expected nil month for 'once' item, got %v", *items[0].Month)
	}
	if items[0].TemplateID != 42 {
		t.Fatalf("expected item to reference template 42, got %d", items[0].TemplateID)
	}
}

func TestGenerateItems_Monthly(t *testing.T) {
	tpl, err := NewChecklistTemplate(1, "Payroll statement", RecurrenceMonthly)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tpl.ID = 7

	items := tpl.GenerateItems()
	if len(items) != 12 {
		t.Fatalf("expected 12 items, got %d", len(items))
	}
	for i, item := range items {
		wantMonth := i + 1
		if item.Month == nil || *item.Month != wantMonth {
			t.Fatalf("item %d: expected month %d, got %v", i, wantMonth, item.Month)
		}
		if item.TemplateID != 7 {
			t.Fatalf("item %d: expected template 7, got %d", i, item.TemplateID)
		}
	}
}
