package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewDocument(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		d, err := domain.NewDocument(1, "statement.pdf", "application/pdf", []byte("content"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.ClientTaxYearID != 1 {
			t.Errorf("ClientTaxYearID = %d, want 1", d.ClientTaxYearID)
		}
		if d.SizeBytes != int64(len("content")) {
			t.Errorf("SizeBytes = %d, want %d", d.SizeBytes, len("content"))
		}
		if d.CategoryID != nil {
			t.Error("expected CategoryID to start unclassified (nil)")
		}
	})

	t.Run("empty filename", func(t *testing.T) {
		if _, err := domain.NewDocument(1, "  ", "application/pdf", []byte("content")); !errors.Is(err, domain.ErrEmptyDocumentFilename) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyDocumentFilename)
		}
	})

	t.Run("empty mime type", func(t *testing.T) {
		if _, err := domain.NewDocument(1, "statement.pdf", "", []byte("content")); !errors.Is(err, domain.ErrEmptyDocumentMimeType) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyDocumentMimeType)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		if _, err := domain.NewDocument(1, "statement.pdf", "application/pdf", nil); !errors.Is(err, domain.ErrEmptyDocumentContent) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyDocumentContent)
		}
	})
}

func TestDocument_AssignCategoryAndUnclassify(t *testing.T) {
	d, err := domain.NewDocument(1, "statement.pdf", "application/pdf", []byte("content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	d.AssignCategory(5)
	if d.CategoryID == nil || *d.CategoryID != 5 {
		t.Fatalf("CategoryID = %v, want 5", d.CategoryID)
	}

	d.Unclassify()
	if d.CategoryID != nil {
		t.Fatalf("CategoryID = %v, want nil after Unclassify", d.CategoryID)
	}
}

func TestDocument_SetProperty(t *testing.T) {
	d, err := domain.NewDocument(1, "statement.pdf", "application/pdf", []byte("content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	d.SetProperty("employer", "Acme Corp")
	if got := d.Properties["employer"]; got != "Acme Corp" {
		t.Errorf("Properties[employer] = %q, want %q", got, "Acme Corp")
	}
}
