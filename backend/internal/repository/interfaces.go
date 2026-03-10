package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

// UserRepository defines the contract for user persistence operations.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByXexUserID(ctx context.Context, xexUserID uuid.UUID) (*domain.User, error)
	FindByReferralCode(ctx context.Context, code string) (*domain.User, error)
	FindByDeviceIDOrIP(ctx context.Context, userID uuid.UUID, deviceID, ip string) ([]domain.User, error)
	Upsert(ctx context.Context, u *domain.User) error
	UpdateProfile(ctx context.Context, id uuid.UUID, displayName, avatarURL, language string) error
	UpdateDeviceInfo(ctx context.Context, id uuid.UUID, deviceID, lastIP string, lastLoginAt time.Time) error
	UpdateTradingTier(ctx context.Context, id uuid.UUID, tradingTier string) error
	UpdateExchangeStatus(ctx context.Context, id uuid.UUID, exchangeStatus string) error
	GetStats(ctx context.Context, userID uuid.UUID) (*domain.UserStats, error)
}

// EventRepository defines the contract for event persistence operations.
type EventRepository interface {
	Create(ctx context.Context, e *domain.Event) error
	Update(ctx context.Context, e *domain.Event) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	FindActive(ctx context.Context) ([]*domain.Event, error)
	FindAll(ctx context.Context) ([]*domain.Event, error)
}

// MatchRepository defines the contract for match persistence operations.
type MatchRepository interface {
	Create(ctx context.Context, m *domain.Match) error
	Update(ctx context.Context, m *domain.Match) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Match, error)
	FindByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain.Match, error)
	UpdateResult(ctx context.Context, id uuid.UUID, homeScore, awayScore int) error
}

// CardRepository defines the contract for card persistence operations.
type CardRepository interface {
	Create(ctx context.Context, c *domain.Card) error
	Update(ctx context.Context, c *domain.Card) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Card, error)
	FindByAvailableDate(ctx context.Context, date time.Time) ([]*domain.Card, error)
	FindUnresolvedByMatch(ctx context.Context, matchID uuid.UUID) ([]*domain.Card, error)
	Resolve(ctx context.Context, cardID uuid.UUID, correctAnswer bool) error
}

// BasketRepository defines the contract for daily basket persistence operations.
type BasketRepository interface {
	Create(ctx context.Context, b *domain.DailyBasket) error
	AddCards(ctx context.Context, basketID uuid.UUID, cardIDs []uuid.UUID) error
	Publish(ctx context.Context, basketID uuid.UUID) error
	FindByDateAndEvent(ctx context.Context, date time.Time, eventID uuid.UUID) (*domain.DailyBasket, error)
	FindPublishedByDate(ctx context.Context, date time.Time) ([]*domain.DailyBasket, error)
	GetCardsForBasket(ctx context.Context, basketID uuid.UUID) ([]*domain.Card, error)
}

// SessionRepository defines the contract for user session persistence operations.
type SessionRepository interface {
	Create(ctx context.Context, s *domain.UserSession) error
	FindByUserAndBasket(ctx context.Context, userID, basketID uuid.UUID) (*domain.UserSession, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.UserSession, error)
	UpdateProgress(ctx context.Context, s *domain.UserSession) error
	Complete(ctx context.Context, sessionID uuid.UUID) error
}

// AnswerRepository defines the contract for user answer persistence operations.
type AnswerRepository interface {
	Create(ctx context.Context, a *domain.UserAnswer) error
	FindBySession(ctx context.Context, sessionID uuid.UUID) ([]*domain.UserAnswer, error)
	FindByCard(ctx context.Context, cardID uuid.UUID) ([]*domain.UserAnswer, error)
	BulkResolve(ctx context.Context, cardID uuid.UUID, correctAnswer bool) error
}
