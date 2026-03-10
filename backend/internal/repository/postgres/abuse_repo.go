package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type AbuseRepo struct {
	db *DB
}

func NewAbuseRepo(db *DB) *AbuseRepo {
	return &AbuseRepo{db: db}
}

// Create inserts a new abuse flag.
func (r *AbuseRepo) Create(ctx context.Context, flag *domain.AbuseFlag) error {
	query := `
		INSERT INTO abuse_flags (id, user_id, flag_type, details, status, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at`

	details := flag.Details
	if details == nil {
		details = json.RawMessage(`{}`)
	}

	err := r.db.Pool.QueryRow(ctx, query,
		flag.ID, flag.UserID, flag.FlagType, details, flag.Status,
	).Scan(&flag.CreatedAt)
	if err != nil {
		return fmt.Errorf("create abuse flag: %w", err)
	}
	return nil
}

// FindPending returns all pending abuse flags, ordered by most recent first.
func (r *AbuseRepo) FindPending(ctx context.Context, limit, offset int) ([]domain.AbuseFlag, error) {
	query := `
		SELECT id, user_id, flag_type, details, status, reviewed_by, reviewed_at, created_at
		FROM abuse_flags
		WHERE status = 'pending'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find pending abuse flags: %w", err)
	}
	defer rows.Close()

	return scanAbuseFlags(rows)
}

// FindByUser returns all abuse flags for a specific user.
func (r *AbuseRepo) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.AbuseFlag, error) {
	query := `
		SELECT id, user_id, flag_type, details, status, reviewed_by, reviewed_at, created_at
		FROM abuse_flags
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find abuse flags by user: %w", err)
	}
	defer rows.Close()

	return scanAbuseFlags(rows)
}

// UpdateStatus updates the status of an abuse flag (reviewed/dismissed).
func (r *AbuseRepo) UpdateStatus(ctx context.Context, flagID, reviewerID uuid.UUID, status string) error {
	now := time.Now().UTC()
	query := `
		UPDATE abuse_flags
		SET status = $2, reviewed_by = $3, reviewed_at = $4
		WHERE id = $1 AND status = 'pending'`

	ct, err := r.db.Pool.Exec(ctx, query, flagID, status, reviewerID, now)
	if err != nil {
		return fmt.Errorf("update abuse flag status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("abuse flag not found or already reviewed")
	}
	return nil
}

// CountByUserAndType counts the number of flags for a user of a specific type.
func (r *AbuseRepo) CountByUserAndType(ctx context.Context, userID uuid.UUID, flagType string) (int, error) {
	query := `SELECT COUNT(*) FROM abuse_flags WHERE user_id = $1 AND flag_type = $2`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, userID, flagType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count abuse flags: %w", err)
	}
	return count, nil
}

func scanAbuseFlags(rows pgx.Rows) ([]domain.AbuseFlag, error) {
	var flags []domain.AbuseFlag
	for rows.Next() {
		var f domain.AbuseFlag
		if err := rows.Scan(
			&f.ID, &f.UserID, &f.FlagType, &f.Details, &f.Status,
			&f.ReviewedBy, &f.ReviewedAt, &f.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan abuse flag: %w", err)
		}
		flags = append(flags, f)
	}
	return flags, nil
}
