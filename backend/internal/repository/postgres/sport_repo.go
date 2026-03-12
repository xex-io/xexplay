package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type SportRepo struct {
	db *DB
}

func NewSportRepo(db *DB) *SportRepo {
	return &SportRepo{db: db}
}

func (r *SportRepo) FindAll(ctx context.Context) ([]*domain.Sport, error) {
	query := `SELECT key, group_name, title, is_active FROM sports ORDER BY group_name, title`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all sports: %w", err)
	}
	defer rows.Close()

	var sports []*domain.Sport
	for rows.Next() {
		var s domain.Sport
		if err := rows.Scan(&s.Key, &s.Group, &s.Title, &s.IsActive); err != nil {
			return nil, fmt.Errorf("scan sport: %w", err)
		}
		sports = append(sports, &s)
	}
	return sports, rows.Err()
}

func (r *SportRepo) FindActive(ctx context.Context) ([]*domain.Sport, error) {
	query := `SELECT key, group_name, title, is_active FROM sports WHERE is_active = true ORDER BY group_name, title`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find active sports: %w", err)
	}
	defer rows.Close()

	var sports []*domain.Sport
	for rows.Next() {
		var s domain.Sport
		if err := rows.Scan(&s.Key, &s.Group, &s.Title, &s.IsActive); err != nil {
			return nil, fmt.Errorf("scan sport: %w", err)
		}
		sports = append(sports, &s)
	}
	return sports, rows.Err()
}

func (r *SportRepo) FindByKey(ctx context.Context, key string) (*domain.Sport, error) {
	query := `SELECT key, group_name, title, is_active FROM sports WHERE key = $1`

	var s domain.Sport
	err := r.db.Pool.QueryRow(ctx, query, key).Scan(&s.Key, &s.Group, &s.Title, &s.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find sport by key: %w", err)
	}
	return &s, nil
}

func (r *SportRepo) Upsert(ctx context.Context, s *domain.Sport) error {
	query := `
		INSERT INTO sports (key, group_name, title, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO UPDATE SET group_name = $2, title = $3`

	_, err := r.db.Pool.Exec(ctx, query, s.Key, s.Group, s.Title, s.IsActive)
	if err != nil {
		return fmt.Errorf("upsert sport: %w", err)
	}
	return nil
}

func (r *SportRepo) SetActive(ctx context.Context, key string, active bool) error {
	query := `UPDATE sports SET is_active = $2 WHERE key = $1`
	ct, err := r.db.Pool.Exec(ctx, query, key, active)
	if err != nil {
		return fmt.Errorf("set sport active: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("sport not found: %s", key)
	}
	return nil
}
