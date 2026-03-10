package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type AuditRepo struct {
	db *DB
}

func NewAuditRepo(db *DB) *AuditRepo {
	return &AuditRepo{db: db}
}

// Create inserts a new audit log entry.
func (r *AuditRepo) Create(ctx context.Context, entry *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, admin_user_id, action, entity_type, entity_id, details, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at`

	details := entry.Details
	if details == nil {
		details = json.RawMessage(`{}`)
	}

	err := r.db.Pool.QueryRow(ctx, query,
		entry.ID, entry.AdminUserID, entry.Action, entry.EntityType,
		entry.EntityID, details, entry.IPAddress,
	).Scan(&entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// FindByAdmin returns audit logs for a specific admin user with pagination.
func (r *AuditRepo) FindByAdmin(ctx context.Context, adminUserID uuid.UUID, limit, offset int) ([]domain.AuditLog, error) {
	query := `
		SELECT id, admin_user_id, action, entity_type, entity_id, details, ip_address, created_at
		FROM audit_logs
		WHERE admin_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, adminUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find audit logs by admin: %w", err)
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

// FindByEntity returns audit logs for a specific entity.
func (r *AuditRepo) FindByEntity(ctx context.Context, entityType, entityID string, limit, offset int) ([]domain.AuditLog, error) {
	query := `
		SELECT id, admin_user_id, action, entity_type, entity_id, details, ip_address, created_at
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Pool.Query(ctx, query, entityType, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find audit logs by entity: %w", err)
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

// FindRecent returns the most recent audit logs.
func (r *AuditRepo) FindRecent(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	query := `
		SELECT id, admin_user_id, action, entity_type, entity_id, details, ip_address, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find recent audit logs: %w", err)
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

func scanAuditLogs(rows pgx.Rows) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(
			&l.ID, &l.AdminUserID, &l.Action, &l.EntityType,
			&l.EntityID, &l.Details, &l.IPAddress, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}
