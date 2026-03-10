package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type LeaderboardRepo struct {
	db *DB
}

func NewLeaderboardRepo(db *DB) *LeaderboardRepo {
	return &LeaderboardRepo{db: db}
}

// UpsertEntry inserts or updates a leaderboard entry, incrementing points and answer counts.
func (r *LeaderboardRepo) UpsertEntry(ctx context.Context, entry *domain.LeaderboardEntry) error {
	query := `
		INSERT INTO leaderboard_entries (id, user_id, event_id, period_type, period_key,
		                                  total_points, correct_answers, wrong_answers, total_answers,
		                                  created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (user_id, period_type, period_key) DO UPDATE SET
			total_points    = leaderboard_entries.total_points + EXCLUDED.total_points,
			correct_answers = leaderboard_entries.correct_answers + EXCLUDED.correct_answers,
			wrong_answers   = leaderboard_entries.wrong_answers + EXCLUDED.wrong_answers,
			total_answers   = leaderboard_entries.total_answers + EXCLUDED.total_answers,
			event_id        = COALESCE(EXCLUDED.event_id, leaderboard_entries.event_id),
			updated_at      = NOW()`

	_, err := r.db.Pool.Exec(ctx, query,
		entry.ID, entry.UserID, entry.EventID, entry.PeriodType, entry.PeriodKey,
		entry.TotalPoints, entry.CorrectAnswers, entry.WrongAnswers, entry.TotalAnswers,
	)
	if err != nil {
		return fmt.Errorf("upsert leaderboard entry: %w", err)
	}
	return nil
}

// GetRanking returns a page of leaderboard rows with user display info, ordered by total_points DESC.
func (r *LeaderboardRepo) GetRanking(ctx context.Context, periodType, periodKey string, limit, offset int) ([]domain.LeaderboardRow, error) {
	query := `
		SELECT
			RANK() OVER (ORDER BY le.total_points DESC) AS rank,
			le.user_id,
			COALESCE(u.display_name, '') AS display_name,
			COALESCE(u.avatar_url, '') AS avatar_url,
			le.total_points,
			le.correct_answers
		FROM leaderboard_entries le
		JOIN users u ON u.id = le.user_id
		WHERE le.period_type = $1 AND le.period_key = $2
		ORDER BY le.total_points DESC, le.correct_answers DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Pool.Query(ctx, query, periodType, periodKey, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard ranking: %w", err)
	}
	defer rows.Close()

	var result []domain.LeaderboardRow
	for rows.Next() {
		var row domain.LeaderboardRow
		if err := rows.Scan(
			&row.Rank, &row.UserID, &row.DisplayName, &row.AvatarURL,
			&row.TotalPoints, &row.CorrectAnswers,
		); err != nil {
			return nil, fmt.Errorf("scan leaderboard row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate leaderboard rows: %w", err)
	}
	return result, nil
}

// GetUserRank returns the requesting user's rank and stats for a given period.
func (r *LeaderboardRepo) GetUserRank(ctx context.Context, userID uuid.UUID, periodType, periodKey string) (*domain.LeaderboardRow, error) {
	query := `
		SELECT rank, user_id, display_name, avatar_url, total_points, correct_answers
		FROM (
			SELECT
				RANK() OVER (ORDER BY le.total_points DESC) AS rank,
				le.user_id,
				COALESCE(u.display_name, '') AS display_name,
				COALESCE(u.avatar_url, '') AS avatar_url,
				le.total_points,
				le.correct_answers
			FROM leaderboard_entries le
			JOIN users u ON u.id = le.user_id
			WHERE le.period_type = $1 AND le.period_key = $2
		) ranked
		WHERE user_id = $3`

	var row domain.LeaderboardRow
	err := r.db.Pool.QueryRow(ctx, query, periodType, periodKey, userID).Scan(
		&row.Rank, &row.UserID, &row.DisplayName, &row.AvatarURL,
		&row.TotalPoints, &row.CorrectAnswers,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user rank: %w", err)
	}
	return &row, nil
}

// GetTopN returns the top N entries for a given period.
func (r *LeaderboardRepo) GetTopN(ctx context.Context, periodType, periodKey string, n int) ([]domain.LeaderboardRow, error) {
	return r.GetRanking(ctx, periodType, periodKey, n, 0)
}

// GetRankingForUsers returns leaderboard rows filtered to specific user IDs, ranked among themselves.
func (r *LeaderboardRepo) GetRankingForUsers(ctx context.Context, periodType, periodKey string, userIDs []uuid.UUID, limit, offset int) ([]domain.LeaderboardRow, error) {
	if len(userIDs) == 0 {
		return []domain.LeaderboardRow{}, nil
	}

	query := `
		SELECT
			RANK() OVER (ORDER BY le.total_points DESC) AS rank,
			le.user_id,
			COALESCE(u.display_name, '') AS display_name,
			COALESCE(u.avatar_url, '') AS avatar_url,
			le.total_points,
			le.correct_answers
		FROM leaderboard_entries le
		JOIN users u ON u.id = le.user_id
		WHERE le.period_type = $1 AND le.period_key = $2
		  AND le.user_id = ANY($3)
		ORDER BY le.total_points DESC, le.correct_answers DESC
		LIMIT $4 OFFSET $5`

	rows, err := r.db.Pool.Query(ctx, query, periodType, periodKey, userIDs, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard ranking for users: %w", err)
	}
	defer rows.Close()

	var result []domain.LeaderboardRow
	for rows.Next() {
		var row domain.LeaderboardRow
		if err := rows.Scan(
			&row.Rank, &row.UserID, &row.DisplayName, &row.AvatarURL,
			&row.TotalPoints, &row.CorrectAnswers,
		); err != nil {
			return nil, fmt.Errorf("scan leaderboard row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate leaderboard rows: %w", err)
	}
	if result == nil {
		result = []domain.LeaderboardRow{}
	}
	return result, nil
}

// CountEntries returns the total number of entries for a given period.
func (r *LeaderboardRepo) CountEntries(ctx context.Context, periodType, periodKey string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM leaderboard_entries
		WHERE period_type = $1 AND period_key = $2`

	var count int
	err := r.db.Pool.QueryRow(ctx, query, periodType, periodKey).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count leaderboard entries: %w", err)
	}
	return count, nil
}
