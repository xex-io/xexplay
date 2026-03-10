package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	SessionStatusActive    = "active"
	SessionStatusCompleted = "completed"
	SessionStatusExpired   = "expired"
)

type UserSession struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	BasketID        uuid.UUID  `json:"basket_id"`
	ShuffleOrder    []int      `json:"shuffle_order"`     // Array of card positions in shuffled order
	CurrentIndex    int        `json:"current_index"`     // Which card the user is currently on
	AnswersUsed     int        `json:"answers_used"`
	SkipsUsed       int        `json:"skips_used"`
	BonusAnswers    int        `json:"bonus_answers"`
	BonusSkips      int        `json:"bonus_skips"`
	Status          string     `json:"status"`
	CardPresentedAt *time.Time `json:"card_presented_at,omitempty"` // When the current card was shown
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// AnswersRemaining returns how many answers the user can still make.
func (s *UserSession) AnswersRemaining() int {
	return MaxAnswers + s.BonusAnswers - s.AnswersUsed
}

// SkipsRemaining returns how many skips the user can still make.
func (s *UserSession) SkipsRemaining() int {
	return MaxSkips + s.BonusSkips - s.SkipsUsed
}

// CardsRemaining returns how many cards are left in the session.
func (s *UserSession) CardsRemaining() int {
	return TotalCards - s.CurrentIndex
}

// IsComplete returns true if all cards have been processed.
func (s *UserSession) IsComplete() bool {
	return s.CurrentIndex >= TotalCards
}

// IsCardExpired returns true if the current card's timer (with grace period) has elapsed.
func (s *UserSession) IsCardExpired() bool {
	if s.CardPresentedAt == nil {
		return false
	}
	deadline := time.Duration(CardTimerSeconds+CardTimerGraceSeconds) * time.Second
	return time.Since(*s.CardPresentedAt) > deadline
}

// SessionView is the client-facing session representation.
type SessionView struct {
	ID               uuid.UUID  `json:"id"`
	Status           string     `json:"status"`
	CurrentIndex     int        `json:"current_index"`
	TotalCards       int        `json:"total_cards"`
	AnswersUsed      int        `json:"answers_used"`
	AnswersRemaining int        `json:"answers_remaining"`
	SkipsUsed        int        `json:"skips_used"`
	SkipsRemaining   int        `json:"skips_remaining"`
	CardPresentedAt  *time.Time `json:"card_presented_at,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

func (s *UserSession) ToView() *SessionView {
	return &SessionView{
		ID:               s.ID,
		Status:           s.Status,
		CurrentIndex:     s.CurrentIndex,
		TotalCards:       TotalCards,
		AnswersUsed:      s.AnswersUsed,
		AnswersRemaining: s.AnswersRemaining(),
		SkipsUsed:        s.SkipsUsed,
		SkipsRemaining:   s.SkipsRemaining(),
		CardPresentedAt:  s.CardPresentedAt,
		StartedAt:        s.StartedAt,
		CompletedAt:      s.CompletedAt,
	}
}
