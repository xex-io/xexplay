package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

const eventColumns = `id, name, slug, description, start_date, end_date,
	is_active, scoring_multiplier, source, sport_key,
	created_at, updated_at`

type EventRepo struct {
	db *DB
}

func NewEventRepo(db *DB) *EventRepo {
	return &EventRepo{db: db}
}

func scanEvent(scan func(dest ...interface{}) error) (*domain.Event, error) {
	var e domain.Event
	var sportKey, source *string
	err := scan(
		&e.ID, &e.Name, &e.Slug, &e.Description, &e.StartDate, &e.EndDate,
		&e.IsActive, &e.ScoringMultiplier, &source, &sportKey,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if sportKey != nil {
		e.SportKey = *sportKey
	}
	if source != nil {
		e.Source = *source
	}
	return &e, err
}

func (r *EventRepo) scanEventRows(ctx context.Context, query string, args ...interface{}) ([]*domain.Event, error) {
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		e, err := scanEvent(rows.Scan)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *EventRepo) Create(ctx context.Context, e *domain.Event) error {
	if e.Source == "" {
		e.Source = "manual"
	}
	query := `
		INSERT INTO events (id, name, slug, description, start_date, end_date,
		                     is_active, scoring_multiplier, source, sport_key,
		                     created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		e.ID, e.Name, e.Slug, e.Description, e.StartDate, e.EndDate,
		e.IsActive, e.ScoringMultiplier, e.Source, nilIfEmpty(e.SportKey),
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
		    is_active = $7, scoring_multiplier = $8, source = $9, sport_key = $10,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		e.ID, e.Name, e.Slug, e.Description, e.StartDate, e.EndDate,
		e.IsActive, e.ScoringMultiplier, e.Source, nilIfEmpty(e.SportKey),
	).Scan(&e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

func (r *EventRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	query := `SELECT ` + eventColumns + ` FROM events WHERE id = $1`

	e, err := scanEvent(r.db.Pool.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find event by id: %w", err)
	}
	return e, nil
}

func (r *EventRepo) FindActive(ctx context.Context) ([]*domain.Event, error) {
	query := `SELECT ` + eventColumns + ` FROM events WHERE is_active = true ORDER BY start_date ASC`
	events, err := r.scanEventRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find active events: %w", err)
	}
	return events, nil
}

func (r *EventRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE events SET is_active = false, updated_at = NOW() WHERE id = $1`
	ct, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete event: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("soft delete event: event not found")
	}
	return nil
}

func (r *EventRepo) FindAll(ctx context.Context) ([]*domain.Event, error) {
	query := `SELECT ` + eventColumns + ` FROM events ORDER BY start_date DESC`
	events, err := r.scanEventRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all events: %w", err)
	}
	return events, nil
}

// FindOrCreateAutoEvent finds an existing auto-event for a sport_key, or creates one.
func (r *EventRepo) FindOrCreateAutoEvent(ctx context.Context, sportKey, leagueName string) (*domain.Event, error) {
	// Try to find existing active auto-event for this sport
	query := `SELECT ` + eventColumns + `
		FROM events
		WHERE sport_key = $1 AND source = 'auto' AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1`

	e, err := scanEvent(r.db.Pool.QueryRow(ctx, query, sportKey).Scan)
	if err == nil {
		return e, nil
	}
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("find auto event: %w", err)
	}

	// Create a new auto-event
	slug := strings.ReplaceAll(strings.ToLower(sportKey), "_", "-") + "-auto"
	nameJSON, _ := json.Marshal(map[string]string{"en": leagueName})
	descJSON, _ := json.Marshal(map[string]string{"en": "Auto-generated event for " + leagueName})

	now := time.Now().UTC()
	newEvent := &domain.Event{
		ID:                uuid.New(),
		Name:              nameJSON,
		Slug:              slug,
		Description:       descJSON,
		StartDate:         now,
		EndDate:           now.AddDate(0, 3, 0), // 3 months from now
		IsActive:          true,
		ScoringMultiplier: 1.0,
		Source:            "auto",
		SportKey:          sportKey,
	}

	if err := r.Create(ctx, newEvent); err != nil {
		return nil, fmt.Errorf("create auto event: %w", err)
	}
	return newEvent, nil
}
