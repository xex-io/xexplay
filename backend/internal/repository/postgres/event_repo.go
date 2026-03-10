package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type EventRepo struct {
	db *DB
}

func NewEventRepo(db *DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Create(ctx context.Context, e *domain.Event) error {
	query := `
		INSERT INTO events (id, name, slug, description, start_date, end_date,
		                     is_active, scoring_multiplier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		e.ID, e.Name, e.Slug, e.Description, e.StartDate, e.EndDate,
		e.IsActive, e.ScoringMultiplier,
	).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}
	return nil
}

func (r *EventRepo) Update(ctx context.Context, e *domain.Event) error {
	query := `
		UPDATE events
		SET name = $2, slug = $3, description = $4, start_date = $5, end_date = $6,
		    is_active = $7, scoring_multiplier = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		e.ID, e.Name, e.Slug, e.Description, e.StartDate, e.EndDate,
		e.IsActive, e.ScoringMultiplier,
	).Scan(&e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

func (r *EventRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	query := `
		SELECT id, name, slug, description, start_date, end_date,
		       is_active, scoring_multiplier, created_at, updated_at
		FROM events
		WHERE id = $1`

	var e domain.Event
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.Name, &e.Slug, &e.Description, &e.StartDate, &e.EndDate,
		&e.IsActive, &e.ScoringMultiplier, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find event by id: %w", err)
	}
	return &e, nil
}

func (r *EventRepo) FindActive(ctx context.Context) ([]*domain.Event, error) {
	query := `
		SELECT id, name, slug, description, start_date, end_date,
		       is_active, scoring_multiplier, created_at, updated_at
		FROM events
		WHERE is_active = true
		ORDER BY start_date ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find active events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.Description, &e.StartDate, &e.EndDate,
			&e.IsActive, &e.ScoringMultiplier, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan active event: %w", err)
		}
		events = append(events, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active events: %w", err)
	}
	return events, nil
}

func (r *EventRepo) FindAll(ctx context.Context) ([]*domain.Event, error) {
	query := `
		SELECT id, name, slug, description, start_date, end_date,
		       is_active, scoring_multiplier, created_at, updated_at
		FROM events
		ORDER BY start_date DESC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.Description, &e.StartDate, &e.EndDate,
			&e.IsActive, &e.ScoringMultiplier, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return events, nil
}
