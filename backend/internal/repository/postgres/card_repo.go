package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type CardRepo struct {
	db *DB
}

func NewCardRepo(db *DB) *CardRepo {
	return &CardRepo{db: db}
}

func (r *CardRepo) Create(ctx context.Context, c *domain.Card) error {
	query := `
		INSERT INTO cards (id, match_id, question_text, tier, high_answer_is_yes,
		                    correct_answer, is_resolved, available_date, expires_at,
		                    created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		c.ID, c.MatchID, c.QuestionText, c.Tier, c.HighAnswerIsYes,
		c.CorrectAnswer, c.IsResolved, c.AvailableDate, c.ExpiresAt,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create card: %w", err)
	}
	return nil
}

func (r *CardRepo) Update(ctx context.Context, c *domain.Card) error {
	query := `
		UPDATE cards
		SET match_id = $2, question_text = $3, tier = $4, high_answer_is_yes = $5,
		    correct_answer = $6, is_resolved = $7, available_date = $8, expires_at = $9,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		c.ID, c.MatchID, c.QuestionText, c.Tier, c.HighAnswerIsYes,
		c.CorrectAnswer, c.IsResolved, c.AvailableDate, c.ExpiresAt,
	).Scan(&c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update card: %w", err)
	}
	return nil
}

func (r *CardRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Card, error) {
	query := `
		SELECT id, match_id, question_text, tier, high_answer_is_yes,
		       correct_answer, is_resolved, available_date, expires_at,
		       created_at, updated_at
		FROM cards
		WHERE id = $1`

	var c domain.Card
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.MatchID, &c.QuestionText, &c.Tier, &c.HighAnswerIsYes,
		&c.CorrectAnswer, &c.IsResolved, &c.AvailableDate, &c.ExpiresAt,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find card by id: %w", err)
	}
	return &c, nil
}

func (r *CardRepo) FindByAvailableDate(ctx context.Context, date time.Time) ([]*domain.Card, error) {
	query := `
		SELECT id, match_id, question_text, tier, high_answer_is_yes,
		       correct_answer, is_resolved, available_date, expires_at,
		       created_at, updated_at
		FROM cards
		WHERE available_date::date = $1::date
		ORDER BY tier, created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("find cards by available date: %w", err)
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		var c domain.Card
		if err := rows.Scan(
			&c.ID, &c.MatchID, &c.QuestionText, &c.Tier, &c.HighAnswerIsYes,
			&c.CorrectAnswer, &c.IsResolved, &c.AvailableDate, &c.ExpiresAt,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		cards = append(cards, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cards: %w", err)
	}
	return cards, nil
}

func (r *CardRepo) Resolve(ctx context.Context, cardID uuid.UUID, correctAnswer bool) error {
	query := `
		UPDATE cards
		SET correct_answer = $2, is_resolved = true, updated_at = NOW()
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, cardID, correctAnswer)
	if err != nil {
		return fmt.Errorf("resolve card: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("resolve card: card not found")
	}
	return nil
}

func (r *CardRepo) FindAll(ctx context.Context) ([]*domain.Card, error) {
	query := `
		SELECT id, match_id, question_text, tier, high_answer_is_yes,
		       correct_answer, is_resolved, available_date, expires_at,
		       created_at, updated_at
		FROM cards
		ORDER BY available_date DESC, tier, created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all cards: %w", err)
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		var c domain.Card
		if err := rows.Scan(
			&c.ID, &c.MatchID, &c.QuestionText, &c.Tier, &c.HighAnswerIsYes,
			&c.CorrectAnswer, &c.IsResolved, &c.AvailableDate, &c.ExpiresAt,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		cards = append(cards, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cards: %w", err)
	}
	return cards, nil
}

func (r *CardRepo) FindUnresolvedByMatch(ctx context.Context, matchID uuid.UUID) ([]*domain.Card, error) {
	query := `
		SELECT id, match_id, question_text, tier, high_answer_is_yes,
		       correct_answer, is_resolved, available_date, expires_at,
		       created_at, updated_at
		FROM cards
		WHERE match_id = $1 AND is_resolved = false
		ORDER BY created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("find unresolved cards by match: %w", err)
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		var c domain.Card
		if err := rows.Scan(
			&c.ID, &c.MatchID, &c.QuestionText, &c.Tier, &c.HighAnswerIsYes,
			&c.CorrectAnswer, &c.IsResolved, &c.AvailableDate, &c.ExpiresAt,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan unresolved card: %w", err)
		}
		cards = append(cards, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unresolved cards: %w", err)
	}
	return cards, nil
}
