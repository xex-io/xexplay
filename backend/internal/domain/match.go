package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	MatchStatusUpcoming  = "upcoming"
	MatchStatusLive      = "live"
	MatchStatusCompleted = "completed"
	MatchStatusCancelled = "cancelled"
)

type Match struct {
	ID          uuid.UUID       `json:"id"`
	EventID     uuid.UUID       `json:"event_id"`
	HomeTeam    string          `json:"home_team"`
	AwayTeam    string          `json:"away_team"`
	KickoffTime time.Time       `json:"kickoff_time"`
	Status      string          `json:"status"`
	HomeScore   *int            `json:"home_score,omitempty"`
	AwayScore   *int            `json:"away_score,omitempty"`
	ResultData  json.RawMessage `json:"result_data,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
