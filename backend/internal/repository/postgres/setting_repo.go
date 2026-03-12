package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type SettingRepo struct {
	db *DB
}

func NewSettingRepo(db *DB) *SettingRepo {
	return &SettingRepo{db: db}
}

func (r *SettingRepo) FindAll(ctx context.Context) ([]*domain.Setting, error) {
	query := `SELECT key, value, COALESCE(description, ''), is_secret, updated_at FROM settings ORDER BY key`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all settings: %w", err)
	}
	defer rows.Close()

	var settings []*domain.Setting
	for rows.Next() {
		var s domain.Setting
		if err := rows.Scan(&s.Key, &s.Value, &s.Description, &s.IsSecret, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		settings = append(settings, &s)
	}
	return settings, rows.Err()
}

func (r *SettingRepo) Get(ctx context.Context, key string) (string, error) {
	query := `SELECT value FROM settings WHERE key = $1`
	var value string
	err := r.db.Pool.QueryRow(ctx, query, key).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return value, nil
}

func (r *SettingRepo) Set(ctx context.Context, key, value string) error {
	query := `
		INSERT INTO settings (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`
	_, err := r.db.Pool.Exec(ctx, query, key, value)
	if err != nil {
		return fmt.Errorf("set setting %s: %w", key, err)
	}
	return nil
}

func (r *SettingRepo) Delete(ctx context.Context, key string) error {
	query := `UPDATE settings SET value = '', updated_at = NOW() WHERE key = $1`
	_, err := r.db.Pool.Exec(ctx, query, key)
	if err != nil {
		return fmt.Errorf("delete setting %s: %w", key, err)
	}
	return nil
}
