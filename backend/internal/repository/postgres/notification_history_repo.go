package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type NotificationHistoryRepo struct {
	db *DB
}

func NewNotificationHistoryRepo(db *DB) *NotificationHistoryRepo {
	return &NotificationHistoryRepo{db: db}
}

// Create inserts a new notification history record.
func (r *NotificationHistoryRepo) Create(ctx context.Context, h *domain.NotificationHistory) error {
	query := `
		INSERT INTO notification_history (id, admin_user_id, title, body, target_type, recipient_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at`

	err := r.db.Pool.QueryRow(ctx, query,
		h.ID, h.AdminUserID, h.Title, h.Body, h.TargetType, h.RecipientCount,
	).Scan(&h.CreatedAt)
	if err != nil {
		return fmt.Errorf("create notification history: %w", err)
	}
	return nil
}

// FindRecent returns recent notification history with pagination.
func (r *NotificationHistoryRepo) FindRecent(ctx context.Context, limit, offset int) ([]domain.NotificationHistory, error) {
	query := `
		SELECT id, admin_user_id, title, body, target_type, recipient_count, created_at
		FROM notification_history
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find recent notification history: %w", err)
	}
	defer rows.Close()

	var results []domain.NotificationHistory
	for rows.Next() {
		var h domain.NotificationHistory
		if err := rows.Scan(
			&h.ID, &h.AdminUserID, &h.Title, &h.Body, &h.TargetType,
			&h.RecipientCount, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification history: %w", err)
		}
		results = append(results, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification history: %w", err)
	}
	return results, nil
}

// Count returns the total number of notification history records.
func (r *NotificationHistoryRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notification_history`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count notification history: %w", err)
	}
	return count, nil
}

// FindByAdmin returns notification history for a specific admin user.
func (r *NotificationHistoryRepo) FindByAdmin(ctx context.Context, adminUserID uuid.UUID, limit, offset int) ([]domain.NotificationHistory, error) {
	query := `
		SELECT id, admin_user_id, title, body, target_type, recipient_count, created_at
		FROM notification_history
		WHERE admin_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, adminUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find notification history by admin: %w", err)
	}
	defer rows.Close()

	var results []domain.NotificationHistory
	for rows.Next() {
		var h domain.NotificationHistory
		if err := rows.Scan(
			&h.ID, &h.AdminUserID, &h.Title, &h.Body, &h.TargetType,
			&h.RecipientCount, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification history: %w", err)
		}
		results = append(results, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification history: %w", err)
	}
	return results, nil
}
