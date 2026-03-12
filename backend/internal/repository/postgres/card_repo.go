package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

const cardColumns = `id, match_id, question_text, tier, high_answer_is_yes,
	correct_answer, is_resolved, available_date, expires_at,
	source, ai_prompt_data, resolution_criteria,
	created_at, updated_at`

type CardRepo struct {
	db *DB
}

func NewCardRepo(db *DB) *CardRepo {
	return &CardRepo{db: db}
}

func scanCard(scan func(dest ...interface{}) error) (*domain.Card, error) {
	var c domain.Card
	err := scan(
		&c.ID, &c.MatchID, &c.QuestionText, &c.Tier, &c.HighAnswerIsYes,
		&c.CorrectAnswer, &c.IsResolved, &c.AvailableDate, &c.ExpiresAt,
		&c.Source, &c.AIPromptData, &c.ResolutionCriteria,
		&c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

func (r *CardRepo) scanCardRows(ctx context.Context, query string, args ...interface{}) ([]*domain.Card, error) {
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		c, err := scanCard(rows.Scan)
		if err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (r *CardRepo) Create(ctx context.Context, c *domain.Card) error {
	if c.Source == "" {
		c.Source = "manual"
	}
	query := `
		INSERT INTO cards (id, match_id, question_text, tier, high_answer_is_yes,
		                    correct_answer, is_resolved, available_date, expires_at,
		                    source, ai_prompt_data, resolution_criteria,
		                    created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		c.ID, c.MatchID, c.QuestionText, c.Tier, c.HighAnswerIsYes,
		c.CorrectAnswer, c.IsResolved, c.AvailableDate, c.ExpiresAt,
		c.Source, c.AIPromptData, nilIfEmpty(c.ResolutionCriteria),
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
		    source = $10, ai_prompt_data = $11, resolution_criteria = $12,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		c.ID, c.MatchID, c.QuestionText, c.Tier, c.HighAnswerIsYes,
		c.CorrectAnswer, c.IsResolved, c.AvailableDate, c.ExpiresAt,
		c.Source, c.AIPromptData, nilIfEmpty(c.ResolutionCriteria),
	).Scan(&c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update card: %w", err)
	}
	return nil
}

func (r *CardRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Card, error) {
	query := `SELECT ` + cardColumns + ` FROM cards WHERE id = $1`

	c, err := scanCard(r.db.Pool.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find card by id: %w", err)
	}
	return c, nil
}

func (r *CardRepo) FindByAvailableDate(ctx context.Context, date time.Time) ([]*domain.Card, error) {
	query := `SELECT ` + cardColumns + `
		FROM cards WHERE available_date::date = $1::date
		ORDER BY tier, created_at ASC`
	cards, err := r.scanCardRows(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("find cards by available date: %w", err)
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

func (r *CardRepo) CountAnswersByCardID(ctx context.Context, cardID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM answers WHERE card_id = $1`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, cardID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count answers by card: %w", err)
	}
	return count, nil
}

func (r *CardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM cards WHERE id = $1`
	ct, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("delete card: card not found")
	}
	return nil
}

func (r *CardRepo) FindAll(ctx context.Context) ([]*domain.Card, error) {
	query := `SELECT ` + cardColumns + `
		FROM cards ORDER BY available_date DESC, tier, created_at ASC`
	cards, err := r.scanCardRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all cards: %w", err)
	}
	return cards, nil
}

func (r *CardRepo) FindUnresolvedByMatch(ctx context.Context, matchID uuid.UUID) ([]*domain.Card, error) {
	query := `SELECT ` + cardColumns + `
		FROM cards WHERE match_id = $1 AND is_resolved = false
		ORDER BY created_at ASC`
	cards, err := r.scanCardRows(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("find unresolved cards by match: %w", err)
	}
	return cards, nil
}

// CountByMatchID returns the number of cards for a match.
func (r *CardRepo) CountByMatchID(ctx context.Context, matchID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM cards WHERE match_id = $1`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, matchID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count cards by match: %w", err)
	}
	return count, nil
}
