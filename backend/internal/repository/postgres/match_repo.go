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

const matchColumns = `id, event_id, home_team, away_team, home_team_i18n, away_team_i18n,
	kickoff_time, status, home_score, away_score, result_data,
	external_id, sport_key, source,
	created_at, updated_at`

type MatchRepo struct {
	db *DB
}

func NewMatchRepo(db *DB) *MatchRepo {
	return &MatchRepo{db: db}
}

func scanMatch(scan func(dest ...interface{}) error) (*domain.Match, error) {
	var m domain.Match
	var extID, sportKey *string
	var homeI18n, awayI18n []byte
	err := scan(
		&m.ID, &m.EventID, &m.HomeTeam, &m.AwayTeam, &homeI18n, &awayI18n,
		&m.KickoffTime, &m.Status, &m.HomeScore, &m.AwayScore, &m.ResultData,
		&extID, &sportKey, &m.Source,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if extID != nil {
		m.ExternalID = *extID
	}
	if sportKey != nil {
		m.SportKey = *sportKey
	}
	if len(homeI18n) > 0 {
		_ = json.Unmarshal(homeI18n, &m.HomeTeamI18n)
	}
	if len(awayI18n) > 0 {
		_ = json.Unmarshal(awayI18n, &m.AwayTeamI18n)
	}
	return &m, err
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func i18nJSON(m map[string]string) []byte {
	if len(m) == 0 {
		return []byte("{}")
	}
	b, _ := json.Marshal(m)
	return b
}

func (r *MatchRepo) Create(ctx context.Context, m *domain.Match) error {
	if m.Source == "" {
		m.Source = "manual"
	}
	query := `
		INSERT INTO matches (id, event_id, home_team, away_team, home_team_i18n, away_team_i18n,
		                      kickoff_time, status, home_score, away_score, result_data,
		                      external_id, sport_key, source,
		                      created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		m.ID, m.EventID, m.HomeTeam, m.AwayTeam, i18nJSON(m.HomeTeamI18n), i18nJSON(m.AwayTeamI18n),
		m.KickoffTime, m.Status, m.HomeScore, m.AwayScore, m.ResultData,
		nilIfEmpty(m.ExternalID), nilIfEmpty(m.SportKey), m.Source,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create match: %w", err)
	}
	return nil
}

func (r *MatchRepo) Update(ctx context.Context, m *domain.Match) error {
	query := `
		UPDATE matches
		SET event_id = $2, home_team = $3, away_team = $4,
		    home_team_i18n = $5, away_team_i18n = $6,
		    kickoff_time = $7, status = $8, home_score = $9, away_score = $10,
		    result_data = $11, external_id = $12, sport_key = $13, source = $14,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		m.ID, m.EventID, m.HomeTeam, m.AwayTeam,
		i18nJSON(m.HomeTeamI18n), i18nJSON(m.AwayTeamI18n),
		m.KickoffTime, m.Status, m.HomeScore, m.AwayScore, m.ResultData,
		nilIfEmpty(m.ExternalID), nilIfEmpty(m.SportKey), m.Source,
	).Scan(&m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update match: %w", err)
	}
	return nil
}

func (r *MatchRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Match, error) {
	query := `SELECT ` + matchColumns + ` FROM matches WHERE id = $1`

	m, err := scanMatch(r.db.Pool.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find match by id: %w", err)
	}
	return m, nil
}

func (r *MatchRepo) FindByExternalID(ctx context.Context, externalID string) (*domain.Match, error) {
	query := `SELECT ` + matchColumns + ` FROM matches WHERE external_id = $1`

	m, err := scanMatch(r.db.Pool.QueryRow(ctx, query, externalID).Scan)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find match by external_id: %w", err)
	}
	return m, nil
}

func (r *MatchRepo) scanMatchRows(ctx context.Context, query string, args ...interface{}) ([]*domain.Match, error) {
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		m, err := scanMatch(rows.Scan)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, rows.Err()
}

func (r *MatchRepo) FindByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain.Match, error) {
	query := `SELECT ` + matchColumns + ` FROM matches WHERE event_id = $1 ORDER BY kickoff_time ASC`
	matches, err := r.scanMatchRows(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("find matches by event_id: %w", err)
	}
	return matches, nil
}

func (r *MatchRepo) FindAll(ctx context.Context) ([]*domain.Match, error) {
	query := `SELECT ` + matchColumns + ` FROM matches ORDER BY kickoff_time DESC`
	matches, err := r.scanMatchRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all matches: %w", err)
	}
	return matches, nil
}

// FindScheduledPastKickoff returns auto-sourced matches that are past kickoff but not yet completed.
func (r *MatchRepo) FindScheduledPastKickoff(ctx context.Context) ([]*domain.Match, error) {
	query := `SELECT ` + matchColumns + `
		FROM matches
		WHERE source = 'auto'
		  AND status IN ($1, $2)
		  AND kickoff_time < NOW()
		ORDER BY kickoff_time ASC`
	matches, err := r.scanMatchRows(ctx, query, domain.MatchStatusUpcoming, domain.MatchStatusLive)
	if err != nil {
		return nil, fmt.Errorf("find scheduled past kickoff: %w", err)
	}
	return matches, nil
}

// FindByDateRange returns matches for a sport within a date range.
func (r *MatchRepo) FindByDateRange(ctx context.Context, from, to time.Time, sportKey string) ([]*domain.Match, error) {
	query := `SELECT ` + matchColumns + `
		FROM matches
		WHERE sport_key = $1 AND kickoff_time >= $2 AND kickoff_time < $3
		ORDER BY kickoff_time ASC`
	matches, err := r.scanMatchRows(ctx, query, sportKey, from, to)
	if err != nil {
		return nil, fmt.Errorf("find matches by date range: %w", err)
	}
	return matches, nil
}

// UpsertFromExternal inserts or updates a match by external_id.
func (r *MatchRepo) UpsertFromExternal(ctx context.Context, m *domain.Match) error {
	query := `
		INSERT INTO matches (id, event_id, home_team, away_team, home_team_i18n, away_team_i18n,
		                      kickoff_time, status, home_score, away_score, result_data,
		                      external_id, sport_key, source,
		                      created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
		ON CONFLICT (external_id) DO UPDATE SET
		    home_team = EXCLUDED.home_team,
		    away_team = EXCLUDED.away_team,
		    home_team_i18n = EXCLUDED.home_team_i18n,
		    away_team_i18n = EXCLUDED.away_team_i18n,
		    kickoff_time = EXCLUDED.kickoff_time,
		    status = EXCLUDED.status,
		    home_score = EXCLUDED.home_score,
		    away_score = EXCLUDED.away_score,
		    result_data = EXCLUDED.result_data,
		    updated_at = NOW()
		RETURNING id, created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		m.ID, m.EventID, m.HomeTeam, m.AwayTeam, i18nJSON(m.HomeTeamI18n), i18nJSON(m.AwayTeamI18n),
		m.KickoffTime, m.Status, m.HomeScore, m.AwayScore, m.ResultData,
		m.ExternalID, m.SportKey, m.Source,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert match from external: %w", err)
	}
	return nil
}

func (r *MatchRepo) CountCardsByMatchID(ctx context.Context, matchID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM cards WHERE match_id = $1`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, matchID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count cards by match: %w", err)
	}
	return count, nil
}

func (r *MatchRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM matches WHERE id = $1`
	ct, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete match: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("delete match: match not found")
	}
	return nil
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
