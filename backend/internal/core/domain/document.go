package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Document is a file a client uploaded (blob + metadata), belonging to one
// of that client's tax years, optionally classified under a
// collector-defined category.
type Document struct {
	ID               uuid.UUID
	ClientTaxYearID  int
	CategoryID       *int // nil = unclassified
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	Content          []byte
	Properties       map[string]string
	UploadedAt       time.Time
	UpdatedAt        time.Time
}

func NewDocument(clientTaxYearID int, originalFilename, mimeType string, content []byte) (*Document, error) {
	originalFilename = strings.TrimSpace(originalFilename)
	if originalFilename == "" {
		return nil, ErrEmptyDocumentFilename
	}
	mimeType = strings.TrimSpace(mimeType)
	if mimeType == "" {
		return nil, ErrEmptyDocumentMimeType
	}
	if len(content) == 0 {
		return nil, ErrEmptyDocumentContent
	}

	now := time.Now()
	return &Document{
		ID:               uuid.New(),
		ClientTaxYearID:  clientTaxYearID,
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
