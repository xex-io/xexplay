package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type PrizePoolRepo struct {
	db *DB
}

func NewPrizePoolRepo(db *DB) *PrizePoolRepo {
	return &PrizePoolRepo{db: db}
}

// Create inserts a new prize pool.
func (r *PrizePoolRepo) Create(ctx context.Context, p *domain.PrizePool) error {
	query := `
		INSERT INTO prize_pools (id, name, description, total_amount, currency, status,
		                          start_date, end_date, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		p.ID, p.Name, p.Description, p.TotalAmount, p.Currency, p.Status,
		p.StartDate, p.EndDate, p.CreatedBy,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create prize pool: %w", err)
	}
	return nil
}

// FindByID returns a single prize pool by ID.
func (r *PrizePoolRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.PrizePool, error) {
	query := `
		SELECT id, name, description, total_amount, currency, status,
		       start_date, end_date, created_by, created_at, updated_at
		FROM prize_pools
		WHERE id = $1`

	var p domain.PrizePool
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.TotalAmount, &p.Currency, &p.Status,
		&p.StartDate, &p.EndDate, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find prize pool by id: %w", err)
	}
	return &p, nil
}

// Update updates an existing prize pool.
func (r *PrizePoolRepo) Update(ctx context.Context, p *domain.PrizePool) error {
	query := `
		UPDATE prize_pools
		SET name = $2, description = $3, total_amount = $4, currency = $5, status = $6,
		    start_date = $7, end_date = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		p.ID, p.Name, p.Description, p.TotalAmount, p.Currency, p.Status,
		p.StartDate, p.EndDate,
	).Scan(&p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update prize pool: %w", err)
	}
	return nil
}

// UpdateStatus updates only the status of a prize pool.
func (r *PrizePoolRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE prize_pools SET status = $2, updated_at = NOW() WHERE id = $1`
	ct, err := r.db.Pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("update prize pool status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("update prize pool status: pool not found")
	}
	return nil
}

// FindAll returns all prize pools with pagination.
func (r *PrizePoolRepo) FindAll(ctx context.Context, limit, offset int) ([]domain.PrizePool, error) {
	query := `
		SELECT id, name, description, total_amount, currency, status,
		       start_date, end_date, created_by, created_at, updated_at
		FROM prize_pools
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find all prize pools: %w", err)
	}
	defer rows.Close()

	return scanPrizePools(rows)
}

// FindByStatus returns prize pools filtered by status.
func (r *PrizePoolRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]domain.PrizePool, error) {
	query := `
		SELECT id, name, description, total_amount, currency, status,
		       start_date, end_date, created_by, created_at, updated_at
		FROM prize_pools
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find prize pools by status: %w", err)
	}
	defer rows.Close()

	return scanPrizePools(rows)
}

// FindDistributionsByPool returns distributions for a specific prize pool.
func (r *PrizePoolRepo) FindDistributionsByPool(ctx context.Context, poolID string, limit, offset int) ([]domain.PrizePoolDistribution, error) {
	query := `
		SELECT id, prize_pool_id, user_id, amount, rank, status, distributed_at
		FROM prize_pool_distributions
		WHERE prize_pool_id = $1
		ORDER BY rank ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, poolID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find distributions by pool: %w", err)
	}
	defer rows.Close()

	var dists []domain.PrizePoolDistribution
	for rows.Next() {
		var d domain.PrizePoolDistribution
		if err := rows.Scan(
			&d.ID, &d.PrizePoolID, &d.UserID, &d.Amount, &d.Rank, &d.Status, &d.DistributedAt,
		); err != nil {
			return nil, fmt.Errorf("scan prize pool distribution: %w", err)
		}
		dists = append(dists, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate prize pool distributions: %w", err)
	}
	return dists, nil
}

// FindAllDistributions returns all distributions with pagination (history view).
func (r *PrizePoolRepo) FindAllDistributions(ctx context.Context, limit, offset int) ([]domain.PrizePoolDistribution, error) {
	query := `
		SELECT id, prize_pool_id, user_id, amount, rank, status, distributed_at
		FROM prize_pool_distributions
		ORDER BY distributed_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find all distributions: %w", err)
	}
	defer rows.Close()

	var dists []domain.PrizePoolDistribution
	for rows.Next() {
		var d domain.PrizePoolDistribution
		if err := rows.Scan(
			&d.ID, &d.PrizePoolID, &d.UserID, &d.Amount, &d.Rank, &d.Status, &d.DistributedAt,
		); err != nil {
			return nil, fmt.Errorf("scan prize pool distribution: %w", err)
		}
		dists = append(dists, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate prize pool distributions: %w", err)
	}
	return dists, nil
}

func scanPrizePools(rows pgx.Rows) ([]domain.PrizePool, error) {
	var pools []domain.PrizePool
	for rows.Next() {
		var p domain.PrizePool
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.TotalAmount, &p.Currency, &p.Status,
			&p.StartDate, &p.EndDate, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan prize pool: %w", err)
		}
		pools = append(pools, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate prize pools: %w", err)
	}
	return pools, nil
}
