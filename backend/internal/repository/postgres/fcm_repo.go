package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type FCMTokenRepo struct {
	db *DB
}

func NewFCMTokenRepo(db *DB) *FCMTokenRepo {
	return &FCMTokenRepo{db: db}
}

// RegisterToken inserts a new FCM token or reactivates an existing one.
func (r *FCMTokenRepo) RegisterToken(ctx context.Context, token *domain.FCMToken) error {
	query := `
		INSERT INTO fcm_tokens (id, user_id, token, device_type, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		ON CONFLICT (token) DO UPDATE SET
			user_id     = EXCLUDED.user_id,
			device_type = EXCLUDED.device_type,
			is_active   = true,
			updated_at  = NOW()
		RETURNING id, created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		token.ID, token.UserID, token.Token, token.DeviceType,
	).Scan(&token.ID, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return fmt.Errorf("register fcm token: %w", err)
	}
	token.IsActive = true
	return nil
}

// DeactivateToken sets a specific token as inactive for a user.
func (r *FCMTokenRepo) DeactivateToken(ctx context.Context, userID uuid.UUID, tokenValue string) error {
	query := `
		UPDATE fcm_tokens
		SET is_active = false, updated_at = NOW()
		WHERE user_id = $1 AND token = $2`

	ct, err := r.db.Pool.Exec(ctx, query, userID, tokenValue)
	if err != nil {
		return fmt.Errorf("deactivate fcm token: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("token not found")
	}
	return nil
}

// FindActiveByUser returns all active FCM tokens for a user.
func (r *FCMTokenRepo) FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]domain.FCMToken, error) {
	query := `
		SELECT id, user_id, token, device_type, is_active, created_at, updated_at
		FROM fcm_tokens
		WHERE user_id = $1 AND is_active = true
		ORDER BY updated_at DESC`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find active fcm tokens by user: %w", err)
	}
	defer rows.Close()

	return scanFCMTokens(rows)
}

// FindAllActive returns all active FCM tokens (for broadcast notifications).
func (r *FCMTokenRepo) FindAllActive(ctx context.Context) ([]domain.FCMToken, error) {
	query := `
		SELECT id, user_id, token, device_type, is_active, created_at, updated_at
		FROM fcm_tokens
		WHERE is_active = true
		ORDER BY updated_at DESC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all active fcm tokens: %w", err)
	}
	defer rows.Close()

	return scanFCMTokens(rows)
}

// DeactivateByTokenValue deactivates a token by its value (e.g., when FCM reports it as invalid).
func (r *FCMTokenRepo) DeactivateByTokenValue(ctx context.Context, tokenValue string) error {
	query := `
		UPDATE fcm_tokens
		SET is_active = false, updated_at = NOW()
		WHERE token = $1`

	_, err := r.db.Pool.Exec(ctx, query, tokenValue)
	if err != nil {
		return fmt.Errorf("deactivate fcm token by value: %w", err)
	}
	return nil
}

func scanFCMTokens(rows pgx.Rows) ([]domain.FCMToken, error) {
	var tokens []domain.FCMToken
	for rows.Next() {
		var t domain.FCMToken
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Token, &t.DeviceType, &t.IsActive,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan fcm token: %w", err)
		}
		tokens = append(tokens, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fcm tokens: %w", err)
	}
	return tokens, nil
}
