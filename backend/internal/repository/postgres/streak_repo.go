package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type StreakRepo struct {
	db *DB
}

func NewStreakRepo(db *DB) *StreakRepo {
	return &StreakRepo{db: db}
}

// FindByUserID returns the streak record for a user, or nil if none exists.
func (r *StreakRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.Streak, error) {
	query := `
		SELECT id, user_id, current_streak, longest_streak, last_played_date,
		       bonus_skips, bonus_answers, created_at, updated_at
		FROM streaks
		WHERE user_id = $1`

	var s domain.Streak
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.CurrentStreak, &s.LongestStreak, &s.LastPlayedDate,
		&s.BonusSkips, &s.BonusAnswers, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find streak by user_id: %w", err)
	}
	return &s, nil
}

// FindStreaksAtRisk returns all streaks where the user played yesterday but not today,
// meaning their streak will break if they don't play before midnight.
func (r *StreakRepo) FindStreaksAtRisk(ctx context.Context) ([]domain.Streak, error) {
	query := `
		SELECT id, user_id, current_streak, longest_streak, last_played_date,
		       bonus_skips, bonus_answers, created_at, updated_at
		FROM streaks
		WHERE current_streak > 0
		  AND last_played_date::date = (CURRENT_DATE - INTERVAL '1 day')::date`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find streaks at risk: %w", err)
	}
	defer rows.Close()

	var streaks []domain.Streak
	for rows.Next() {
		var s domain.Streak
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.CurrentStreak, &s.LongestStreak, &s.LastPlayedDate,
			&s.BonusSkips, &s.BonusAnswers, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan streak at risk: %w", err)
		}
		streaks = append(streaks, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate streaks at risk: %w", err)
	}
	return streaks, nil
}

// Upsert inserts a new streak record or updates the existing one for the user.
func (r *StreakRepo) Upsert(ctx context.Context, s *domain.Streak) error {
	query := `
		INSERT INTO streaks (id, user_id, current_streak, longest_streak, last_played_date,
		                     bonus_skips, bonus_answers, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			current_streak  = EXCLUDED.current_streak,
			longest_streak  = EXCLUDED.longest_streak,
			last_played_date = EXCLUDED.last_played_date,
			bonus_skips     = EXCLUDED.bonus_skips,
			bonus_answers   = EXCLUDED.bonus_answers,
			updated_at      = NOW()
		RETURNING id, created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		s.ID, s.UserID, s.CurrentStreak, s.LongestStreak, s.LastPlayedDate,
		s.BonusSkips, s.BonusAnswers,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert streak: %w", err)
	}
	return nil
}
