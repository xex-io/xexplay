package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type ReferralRepo struct {
	db *DB
}

func NewReferralRepo(db *DB) *ReferralRepo {
	return &ReferralRepo{db: db}
}

// Create inserts a new referral record.
func (r *ReferralRepo) Create(ctx context.Context, ref *domain.Referral) error {
	query := `
		INSERT INTO referrals (id, referrer_id, referred_id, status, reward_granted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		ref.ID, ref.ReferrerID, ref.ReferredID, ref.Status, ref.RewardGranted,
	).Scan(&ref.CreatedAt, &ref.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create referral: %w", err)
	}
	return nil
}

// UpdateStatus updates the status and reward_granted fields of a referral.
func (r *ReferralRepo) UpdateStatus(ctx context.Context, referredID uuid.UUID, status string, rewardGranted bool) error {
	query := `
		UPDATE referrals
		SET status = $2, reward_granted = $3, updated_at = NOW()
		WHERE referred_id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, referredID, status, rewardGranted)
	if err != nil {
		return fmt.Errorf("update referral status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("referral not found for referred_id %s", referredID)
	}
	return nil
}

// CountByReferrer returns the total number of referrals and completed referrals for a user.
func (r *ReferralRepo) CountByReferrer(ctx context.Context, referrerID uuid.UUID) (*domain.ReferralStats, error) {
	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'first_session') AS completed
		FROM referrals
		WHERE referrer_id = $1`

	var stats domain.ReferralStats
	err := r.db.Pool.QueryRow(ctx, query, referrerID).Scan(&stats.TotalReferrals, &stats.CompletedReferrals)
	if err != nil {
		return nil, fmt.Errorf("count referrals by referrer: %w", err)
	}
	return &stats, nil
}

// FindByReferrer returns all referrals where the user is the referrer.
func (r *ReferralRepo) FindByReferrer(ctx context.Context, referrerID uuid.UUID) ([]domain.Referral, error) {
	query := `
		SELECT id, referrer_id, referred_id, status, reward_granted, created_at, updated_at
		FROM referrals
		WHERE referrer_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Pool.Query(ctx, query, referrerID)
	if err != nil {
		return nil, fmt.Errorf("find referrals by referrer: %w", err)
	}
	defer rows.Close()

	var referrals []domain.Referral
	for rows.Next() {
		var ref domain.Referral
		if err := rows.Scan(
			&ref.ID, &ref.ReferrerID, &ref.ReferredID, &ref.Status,
			&ref.RewardGranted, &ref.CreatedAt, &ref.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan referral: %w", err)
		}
		referrals = append(referrals, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate referrals: %w", err)
	}
	return referrals, nil
}

// CountAll returns the total number of referrals.
func (r *ReferralRepo) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM referrals`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count all referrals: %w", err)
	}
	return count, nil
}

// CountConverted returns referrals that reached first_session status.
func (r *ReferralRepo) CountConverted(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM referrals WHERE status = 'first_session'`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count converted referrals: %w", err)
	}
	return count, nil
}

// CountActiveReferrers returns the number of distinct users who have at least one referral.
func (r *ReferralRepo) CountActiveReferrers(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(DISTINCT referrer_id) FROM referrals`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active referrers: %w", err)
	}
	return count, nil
}

// TopReferrers returns the top referrers by referral count.
func (r *ReferralRepo) TopReferrers(ctx context.Context, limit int) ([]domain.TopReferrer, error) {
	query := `
		SELECT r.referrer_id, u.display_name, u.email, COUNT(*) AS referral_count,
		       COUNT(*) FILTER (WHERE r.status = 'first_session') AS converted_count
		FROM referrals r
		JOIN users u ON u.id = r.referrer_id
		GROUP BY r.referrer_id, u.display_name, u.email
		ORDER BY referral_count DESC
		LIMIT $1`

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("top referrers: %w", err)
	}
	defer rows.Close()

	var results []domain.TopReferrer
	for rows.Next() {
		var t domain.TopReferrer
		if err := rows.Scan(&t.UserID, &t.DisplayName, &t.Email, &t.ReferralCount, &t.ConvertedCount); err != nil {
			return nil, fmt.Errorf("scan top referrer: %w", err)
		}
		results = append(results, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top referrers: %w", err)
	}
	return results, nil
}

// FindByReferred returns the referral record for a referred user, or nil if none exists.
func (r *ReferralRepo) FindByReferred(ctx context.Context, referredID uuid.UUID) (*domain.Referral, error) {
	query := `
		SELECT id, referrer_id, referred_id, status, reward_granted, created_at, updated_at
		FROM referrals
		WHERE referred_id = $1`

	var ref domain.Referral
	err := r.db.Pool.QueryRow(ctx, query, referredID).Scan(
		&ref.ID, &ref.ReferrerID, &ref.ReferredID, &ref.Status,
		&ref.RewardGranted, &ref.CreatedAt, &ref.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find referral by referred: %w", err)
	}
	return &ref, nil
}
