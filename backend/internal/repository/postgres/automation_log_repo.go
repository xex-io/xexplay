package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type AutomationLogRepo struct {
	db *DB
}

func NewAutomationLogRepo(db *DB) *AutomationLogRepo {
	return &AutomationLogRepo{db: db}
}

func (r *AutomationLogRepo) Create(ctx context.Context, log *domain.AutomationLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	query := `
		INSERT INTO automation_logs (id, job_name, status, details, items_processed, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at`

	err := r.db.Pool.QueryRow(ctx, query,
		log.ID, log.JobName, log.Status, log.Details, log.ItemsProcessed, log.ErrorMessage,
	).Scan(&log.CreatedAt)
	if err != nil {
		return fmt.Errorf("create automation log: %w", err)
	}
	return nil
}

func (r *AutomationLogRepo) FindRecent(ctx context.Context, limit int) ([]*domain.AutomationLog, error) {
	query := `
		SELECT id, job_name, status, details, items_processed, error_message, created_at
		FROM automation_logs
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("find recent automation logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AutomationLog
	for rows.Next() {
		var l domain.AutomationLog
		if err := rows.Scan(&l.ID, &l.JobName, &l.Status, &l.Details, &l.ItemsProcessed, &l.ErrorMessage, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan automation log: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}

func (r *AutomationLogRepo) FindByJob(ctx context.Context, jobName string, limit int) ([]*domain.AutomationLog, error) {
	query := `
		SELECT id, job_name, status, details, items_processed, error_message, created_at
		FROM automation_logs
		WHERE job_name = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.Pool.Query(ctx, query, jobName, limit)
	if err != nil {
		return nil, fmt.Errorf("find automation logs by job: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AutomationLog
	for rows.Next() {
		var l domain.AutomationLog
		if err := rows.Scan(&l.ID, &l.JobName, &l.Status, &l.Details, &l.ItemsProcessed, &l.ErrorMessage, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan automation log: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}
