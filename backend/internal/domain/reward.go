package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RewardType constants.
const (
	RewardToken       = "token"
	RewardBonusSkip   = "bonus_skip"
	RewardBonusAnswer = "bonus_answer"
	RewardBadge       = "badge"
)

// Reward distribution status constants.
const (
	StatusPending  = "pending"
	StatusClaimed  = "claimed"
	StatusCredited = "credited"
	StatusExpired  = "expired"
)

// RewardConfig defines a reward rule for a specific period type and rank range.
type RewardConfig struct {
	ID          uuid.UUID       `json:"id"`
	PeriodType  string          `json:"period_type"`
	RankFrom    int             `json:"rank_from"`
	RankTo      int             `json:"rank_to"`
	RewardType  string          `json:"reward_type"`
	Amount      float64         `json:"amount"`
	Description json.RawMessage `json:"description,omitempty"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// RewardDistribution represents a reward assigned to a user for a specific period.
type RewardDistribution struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	RewardConfigID *uuid.UUID `json:"reward_config_id,omitempty"`
	PeriodType     string     `json:"period_type"`
	PeriodKey      string     `json:"period_key"`
	RewardType     string     `json:"reward_type"`
	Amount         float64    `json:"amount"`
	Rank           int        `json:"rank"`
	Status         string     `json:"status"`
	ClaimedAt      *time.Time `json:"claimed_at,omitempty"`
	CreditedAt     *time.Time `json:"credited_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// RewardLeaderboardEntry is a lightweight struct used when distributing rewards.
type RewardLeaderboardEntry struct {
	UserID      uuid.UUID `json:"user_id"`
	Rank        int       `json:"rank"`
	TotalPoints int       `json:"total_points"`
}
