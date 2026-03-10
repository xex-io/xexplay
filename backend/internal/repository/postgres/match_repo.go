package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type MatchRepo struct {
	db *DB
}

func NewMatchRepo(db *DB) *MatchRepo {
	return &MatchRepo{db: db}
}

func (r *MatchRepo) Create(ctx context.Context, m *domain.Match) error {
	query := `
		INSERT INTO matches (id, event_id, home_team, away_team, kickoff_time,
		                      status, home_score, away_score, result_data,
		                      created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		m.ID, m.EventID, m.HomeTeam, m.AwayTeam, m.KickoffTime,
		m.Status, m.HomeScore, m.AwayScore, m.ResultData,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create match: %w", err)
	}
	return nil
}

func (r *MatchRepo) Update(ctx context.Context, m *domain.Match) error {
	query := `
		UPDATE matches
		SET event_id = $2, home_team = $3, away_team = $4, kickoff_time = $5,
		    status = $6, home_score = $7, away_score = $8, result_data = $9,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		m.ID, m.EventID, m.HomeTeam, m.AwayTeam, m.KickoffTime,
		m.Status, m.HomeScore, m.AwayScore, m.ResultData,
	).Scan(&m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update match: %w", err)
	}
	return nil
}

func (r *MatchRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Match, error) {
	query := `
		SELECT id, event_id, home_team, away_team, kickoff_time,
		       status, home_score, away_score, result_data,
		       created_at, updated_at
		FROM matches
		WHERE id = $1`

	var m domain.Match
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.EventID, &m.HomeTeam, &m.AwayTeam, &m.KickoffTime,
		&m.Status, &m.HomeScore, &m.AwayScore, &m.ResultData,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find match by id: %w", err)
	}
	return &m, nil
}

func (r *MatchRepo) FindByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain.Match, error) {
	query := `
		SELECT id, event_id, home_team, away_team, kickoff_time,
		       status, home_score, away_score, result_data,
		       created_at, updated_at
		FROM matches
		WHERE event_id = $1
		ORDER BY kickoff_time ASC`

	rows, err := r.db.Pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("find matches by event_id: %w", err)
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		var m domain.Match
		if err := rows.Scan(
			&m.ID, &m.EventID, &m.HomeTeam, &m.AwayTeam, &m.KickoffTime,
			&m.Status, &m.HomeScore, &m.AwayScore, &m.ResultData,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan match: %w", err)
		}
		matches = append(matches, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate matches: %w", err)
	}
	return matches, nil
}

func (r *MatchRepo) UpdateResult(ctx context.Context, id uuid.UUID, homeScore, awayScore int) error {
	query := `
		UPDATE matches
		SET home_score = $2, away_score = $3, status = $4, updated_at = NOW()
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, id, homeScore, awayScore, domain.MatchStatusCompleted)
	if err != nil {
		return fmt.Errorf("update match result: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("update match result: match not found")
	}
	return nil
}
