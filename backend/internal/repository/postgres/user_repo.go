package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByXexUserID(ctx context.Context, xexUserID uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, xex_user_id, display_name, email, avatar_url, role,
		       referral_code, referred_by, language, total_points, is_active,
		       COALESCE(trading_tier, ''), COALESCE(exchange_status, ''),
		       created_at, updated_at
		FROM users
		WHERE xex_user_id = $1`

	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, xexUserID).Scan(
		&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
		&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
		&u.TradingTier, &u.ExchangeStatus,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by xex_user_id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) Upsert(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, xex_user_id, display_name, email, avatar_url, role,
		                    referral_code, referred_by, language, total_points, is_active,
		                    trading_tier, exchange_status,
		                    created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		ON CONFLICT (xex_user_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			email        = EXCLUDED.email,
			avatar_url   = EXCLUDED.avatar_url,
			language     = EXCLUDED.language,
			updated_at   = NOW()
		RETURNING id, created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		u.ID, u.XexUserID, u.DisplayName, u.Email, u.AvatarURL, u.Role,
		u.ReferralCode, u.ReferredBy, u.Language, u.TotalPoints, u.IsActive,
		u.TradingTier, u.ExchangeStatus,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, xex_user_id, display_name, email, avatar_url, role,
		       referral_code, referred_by, language, total_points, is_active,
		       COALESCE(trading_tier, ''), COALESCE(exchange_status, ''),
		       created_at, updated_at
		FROM users
		WHERE id = $1`

	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
		&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
		&u.TradingTier, &u.ExchangeStatus,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, displayName, avatarURL, language string) error {
	query := `
		UPDATE users
		SET display_name = $2, avatar_url = $3, language = $4, updated_at = NOW()
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, id, displayName, avatarURL, language)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("update user profile: user not found")
	}
	return nil
}

func (r *UserRepo) FindByReferralCode(ctx context.Context, code string) (*domain.User, error) {
	query := `
		SELECT id, xex_user_id, display_name, email, avatar_url, role,
		       referral_code, referred_by, language, total_points, is_active,
		       COALESCE(trading_tier, ''), COALESCE(exchange_status, ''),
		       created_at, updated_at
		FROM users
		WHERE referral_code = $1`

	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
		&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
		&u.TradingTier, &u.ExchangeStatus,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by referral code: %w", err)
	}
	return &u, nil
}

// ListPaginated returns a page of users ordered by creation date.
func (r *UserRepo) ListPaginated(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, xex_user_id, display_name, email, avatar_url, role,
		       referral_code, referred_by, language, total_points, is_active,
		       COALESCE(trading_tier, ''), COALESCE(exchange_status, ''),
		       created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users paginated: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
			&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
			&u.TradingTier, &u.ExchangeStatus,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

// Count returns the total number of users.
func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// UpdateAdmin updates admin-controlled fields (role, is_active) for a user.
func (r *UserRepo) UpdateAdmin(ctx context.Context, id uuid.UUID, role string, isActive bool) error {
	query := `
		UPDATE users
		SET role = $2, is_active = $3, updated_at = NOW()
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, id, role, isActive)
	if err != nil {
		return fmt.Errorf("admin update user: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("admin update user: user not found")
	}
	return nil
}

// UpdateDeviceInfo updates the device fingerprint columns for a user.
func (r *UserRepo) UpdateDeviceInfo(ctx context.Context, id uuid.UUID, deviceID, lastIP string, lastLoginAt time.Time) error {
	query := `
		UPDATE users
		SET device_id = $2, last_ip = $3, last_login_at = $4, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id, deviceID, lastIP, lastLoginAt)
	if err != nil {
		return fmt.Errorf("update device info: %w", err)
	}
	return nil
}

// FindByDeviceIDOrIP returns users that share the given device_id or last_ip, excluding the given userID.
func (r *UserRepo) FindByDeviceIDOrIP(ctx context.Context, userID uuid.UUID, deviceID, ip string) ([]domain.User, error) {
	query := `
		SELECT id, xex_user_id, display_name, email, avatar_url, role,
		       referral_code, referred_by, language, total_points, is_active,
		       COALESCE(trading_tier, ''), COALESCE(exchange_status, ''),
		       created_at, updated_at
		FROM users
		WHERE id != $1
		  AND ((device_id = $2 AND $2 != '') OR (last_ip = $3 AND $3 != ''))`

	rows, err := r.db.Pool.Query(ctx, query, userID, deviceID, ip)
	if err != nil {
		return nil, fmt.Errorf("find users by device_id or ip: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
			&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
			&u.TradingTier, &u.ExchangeStatus,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

// UpdateTradingTier updates the trading tier for a user (set by Exchange sync or admin).
func (r *UserRepo) UpdateTradingTier(ctx context.Context, id uuid.UUID, tradingTier string) error {
	query := `
		UPDATE users
		SET trading_tier = $2, updated_at = NOW()
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, id, tradingTier)
	if err != nil {
		return fmt.Errorf("update trading tier: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("update trading tier: user not found")
	}
	return nil
}

// UpdateExchangeStatus updates the Exchange account status for a user.
func (r *UserRepo) UpdateExchangeStatus(ctx context.Context, id uuid.UUID, exchangeStatus string) error {
	query := `
		UPDATE users
		SET exchange_status = $2, updated_at = NOW()
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, id, exchangeStatus)
	if err != nil {
		return fmt.Errorf("update exchange status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("update exchange status: user not found")
	}
	return nil
}

// Search returns users matching the query by email or display_name (ILIKE).
func (r *UserRepo) Search(ctx context.Context, q string, limit, offset int) ([]domain.User, error) {
	pattern := "%" + q + "%"
	sqlStr := `
		SELECT id, xex_user_id, display_name, email, avatar_url, role,
		       referral_code, referred_by, language, total_points, is_active,
		       COALESCE(trading_tier, ''), COALESCE(exchange_status, ''),
		       created_at, updated_at
		FROM users
		WHERE email ILIKE $1 OR display_name ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, sqlStr, pattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
			&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
			&u.TradingTier, &u.ExchangeStatus,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

// CountActiveUsers returns the count of distinct users with sessions since the given time.
func (r *UserRepo) CountActiveUsers(ctx context.Context, since time.Time) (int, error) {
	sqlStr := `
		SELECT COUNT(DISTINCT user_id)
		FROM user_sessions
		WHERE started_at >= $1`
	var count int
	err := r.db.Pool.QueryRow(ctx, sqlStr, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active users: %w", err)
	}
	return count, nil
}

// CountLinkedExchangeUsers returns the count of users with a non-empty exchange_status.
func (r *UserRepo) CountLinkedExchangeUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE exchange_status != '' AND exchange_status IS NOT NULL`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count linked exchange users: %w", err)
	}
	return count, nil
}

// TradingTierDistribution returns counts grouped by trading tier.
func (r *UserRepo) TradingTierDistribution(ctx context.Context) (map[string]int, error) {
	sqlStr := `
		SELECT COALESCE(NULLIF(trading_tier, ''), 'none') AS tier, COUNT(*)
		FROM users
		GROUP BY tier`
	rows, err := r.db.Pool.Query(ctx, sqlStr)
	if err != nil {
		return nil, fmt.Errorf("trading tier distribution: %w", err)
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var tier string
		var count int
		if err := rows.Scan(&tier, &count); err != nil {
			return nil, fmt.Errorf("scan tier distribution: %w", err)
		}
		dist[tier] = count
	}
	return dist, nil
}

// UpdateIsActive updates only the is_active flag for a user.
func (r *UserRepo) UpdateIsActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	sqlStr := `UPDATE users SET is_active = $2, updated_at = NOW() WHERE id = $1`
	ct, err := r.db.Pool.Exec(ctx, sqlStr, id, isActive)
	if err != nil {
		return fmt.Errorf("update user is_active: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("update user is_active: user not found")
	}
	return nil
}

func (r *UserRepo) GetStats(ctx context.Context, userID uuid.UUID) (*domain.UserStats, error) {
	query := `
		SELECT
			COALESCE(u.total_points, 0),
			COUNT(DISTINCT s.id),
			COUNT(a.id),
			COUNT(a.id) FILTER (WHERE a.is_correct = true),
			0, 0
		FROM users u
		LEFT JOIN user_sessions s ON s.user_id = u.id
		LEFT JOIN user_answers a  ON a.user_id = u.id
		WHERE u.id = $1
		GROUP BY u.total_points`

	var stats domain.UserStats
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&stats.TotalPoints,
		&stats.TotalSessions,
		&stats.TotalAnswers,
		&stats.CorrectAnswers,
		&stats.CurrentStreak,
		&stats.LongestStreak,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &domain.UserStats{}, nil
		}
		return nil, fmt.Errorf("get user stats: %w", err)
	}
	return &stats, nil
}
