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
		       created_at, updated_at
		FROM users
		WHERE xex_user_id = $1`

	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, xexUserID).Scan(
		&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
		&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
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
		                    created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
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
		       created_at, updated_at
		FROM users
		WHERE id = $1`

	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
		&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
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
		       created_at, updated_at
		FROM users
		WHERE referral_code = $1`

	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&u.ID, &u.XexUserID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.Role,
		&u.ReferralCode, &u.ReferredBy, &u.Language, &u.TotalPoints, &u.IsActive,
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
