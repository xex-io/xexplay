package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type RewardRepo struct {
	db *DB
}

func NewRewardRepo(db *DB) *RewardRepo {
	return &RewardRepo{db: db}
}

// --- RewardConfig ---

// CreateConfig inserts a new reward config.
func (r *RewardRepo) CreateConfig(ctx context.Context, cfg *domain.RewardConfig) error {
	query := `
		INSERT INTO reward_configs (id, period_type, rank_from, rank_to, reward_type,
		                            amount, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		cfg.ID, cfg.PeriodType, cfg.RankFrom, cfg.RankTo, cfg.RewardType,
		cfg.Amount, cfg.Description, cfg.IsActive,
	).Scan(&cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create reward config: %w", err)
	}
	return nil
}

// UpdateConfig updates an existing reward config.
func (r *RewardRepo) UpdateConfig(ctx context.Context, cfg *domain.RewardConfig) error {
	query := `
		UPDATE reward_configs
		SET period_type = $2, rank_from = $3, rank_to = $4, reward_type = $5,
		    amount = $6, description = $7, is_active = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		cfg.ID, cfg.PeriodType, cfg.RankFrom, cfg.RankTo, cfg.RewardType,
		cfg.Amount, cfg.Description, cfg.IsActive,
	).Scan(&cfg.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("reward config not found")
		}
		return fmt.Errorf("update reward config: %w", err)
	}
	return nil
}

// FindActiveConfigs returns all active reward configs for a given period type.
func (r *RewardRepo) FindActiveConfigs(ctx context.Context, periodType string) ([]domain.RewardConfig, error) {
	query := `
		SELECT id, period_type, rank_from, rank_to, reward_type,
		       amount, description, is_active, created_at, updated_at
		FROM reward_configs
		WHERE is_active = true AND period_type = $1
		ORDER BY rank_from ASC`

	rows, err := r.db.Pool.Query(ctx, query, periodType)
	if err != nil {
		return nil, fmt.Errorf("find active configs: %w", err)
	}
	defer rows.Close()

	var configs []domain.RewardConfig
	for rows.Next() {
		var c domain.RewardConfig
		if err := rows.Scan(
			&c.ID, &c.PeriodType, &c.RankFrom, &c.RankTo, &c.RewardType,
			&c.Amount, &c.Description, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reward config: %w", err)
		}
		configs = append(configs, c)
	}
	return configs, nil
}

// FindConfigsByPeriod returns all active configs for a given period type (alias for FindActiveConfigs).
func (r *RewardRepo) FindConfigsByPeriod(ctx context.Context, periodType string) ([]domain.RewardConfig, error) {
	return r.FindActiveConfigs(ctx, periodType)
}

// FindAllConfigs returns all reward configs (active and inactive).
func (r *RewardRepo) FindAllConfigs(ctx context.Context) ([]domain.RewardConfig, error) {
	query := `
		SELECT id, period_type, rank_from, rank_to, reward_type,
		       amount, description, is_active, created_at, updated_at
		FROM reward_configs
		ORDER BY period_type, rank_from ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all configs: %w", err)
	}
	defer rows.Close()

	var configs []domain.RewardConfig
	for rows.Next() {
		var c domain.RewardConfig
		if err := rows.Scan(
			&c.ID, &c.PeriodType, &c.RankFrom, &c.RankTo, &c.RewardType,
			&c.Amount, &c.Description, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reward config: %w", err)
		}
		configs = append(configs, c)
	}
	return configs, nil
}

// --- RewardDistribution ---

// CreateDistribution inserts a new reward distribution record.
func (r *RewardRepo) CreateDistribution(ctx context.Context, dist *domain.RewardDistribution) error {
	query := `
		INSERT INTO reward_distributions (id, user_id, reward_config_id, period_type, period_key,
		                                   reward_type, amount, rank, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		dist.ID, dist.UserID, dist.RewardConfigID, dist.PeriodType, dist.PeriodKey,
		dist.RewardType, dist.Amount, dist.Rank, dist.Status,
	).Scan(&dist.CreatedAt, &dist.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create reward distribution: %w", err)
	}
	return nil
}

// FindPendingByUser returns all pending reward distributions for a user.
func (r *RewardRepo) FindPendingByUser(ctx context.Context, userID uuid.UUID) ([]domain.RewardDistribution, error) {
	query := `
		SELECT id, user_id, reward_config_id, period_type, period_key,
		       reward_type, amount, rank, status, claimed_at, credited_at,
		       created_at, updated_at
		FROM reward_distributions
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY created_at DESC`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find pending rewards: %w", err)
	}
	defer rows.Close()

	return scanDistributions(rows)
}

// FindByUser returns reward distributions for a user with pagination.
func (r *RewardRepo) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.RewardDistribution, error) {
	query := `
		SELECT id, user_id, reward_config_id, period_type, period_key,
		       reward_type, amount, rank, status, claimed_at, credited_at,
		       created_at, updated_at
		FROM reward_distributions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find user rewards: %w", err)
	}
	defer rows.Close()

	return scanDistributions(rows)
}

// FindDistributionByID returns a single reward distribution by its ID.
func (r *RewardRepo) FindDistributionByID(ctx context.Context, distributionID uuid.UUID) (*domain.RewardDistribution, error) {
	query := `
		SELECT id, user_id, reward_config_id, period_type, period_key,
		       reward_type, amount, rank, status, claimed_at, credited_at,
		       created_at, updated_at
		FROM reward_distributions
		WHERE id = $1`

	var d domain.RewardDistribution
	err := r.db.Pool.QueryRow(ctx, query, distributionID).Scan(
		&d.ID, &d.UserID, &d.RewardConfigID, &d.PeriodType, &d.PeriodKey,
		&d.RewardType, &d.Amount, &d.Rank, &d.Status, &d.ClaimedAt, &d.CreditedAt,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find distribution by id: %w", err)
	}
	return &d, nil
}

// ClaimReward updates a pending reward distribution to claimed status.
func (r *RewardRepo) ClaimReward(ctx context.Context, distributionID, userID uuid.UUID) error {
	query := `
		UPDATE reward_distributions
		SET status = 'claimed', claimed_at = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'pending'`

	now := time.Now().UTC()
	ct, err := r.db.Pool.Exec(ctx, query, distributionID, userID, now)
	if err != nil {
		return fmt.Errorf("claim reward: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("reward not found or already claimed")
	}
	return nil
}

// CreditReward updates a claimed reward distribution to credited status.
func (r *RewardRepo) CreditReward(ctx context.Context, distributionID, userID uuid.UUID) error {
	query := `
		UPDATE reward_distributions
		SET status = 'credited', credited_at = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'claimed'`

	now := time.Now().UTC()
	ct, err := r.db.Pool.Exec(ctx, query, distributionID, userID, now)
	if err != nil {
		return fmt.Errorf("credit reward: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("reward not found or not in claimed status")
	}
	return nil
}

// FindDistributionHistory returns all distributions with pagination (for admin).
func (r *RewardRepo) FindDistributionHistory(ctx context.Context, limit, offset int) ([]domain.RewardDistribution, error) {
	query := `
		SELECT id, user_id, reward_config_id, period_type, period_key,
		       reward_type, amount, rank, status, claimed_at, credited_at,
		       created_at, updated_at
		FROM reward_distributions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find distribution history: %w", err)
	}
	defer rows.Close()

	return scanDistributions(rows)
}

func scanDistributions(rows pgx.Rows) ([]domain.RewardDistribution, error) {
	var dists []domain.RewardDistribution
	for rows.Next() {
		var d domain.RewardDistribution
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.RewardConfigID, &d.PeriodType, &d.PeriodKey,
			&d.RewardType, &d.Amount, &d.Rank, &d.Status, &d.ClaimedAt, &d.CreditedAt,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reward distribution: %w", err)
		}
		dists = append(dists, d)
	}
	return dists, nil
}
