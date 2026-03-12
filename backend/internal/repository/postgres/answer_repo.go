package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type AnswerRepo struct {
	db *DB
}

func NewAnswerRepo(db *DB) *AnswerRepo {
	return &AnswerRepo{db: db}
}

func (r *AnswerRepo) Create(ctx context.Context, a *domain.UserAnswer) error {
	query := `
		INSERT INTO user_answers (id, session_id, card_id, user_id, answer,
		                           points_earned, is_correct, answered_at, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Pool.Exec(ctx, query,
		a.ID, a.SessionID, a.CardID, a.UserID, a.Answer,
		a.PointsEarned, a.IsCorrect, a.AnsweredAt, a.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("create answer: %w", err)
	}
	return nil
}

func (r *AnswerRepo) FindBySession(ctx context.Context, sessionID uuid.UUID) ([]*domain.UserAnswer, error) {
	query := `
		SELECT id, session_id, card_id, user_id, answer,
		       points_earned, is_correct, answered_at, resolved_at
		FROM user_answers
		WHERE session_id = $1
		ORDER BY answered_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("find answers by session: %w", err)
	}
	defer rows.Close()

	var answers []*domain.UserAnswer
	for rows.Next() {
		var a domain.UserAnswer
		if err := rows.Scan(
			&a.ID, &a.SessionID, &a.CardID, &a.UserID, &a.Answer,
			&a.PointsEarned, &a.IsCorrect, &a.AnsweredAt, &a.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("scan answer: %w", err)
		}
		answers = append(answers, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate answers: %w", err)
	}
	return answers, nil
}

func (r *AnswerRepo) FindByCard(ctx context.Context, cardID uuid.UUID) ([]*domain.UserAnswer, error) {
	query := `
		SELECT id, session_id, card_id, user_id, answer,
		       points_earned, is_correct, answered_at, resolved_at
		FROM user_answers
		WHERE card_id = $1
		ORDER BY answered_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, cardID)
	if err != nil {
		return nil, fmt.Errorf("find answers by card: %w", err)
	}
	defer rows.Close()

	var answers []*domain.UserAnswer
	for rows.Next() {
		var a domain.UserAnswer
		if err := rows.Scan(
			&a.ID, &a.SessionID, &a.CardID, &a.UserID, &a.Answer,
			&a.PointsEarned, &a.IsCorrect, &a.AnsweredAt, &a.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("scan answer: %w", err)
		}
		answers = append(answers, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate answers: %w", err)
	}
	return answers, nil
}

// FindByUserID returns recent answers for a specific user.
func (r *AnswerRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.UserAnswer, error) {
	query := `
		SELECT id, session_id, card_id, user_id, answer,
		       points_earned, is_correct, answered_at, resolved_at
		FROM user_answers
		WHERE user_id = $1
		ORDER BY answered_at DESC
		LIMIT $2`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("find answers by user_id: %w", err)
	}
	defer rows.Close()

	var answers []*domain.UserAnswer
	for rows.Next() {
		var a domain.UserAnswer
		if err := rows.Scan(
			&a.ID, &a.SessionID, &a.CardID, &a.UserID, &a.Answer,
			&a.PointsEarned, &a.IsCorrect, &a.AnsweredAt, &a.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("scan answer: %w", err)
		}
		answers = append(answers, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate answers: %w", err)
	}
	return answers, nil
}

// CountCorrectIncorrect returns the total correct and incorrect answer counts.
func (r *AnswerRepo) CountCorrectIncorrect(ctx context.Context) (correct int, incorrect int, err error) {
	sqlStr := `
		SELECT
			COUNT(*) FILTER (WHERE is_correct = true),
			COUNT(*) FILTER (WHERE is_correct = false)
		FROM user_answers
		WHERE is_correct IS NOT NULL`

	err = r.db.Pool.QueryRow(ctx, sqlStr).Scan(&correct, &incorrect)
	if err != nil {
		return 0, 0, fmt.Errorf("count correct/incorrect answers: %w", err)
	}
	return correct, incorrect, nil
}

// BulkResolve resolves all answers for a given card. It fetches the card to determine
// tier-based scoring, then updates each answer's is_correct, points_earned, and resolved_at.
// It also adds earned points to each user's total_points.
func (r *AnswerRepo) BulkResolve(ctx context.Context, cardID uuid.UUID, correctAnswer bool) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx for bulk resolve: %w", err)
	}
	defer tx.Rollback(ctx)

	// Fetch the card to determine tier and scoring
	cardQuery := `
		SELECT id, match_id, question_text, tier, high_answer_is_yes,
		       correct_answer, is_resolved, available_date, expires_at,
		       source, ai_prompt_data, resolution_criteria,
		       created_at, updated_at
		FROM cards
		WHERE id = $1`

	card, err := scanCard(tx.QueryRow(ctx, cardQuery, cardID).Scan)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("bulk resolve: card not found")
		}
		return fmt.Errorf("bulk resolve: fetch card: %w", err)
	}

	// Fetch all unresolved answers for this card
	answersQuery := `
		SELECT id, user_id, answer
		FROM user_answers
		WHERE card_id = $1 AND is_correct IS NULL`

	rows, err := tx.Query(ctx, answersQuery, cardID)
	if err != nil {
		return fmt.Errorf("bulk resolve: fetch answers: %w", err)
	}
	defer rows.Close()

	type pendingAnswer struct {
		id     uuid.UUID
		userID uuid.UUID
		answer bool
	}

	var pending []pendingAnswer
	for rows.Next() {
		var p pendingAnswer
		if err := rows.Scan(&p.id, &p.userID, &p.answer); err != nil {
			return fmt.Errorf("bulk resolve: scan answer: %w", err)
		}
		pending = append(pending, p)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("bulk resolve: iterate answers: %w", err)
	}
	rows.Close()

	// Resolve each answer
	updateAnswerQuery := `
		UPDATE user_answers
		SET is_correct = $2, points_earned = $3, resolved_at = NOW()
		WHERE id = $1`

	updateUserPointsQuery := `
		UPDATE users
		SET total_points = total_points + $2, updated_at = NOW()
		WHERE id = $1`

	for _, p := range pending {
		isCorrect := p.answer == correctAnswer
		points := 0
		if isCorrect {
			points = card.PointsForAnswer(p.answer)
		}

		if _, err := tx.Exec(ctx, updateAnswerQuery, p.id, isCorrect, points); err != nil {
			return fmt.Errorf("bulk resolve: update answer %s: %w", p.id, err)
		}

		if points > 0 {
			if _, err := tx.Exec(ctx, updateUserPointsQuery, p.userID, points); err != nil {
				return fmt.Errorf("bulk resolve: update user points %s: %w", p.userID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("bulk resolve: commit: %w", err)
	}
	return nil
}
