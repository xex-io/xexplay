package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserAnswer struct {
	ID           uuid.UUID  `json:"id"`
	SessionID    uuid.UUID  `json:"session_id"`
	CardID       uuid.UUID  `json:"card_id"`
	UserID       uuid.UUID  `json:"user_id"`
	Answer       bool       `json:"answer"`        // true=Yes, false=No
	PointsEarned int        `json:"points_earned"` // Set after resolution
	IsCorrect    *bool      `json:"is_correct"`    // NULL until resolved
	AnsweredAt   time.Time  `json:"answered_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

type AnswerResult struct {
	CardID          uuid.UUID    `json:"card_id"`
	Answer          bool         `json:"answer"`
	AutoSkipped     bool         `json:"auto_skipped"`
	SessionProgress *SessionView `json:"session_progress"`
}

type SkipResult struct {
	AutoSkipped     bool         `json:"auto_skipped"`
	SessionProgress *SessionView `json:"session_progress"`
}
