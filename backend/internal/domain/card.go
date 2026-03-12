package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	TierGold   = "gold"
	TierSilver = "silver"
	TierWhite  = "white"
	TierVIP    = "vip" // Exclusive tier for active Exchange traders

	// Fixed scoring per tier
	GoldHighPoints   = 20
	GoldLowPoints    = 5
	SilverHighPoints = 15
	SilverLowPoints  = 10
	WhitePoints      = 10
	VIPHighPoints    = 30 // VIP cards award more points
	VIPLowPoints     = 10

	// Fixed tier counts per basket
	GoldCount   = 3
	SilverCount = 5
	WhiteCount  = 7
	TotalCards  = 15

	// Resource limits per session
	MaxAnswers = 10
	MaxSkips   = 5

	// Timer
	CardTimerSeconds      = 40
	CardTimerGraceSeconds = 2 // Grace period for network latency
)

type Card struct {
	ID                 uuid.UUID       `json:"id"`
	MatchID            uuid.UUID       `json:"match_id"`
	QuestionText       json.RawMessage `json:"question_text"`     // JSONB: {"en": "...", "fa": "..."}
	Tier               string          `json:"tier"`               // gold, silver, white
	HighAnswerIsYes    *bool           `json:"high_answer_is_yes"` // NULL for white
	CorrectAnswer      *bool           `json:"correct_answer"`     // NULL until resolved
	IsResolved         bool            `json:"is_resolved"`
	AvailableDate      time.Time       `json:"available_date"`
	ExpiresAt          time.Time       `json:"expires_at"`
	Source             string          `json:"source"`
	AIPromptData       json.RawMessage `json:"ai_prompt_data,omitempty"`
	ResolutionCriteria string          `json:"resolution_criteria,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// PointsForAnswer returns the points a user would earn for a given answer on this card, if correct.
func (c *Card) PointsForAnswer(answer bool) int {
	switch c.Tier {
	case TierVIP:
		if c.HighAnswerIsYes == nil {
			return VIPLowPoints
		}
		if answer == *c.HighAnswerIsYes {
			return VIPHighPoints
		}
		return VIPLowPoints
	case TierGold:
		if c.HighAnswerIsYes == nil {
			return GoldLowPoints
		}
		if answer == *c.HighAnswerIsYes {
			return GoldHighPoints
		}
		return GoldLowPoints
	case TierSilver:
		if c.HighAnswerIsYes == nil {
			return SilverLowPoints
		}
		if answer == *c.HighAnswerIsYes {
			return SilverHighPoints
		}
		return SilverLowPoints
	default: // white
		return WhitePoints
	}
}

// CardView is the client-facing card representation (hides correct_answer).
type CardView struct {
	ID           uuid.UUID       `json:"id"`
	MatchID      uuid.UUID       `json:"match_id"`
	QuestionText json.RawMessage `json:"question_text"`
	Tier         string          `json:"tier"`
	YesPoints    int             `json:"yes_points"`
	NoPoints     int             `json:"no_points"`
	ExpiresAt    time.Time       `json:"expires_at"`
}

// ToView converts a Card to a CardView (safe for client).
func (c *Card) ToView() *CardView {
	yes := true
	no := false
	return &CardView{
		ID:           c.ID,
		MatchID:      c.MatchID,
		QuestionText: c.QuestionText,
		Tier:         c.Tier,
		YesPoints:    c.PointsForAnswer(yes),
		NoPoints:     c.PointsForAnswer(no),
		ExpiresAt:    c.ExpiresAt,
	}
}
