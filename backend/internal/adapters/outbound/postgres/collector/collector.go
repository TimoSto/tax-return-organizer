// Package collector contains the PostgreSQL implementation of the collector repository.
package collector

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

// uniqueViolation is the SQLSTATE of a violated unique or primary key constraint.
const uniqueViolation = "23505"

var _ ports.CollectorRepository = (*Repository)(nil)

// Repository is the Postgres implementation of ports.CollectorRepository.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a repository backed by pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetCollector implements ports.CollectorRepository.
func (r *Repository) GetCollector(ctx context.Context, id domain.CollectorID) (*domain.Collector, error) {
	var row collectorRow

	err := r.pool.QueryRow(ctx, `SELECT id, name, email FROM collectors WHERE id = $1`, string(id)).
		Scan(&row.ID, &row.Name, &row.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCollectorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}

	return row.toDomain(), nil
}

// SaveCollector implements ports.CollectorRepository.
func (r *Repository) SaveCollector(ctx context.Context, collector *domain.Collector) error {
	row := newCollectorRow(collector)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO collectors (id, name, email) VALUES ($1, $2, $3)`,
		row.ID, row.Name, row.Email)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return domain.ErrCollectorIDTaken
	}

	if err != nil {
		return fmt.Errorf("save collector: %w", err)
	}

	return nil
}

// collectorRow mirrors the collectors table and keeps SQL concerns out of the domain type.
type collectorRow struct {
	ID    string
	Name  string
	Email string
}

func newCollectorRow(c *domain.Collector) collectorRow {
	return collectorRow{ID: string(c.ID), Name: c.Name, Email: c.Email}
}

func (r collectorRow) toDomain() *domain.Collector {
	return &domain.Collector{ID: domain.CollectorID(r.ID), Name: r.Name, Email: r.Email}
}
