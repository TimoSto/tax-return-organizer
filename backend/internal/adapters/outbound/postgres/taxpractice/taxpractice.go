// Package taxpractice contains the PostgreSQL implementation of the tax practice repository.
package taxpractice

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// uniqueViolation is the SQLSTATE of a violated unique or primary key constraint.
const uniqueViolation = "23505"

var _ outbound.TaxPracticeRepository = (*Repository)(nil)

// Repository is the Postgres implementation of outbound.TaxPracticeRepository.
type Repository struct {
	db pgdb.DBTX
}

// NewRepository returns a repository backed by db, the pool or a transaction.
func NewRepository(db pgdb.DBTX) *Repository {
	return &Repository{db: db}
}

// GetTaxPractice implements outbound.TaxPracticeRepository.
func (r *Repository) GetTaxPractice(ctx context.Context, id domain.TaxPracticeID) (*domain.TaxPractice, error) {
	var row taxPracticeRow

	err := r.db.QueryRow(ctx, `SELECT id, name, email FROM tax_practices WHERE id = $1`, uuid.UUID(id)).
		Scan(&row.ID, &row.Name, &row.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTaxPracticeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query tax practice: %w", err)
	}

	return row.toDomain()
}

// SaveTaxPractice implements outbound.TaxPracticeRepository.
func (r *Repository) SaveTaxPractice(ctx context.Context, taxPractice *domain.TaxPractice) error {
	row := newTaxPracticeRow(taxPractice)

	_, err := r.db.Exec(ctx, `
		INSERT INTO tax_practices (id, name, email) VALUES ($1, $2, $3)`,
		row.ID, row.Name, row.Email)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return domain.ErrTaxPracticeIDTaken
	}

	if err != nil {
		return fmt.Errorf("insert tax practice: %w", err)
	}

	return nil
}

// taxPracticeRow mirrors the tax practices table and keeps SQL concerns out of the domain type.
type taxPracticeRow struct {
	ID    uuid.UUID
	Name  string
	Email string
}

func newTaxPracticeRow(c *domain.TaxPractice) taxPracticeRow {
	return taxPracticeRow{ID: uuid.UUID(c.ID()), Name: c.Name(), Email: string(c.Email())}
}

func (r taxPracticeRow) toDomain() (*domain.TaxPractice, error) {
	c, err := domain.RehydrateTaxPractice(domain.TaxPracticeID(r.ID), r.Name, domain.Email(r.Email))
	if err != nil {
		return nil, outbound.CorruptData("tax practice "+r.ID.String(), err)
	}

	return c, nil
}
