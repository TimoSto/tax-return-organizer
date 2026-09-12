package domain_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewDocument_RejectsEmptyFilename(t *testing.T) {
	if _, err := domain.NewDocument(1, "  ", "application/pdf", []byte("x")); err == nil {
		t.Fatal("expected error for empty filename, got nil")
	}
}

func TestNewDocument_RejectsEmptyContent(t *testing.T) {
	if _, err := domain.NewDocument(1, "payslip.pdf", "application/pdf", nil); err == nil {
		t.Fatal("expected error for empty content, got nil")
	}
}

func TestNewDocument_SetsSizeFromContent(t *testing.T) {
	doc, err := domain.NewDocument(1, "payslip.pdf", "application/pdf", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.SizeBytes != 5 {
		t.Fatalf("expected size 5, got %d", doc.SizeBytes)
	}
	if doc.CategoryID != nil {
		t.Fatalf("expected new document to be unclassified, got category %d", *doc.CategoryID)
	}
}

func TestDocument_AssignCategoryThenUnclassify(t *testing.T) {
	doc, err := domain.NewDocument(1, "payslip.pdf", "application/pdf", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc.AssignCategory(3)
	if doc.CategoryID == nil || *doc.CategoryID != 3 {
		t.Fatalf("expected category 3, got %v", doc.CategoryID)
	}

	doc.Unclassify()
	if doc.CategoryID != nil {
		t.Fatalf("expected nil category after Unclassify, got %d", *doc.CategoryID)
	}
}
