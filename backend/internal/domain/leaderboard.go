package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	PeriodDaily      = "daily"
	PeriodWeekly     = "weekly"
	PeriodTournament = "tournament"
	PeriodAllTime    = "all_time"
)

// LeaderboardEntry is the persistent leaderboard record stored in PostgreSQL.
type LeaderboardEntry struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	EventID        *uuid.UUID `json:"event_id,omitempty"`
	PeriodType     string     `json:"period_type"`
	PeriodKey      string     `json:"period_key"`
	TotalPoints    int        `json:"total_points"`
	CorrectAnswers int        `json:"correct_answers"`
	WrongAnswers   int        `json:"wrong_answers"`
	TotalAnswers   int        `json:"total_answers"`
	Rank           int        `json:"rank"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// LeaderboardRow is the view struct returned in API responses.
type LeaderboardRow struct {
	Rank           int       `json:"rank"`
	UserID         uuid.UUID `json:"user_id"`
	DisplayName    string    `json:"display_name"`
	AvatarURL      string    `json:"avatar_url"`
	TotalPoints    int       `json:"total_points"`
	CorrectAnswers int       `json:"correct_answers"`
}

// LeaderboardResponse is the full API response for a leaderboard query.
type LeaderboardResponse struct {
	PeriodType string           `json:"period_type"`
	PeriodKey  string           `json:"period_key"`
	Entries    []LeaderboardRow `json:"entries"`
	UserRank   *LeaderboardRow  `json:"user_rank,omitempty"`
	Total      int              `json:"total"`
}
