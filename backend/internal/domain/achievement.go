package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ConditionType constants for achievements.
const (
	ConditionFirstPrediction = "first_prediction"
	ConditionPerfectDay      = "perfect_day"
	ConditionStreak10        = "streak_10"
	ConditionStreak30        = "streak_30"
	ConditionChampion        = "champion"
	ConditionReferrals5      = "referrals_5"
	ConditionReferrals10     = "referrals_10"
	ConditionSessions50      = "sessions_50"
	ConditionSessions100     = "sessions_100"
	ConditionCorrect500      = "correct_500"
)

// Achievement represents an achievement definition.
type Achievement struct {
	ID             uuid.UUID       `json:"id"`
	Key            string          `json:"key"`
	Name           json.RawMessage `json:"name"`
	Description    json.RawMessage `json:"description"`
	Icon           string          `json:"icon"`
	Category       string          `json:"category"`
	ConditionType  string          `json:"condition_type"`
	ConditionValue int             `json:"condition_value"`
	CreatedAt      time.Time       `json:"created_at"`
}

// UserAchievement represents an achievement earned by a user.
type UserAchievement struct {
	ID            uuid.UUID   `json:"id"`
	UserID        uuid.UUID   `json:"user_id"`
	AchievementID uuid.UUID   `json:"achievement_id"`
	EarnedAt      time.Time   `json:"earned_at"`
	Achievement   *Achievement `json:"achievement,omitempty"`
}
