package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type SessionRepo struct {
	db *DB
}

func NewSessionRepo(db *DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, s *domain.UserSession) error {
	query := `
		INSERT INTO user_sessions (id, user_id, basket_id, shuffle_order, current_index,
		                            answers_used, skips_used, bonus_answers, bonus_skips,
		                            status, card_presented_at, started_at, completed_at,
		                            created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		s.ID, s.UserID, s.BasketID, s.ShuffleOrder, s.CurrentIndex,
		s.AnswersUsed, s.SkipsUsed, s.BonusAnswers, s.BonusSkips,
		s.Status, s.CardPresentedAt, s.StartedAt, s.CompletedAt,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *SessionRepo) FindByUserAndBasket(ctx context.Context, userID, basketID uuid.UUID) (*domain.UserSession, error) {
	query := `
		SELECT id, user_id, basket_id, shuffle_order, current_index,
		       answers_used, skips_used, bonus_answers, bonus_skips,
		       status, card_presented_at, started_at, completed_at, created_at, updated_at
		FROM user_sessions
		WHERE user_id = $1 AND basket_id = $2`

	var s domain.UserSession
	err := r.db.Pool.QueryRow(ctx, query, userID, basketID).Scan(
		&s.ID, &s.UserID, &s.BasketID, &s.ShuffleOrder, &s.CurrentIndex,
		&s.AnswersUsed, &s.SkipsUsed, &s.BonusAnswers, &s.BonusSkips,
		&s.Status, &s.CardPresentedAt, &s.StartedAt, &s.CompletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find session by user and basket: %w", err)
	}
	return &s, nil
}

func (r *SessionRepo) UpdateProgress(ctx context.Context, s *domain.UserSession) error {
	query := `
		UPDATE user_sessions
		SET current_index = $2, answers_used = $3, skips_used = $4,
		    bonus_answers = $5, bonus_skips = $6, status = $7,
		    card_presented_at = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		s.ID, s.CurrentIndex, s.AnswersUsed, s.SkipsUsed,
		s.BonusAnswers, s.BonusSkips, s.Status, s.CardPresentedAt,
	).Scan(&s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update session progress: %w", err)
	}
	return nil
}

func (r *SessionRepo) Complete(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE user_sessions
		SET status = $2, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, sessionID, domain.SessionStatusCompleted)
	if err != nil {
		return fmt.Errorf("complete session: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("complete session: session not found")
	}
	return nil
}

func (r *SessionRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.UserSession, error) {
	query := `
		SELECT id, user_id, basket_id, shuffle_order, current_index,
		       answers_used, skips_used, bonus_answers, bonus_skips,
		       status, card_presented_at, started_at, completed_at, created_at, updated_at
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("find sessions by user_id: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.UserSession
	for rows.Next() {
		var s domain.UserSession
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.BasketID, &s.ShuffleOrder, &s.CurrentIndex,
			&s.AnswersUsed, &s.SkipsUsed, &s.BonusAnswers, &s.BonusSkips,
			&s.Status, &s.CardPresentedAt, &s.StartedAt, &s.CompletedAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}
	return sessions, nil
}
