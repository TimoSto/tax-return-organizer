// Package collector contains the PostgreSQL implementation of the collector repository.
package collector

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

// uniqueViolation is the SQLSTATE of a violated unique or primary key constraint.
const uniqueViolation = "23505"

var _ port.CollectorRepository = (*Repository)(nil)

// Repository is the Postgres implementation of port.CollectorRepository.
type Repository struct {
	db pgdb.DBTX
}

// NewRepository returns a repository backed by db, the pool or a transaction.
func NewRepository(db pgdb.DBTX) *Repository {
	return &Repository{db: db}
}

// GetCollector implements port.CollectorRepository.
func (r *Repository) GetCollector(ctx context.Context, id domain.CollectorID) (*domain.Collector, error) {
	var row collectorRow

	err := r.db.QueryRow(ctx, `SELECT id, name, email FROM collectors WHERE id = $1`, string(id)).
		Scan(&row.ID, &row.Name, &row.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCollectorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query collector: %w", err)
	}

	return row.toDomain()
}

// SaveCollector implements port.CollectorRepository.
func (r *Repository) SaveCollector(ctx context.Context, collector *domain.Collector) error {
	row := newCollectorRow(collector)

	_, err := r.db.Exec(ctx, `
		INSERT INTO collectors (id, name, email) VALUES ($1, $2, $3)`,
		row.ID, row.Name, row.Email)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return domain.ErrCollectorIDTaken
	}

	if err != nil {
		return fmt.Errorf("insert collector: %w", err)
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
	return collectorRow{ID: string(c.ID()), Name: c.Name(), Email: string(c.Email())}
}

func (r collectorRow) toDomain() (*domain.Collector, error) {
	c, err := domain.RehydrateCollector(domain.CollectorID(r.ID), r.Name, domain.Email(r.Email))
	if err != nil {
		return nil, port.CorruptData("collector "+r.ID, err)
	}

	return c, nil
}
