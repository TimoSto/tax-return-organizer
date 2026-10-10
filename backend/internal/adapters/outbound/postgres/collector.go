package postgres

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

var _ ports.CollectorRepository = (*CollectorRepository)(nil)

type CollectorRepository struct {
	pool *pgxpool.Pool
}

func NewCollectorRepository(pool *pgxpool.Pool) *CollectorRepository {
	return &CollectorRepository{pool: pool}
}

func (r *CollectorRepository) GetCollector(ctx context.Context, id uuid.UUID) (*domain.Collector, error) {
	var c domain.Collector

	err := r.pool.QueryRow(ctx, `SELECT id, name, email FROM collectors WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCollectorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}

	return &c, nil
}

func (r *CollectorRepository) SaveCollector(ctx context.Context, collector *domain.Collector) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO collectors (id, name, email) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email`,
		collector.ID, collector.Name, collector.Email)
	if err != nil {
		return fmt.Errorf("save collector: %w", err)
	}

	return nil
}
