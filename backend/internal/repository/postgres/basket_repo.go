package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type BasketRepo struct {
	db *DB
}

func NewBasketRepo(db *DB) *BasketRepo {
	return &BasketRepo{db: db}
}

func (r *BasketRepo) Create(ctx context.Context, b *domain.DailyBasket) error {
	query := `
		INSERT INTO daily_baskets (id, basket_date, event_id, is_published, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at`

	err := r.db.Pool.QueryRow(ctx, query,
		b.ID, b.BasketDate, b.EventID, b.IsPublished,
	).Scan(&b.CreatedAt)
	if err != nil {
		return fmt.Errorf("create basket: %w", err)
	}
	return nil
}

func (r *BasketRepo) AddCards(ctx context.Context, basketID uuid.UUID, cardIDs []uuid.UUID) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx for add cards: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, cardID := range cardIDs {
		query := `
			INSERT INTO daily_basket_cards (id, basket_id, card_id, position)
			VALUES ($1, $2, $3, $4)`

		_, err := tx.Exec(ctx, query, uuid.New(), basketID, cardID, i+1)
		if err != nil {
			return fmt.Errorf("add card %d to basket: %w", i+1, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit add cards: %w", err)
	}
	return nil
}

func (r *BasketRepo) Publish(ctx context.Context, basketID uuid.UUID) error {
	// Validate tier composition: 3 Gold + 5 Silver + 7 White = 15 total
	countQuery := `
		SELECT
			COUNT(*) FILTER (WHERE c.tier = $2),
			COUNT(*) FILTER (WHERE c.tier = $3),
			COUNT(*) FILTER (WHERE c.tier = $4)
		FROM daily_basket_cards bc
		JOIN cards c ON c.id = bc.card_id
		WHERE bc.basket_id = $1`

	var goldCount, silverCount, whiteCount int
	err := r.db.Pool.QueryRow(ctx, countQuery,
		basketID, domain.TierGold, domain.TierSilver, domain.TierWhite,
	).Scan(&goldCount, &silverCount, &whiteCount)
	if err != nil {
		return fmt.Errorf("count basket cards by tier: %w", err)
	}

	if goldCount != domain.GoldCount {
		return fmt.Errorf("publish basket: expected %d gold cards, got %d", domain.GoldCount, goldCount)
	}
	if silverCount != domain.SilverCount {
		return fmt.Errorf("publish basket: expected %d silver cards, got %d", domain.SilverCount, silverCount)
	}
	if whiteCount != domain.WhiteCount {
		return fmt.Errorf("publish basket: expected %d white cards, got %d", domain.WhiteCount, whiteCount)
	}

	publishQuery := `
		UPDATE daily_baskets
		SET is_published = true
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, publishQuery, basketID)
	if err != nil {
		return fmt.Errorf("publish basket: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("publish basket: basket not found")
	}
	return nil
}

func (r *BasketRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.DailyBasket, error) {
	query := `
		SELECT id, basket_date, event_id, is_published, created_at
		FROM daily_baskets
		WHERE id = $1`

	var b domain.DailyBasket
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.BasketDate, &b.EventID, &b.IsPublished, &b.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find basket by id: %w", err)
	}
	return &b, nil
}

func (r *BasketRepo) FindAll(ctx context.Context) ([]*domain.DailyBasket, error) {
	query := `
		SELECT id, basket_date, event_id, is_published, created_at
		FROM daily_baskets
		ORDER BY basket_date DESC, created_at DESC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all baskets: %w", err)
	}
	defer rows.Close()

	var baskets []*domain.DailyBasket
	for rows.Next() {
		var b domain.DailyBasket
		if err := rows.Scan(
			&b.ID, &b.BasketDate, &b.EventID, &b.IsPublished, &b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan basket: %w", err)
		}
		baskets = append(baskets, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate baskets: %w", err)
	}
	return baskets, nil
}

func (r *BasketRepo) Update(ctx context.Context, b *domain.DailyBasket) error {
	query := `
		UPDATE daily_baskets
		SET basket_date = $2, event_id = $3, is_published = $4
		WHERE id = $1`

	ct, err := r.db.Pool.Exec(ctx, query, b.ID, b.BasketDate, b.EventID, b.IsPublished)
	if err != nil {
		return fmt.Errorf("update basket: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("update basket: basket not found")
	}
	return nil
}

func (r *BasketRepo) RemoveAllCards(ctx context.Context, basketID uuid.UUID) error {
	query := `DELETE FROM daily_basket_cards WHERE basket_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, basketID)
	if err != nil {
		return fmt.Errorf("remove all cards from basket: %w", err)
	}
	return nil
}

func (r *BasketRepo) FindByDateAndEvent(ctx context.Context, date time.Time, eventID uuid.UUID) (*domain.DailyBasket, error) {
	query := `
		SELECT id, basket_date, event_id, is_published, created_at
		FROM daily_baskets
		WHERE basket_date::date = $1::date AND event_id = $2`

	var b domain.DailyBasket
	err := r.db.Pool.QueryRow(ctx, query, date, eventID).Scan(
		&b.ID, &b.BasketDate, &b.EventID, &b.IsPublished, &b.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find basket by date and event: %w", err)
	}
	return &b, nil
}

func (r *BasketRepo) FindPublishedByDate(ctx context.Context, date time.Time) ([]*domain.DailyBasket, error) {
	query := `
		SELECT id, basket_date, event_id, is_published, created_at
		FROM daily_baskets
		WHERE basket_date::date = $1::date AND is_published = true
		ORDER BY created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("find published baskets by date: %w", err)
	}
	defer rows.Close()

	var baskets []*domain.DailyBasket
	for rows.Next() {
		var b domain.DailyBasket
		if err := rows.Scan(
			&b.ID, &b.BasketDate, &b.EventID, &b.IsPublished, &b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan basket: %w", err)
		}
		baskets = append(baskets, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate baskets: %w", err)
	}
	return baskets, nil
}

func (r *BasketRepo) GetCardsForBasket(ctx context.Context, basketID uuid.UUID) ([]*domain.Card, error) {
	query := `
		SELECT c.id, c.match_id, c.question_text, c.tier, c.high_answer_is_yes,
		       c.correct_answer, c.is_resolved, c.available_date, c.expires_at,
		       c.created_at, c.updated_at
		FROM cards c
		JOIN daily_basket_cards bc ON bc.card_id = c.id
		WHERE bc.basket_id = $1
		ORDER BY bc.position ASC`

	rows, err := r.db.Pool.Query(ctx, query, basketID)
	if err != nil {
		return nil, fmt.Errorf("get cards for basket: %w", err)
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
			return nil, fmt.Errorf("scan basket card: %w", err)
		}
		cards = append(cards, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate basket cards: %w", err)
	}
	return cards, nil
}
