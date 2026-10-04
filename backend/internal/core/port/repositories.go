package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

type CollectorRepository interface {
	Save(ctx context.Context, collector *domain.Collector) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Collector, error)
	List(ctx context.Context) ([]*domain.Collector, error)
}

type ClientRepository interface {
	Save(ctx context.Context, client *domain.Client) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
	ListByCollector(ctx context.Context, collectorID uuid.UUID) ([]*domain.Client, error)
}

type ClientTaxYearRepository interface {
	Save(ctx context.Context, taxYear *domain.ClientTaxYear) error
	FindByID(ctx context.Context, id int) (*domain.ClientTaxYear, error)
	ListByClient(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientTaxYear, error)
}

type CategoryRepository interface {
	Save(ctx context.Context, category *domain.Category) error
	FindByID(ctx context.Context, id int) (*domain.Category, error)
	ListByCollector(ctx context.Context, collectorID uuid.UUID) ([]*domain.Category, error)
}

type ChecklistTemplateRepository interface {
	Save(ctx context.Context, template *domain.ChecklistTemplate) error
	FindByID(ctx context.Context, id int) (*domain.ChecklistTemplate, error)
	ListByCollector(ctx context.Context, collectorID uuid.UUID) ([]*domain.ChecklistTemplate, error)
	Delete(ctx context.Context, id int) error
}

type ChecklistItemRepository interface {
	Save(ctx context.Context, item *domain.ChecklistItem) error
	ListByTemplate(ctx context.Context, templateID int) ([]*domain.ChecklistItem, error)
	ListByClientTaxYear(ctx context.Context, clientTaxYearID int) ([]*domain.ChecklistItem, error)
}

type DocumentRepository interface {
	Save(ctx context.Context, document *domain.Document) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Document, error)
	ListByClientTaxYear(ctx context.Context, clientTaxYearID int) ([]*domain.Document, error)
	ListByCategory(ctx context.Context, categoryID int) ([]*domain.Document, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
