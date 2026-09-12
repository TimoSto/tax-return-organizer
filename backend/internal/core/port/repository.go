package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type TaxYearRepository interface {
	Save(ctx context.Context, taxYear *domain.TaxYear) error
	FindByYear(ctx context.Context, year int) (*domain.TaxYear, error)
	List(ctx context.Context) ([]*domain.TaxYear, error)
}

type CategoryRepository interface {
	Save(ctx context.Context, category *domain.Category) error
	FindByID(ctx context.Context, id int) (*domain.Category, error)
	List(ctx context.Context) ([]*domain.Category, error)
}

type DocumentRepository interface {
	Save(ctx context.Context, document *domain.Document) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Document, error)
	ListByTaxYear(ctx context.Context, taxYearID int) ([]*domain.Document, error)
	ListByCategory(ctx context.Context, categoryID int) ([]*domain.Document, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ChecklistTemplateRepository interface {
	Save(ctx context.Context, template *domain.ChecklistTemplate) error
	FindByID(ctx context.Context, id int) (*domain.ChecklistTemplate, error)
	ListByTaxYear(ctx context.Context, taxYearID int) ([]*domain.ChecklistTemplate, error)
	Delete(ctx context.Context, id int) error
}

type ChecklistItemRepository interface {
	Save(ctx context.Context, item *domain.ChecklistItem) error
	ListByTemplate(ctx context.Context, templateID int) ([]*domain.ChecklistItem, error)
}
