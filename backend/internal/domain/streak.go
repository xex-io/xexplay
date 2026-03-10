package domain

import (
	"time"

	"github.com/google/uuid"
)

// Milestone thresholds in days.
const (
	MilestoneDay3  = 3
	MilestoneDay7  = 7
	MilestoneDay10 = 10
	MilestoneDay14 = 14
	MilestoneDay21 = 21
	MilestoneDay30 = 30
)

// Streak represents a user's consecutive daily play streak.
type Streak struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	CurrentStreak  int        `json:"current_streak"`
	LongestStreak  int        `json:"longest_streak"`
	LastPlayedDate *time.Time `json:"last_played_date,omitempty"`
	BonusSkips     int        `json:"bonus_skips"`
	BonusAnswers   int        `json:"bonus_answers"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// StreakMilestone defines the rewards given when a streak milestone is reached.
type StreakMilestone struct {
	Days         int     `json:"days"`
	BonusSkips   int     `json:"bonus_skips"`
	BonusAnswers int     `json:"bonus_answers"`
	TokenReward  float64 `json:"token_reward"`
	Description  string  `json:"description"`
}

// GetMilestones returns all defined streak milestones.
func GetMilestones() []StreakMilestone {
	return []StreakMilestone{
		{Days: MilestoneDay3, BonusSkips: 0, BonusAnswers: 0, TokenReward: 0, Description: "3-day streak! Keep going!"},
		{Days: MilestoneDay7, BonusSkips: 1, BonusAnswers: 0, TokenReward: 0, Description: "7-day streak! +1 bonus skip"},
		{Days: MilestoneDay10, BonusSkips: 1, BonusAnswers: 0, TokenReward: 1.0, Description: "10-day streak! +1 bonus skip + token reward"},
		{Days: MilestoneDay14, BonusSkips: 0, BonusAnswers: 1, TokenReward: 0, Description: "14-day streak! +1 bonus answer"},
		{Days: MilestoneDay21, BonusSkips: 1, BonusAnswers: 1, TokenReward: 2.0, Description: "21-day streak! +1 bonus skip + +1 bonus answer + token reward"},
		{Days: MilestoneDay30, BonusSkips: 2, BonusAnswers: 1, TokenReward: 5.0, Description: "30-day streak! +2 bonus skips + +1 bonus answer + token reward"},
	}
}

// CheckMilestone returns the milestone if the given day count exactly matches one, or nil.
func CheckMilestone(days int) *StreakMilestone {
	for _, m := range GetMilestones() {
		if m.Days == days {
			return &m
		}
	}
	return nil
}
