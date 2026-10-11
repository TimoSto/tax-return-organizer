package collector

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

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
func (r *Repository) GetCollector(ctx context.Context, id uuid.UUID) (*domain.Collector, error) {
	var row collectorRow

	err := r.pool.QueryRow(ctx, `SELECT id, name, email FROM collectors WHERE id = $1`, id).
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
		INSERT INTO collectors (id, name, email) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email`,
		row.ID, row.Name, row.Email)
	if err != nil {
		return fmt.Errorf("save collector: %w", err)
	}

	return nil
}

// collectorRow mirrors the collectors table and keeps SQL concerns out of the domain type.
type collectorRow struct {
	ID    uuid.UUID
	Name  string
	Email string
}

func newCollectorRow(c *domain.Collector) collectorRow {
	return collectorRow{ID: c.ID, Name: c.Name, Email: c.Email}
}

func (r collectorRow) toDomain() *domain.Collector {
	return &domain.Collector{ID: r.ID, Name: r.Name, Email: r.Email}
}
