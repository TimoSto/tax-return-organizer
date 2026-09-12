package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID               uuid.UUID
	TaxYearID        int
	CategoryID       *int // nil = unclassified
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	Content          []byte
	Properties       map[string]string
	UploadedAt       time.Time
	UpdatedAt        time.Time
}

func NewDocument(taxYearID int, originalFilename, mimeType string, content []byte) (*Document, error) {
	originalFilename = strings.TrimSpace(originalFilename)
	if originalFilename == "" {
		return nil, fmt.Errorf("document filename must not be empty")
	}
	mimeType = strings.TrimSpace(mimeType)
	if mimeType == "" {
		return nil, fmt.Errorf("document mime type must not be empty")
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("document content must not be empty")
	}

	now := time.Now()
	return &Document{
		ID:               uuid.New(),
		TaxYearID:        taxYearID,
		OriginalFilename: originalFilename,
		MimeType:         mimeType,
		SizeBytes:        int64(len(content)),
		Content:          content,
		Properties:       map[string]string{},
		UploadedAt:       now,
		UpdatedAt:        now,
	}, nil
}

func (d *Document) AssignCategory(categoryID int) {
	d.CategoryID = &categoryID
	d.UpdatedAt = time.Now()
}

func (d *Document) Unclassify() {
	d.CategoryID = nil
	d.UpdatedAt = time.Now()
}

func (d *Document) SetProperty(key, value string) {
	d.Properties[key] = value
	d.UpdatedAt = time.Now()
}
