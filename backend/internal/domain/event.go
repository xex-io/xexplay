package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID                uuid.UUID       `json:"id"`
	Name              json.RawMessage `json:"name"`                // JSONB: {"en": "...", "fa": "..."}
	Slug              string          `json:"slug"`
	Description       json.RawMessage `json:"description,omitempty"` // JSONB
	StartDate         time.Time       `json:"start_date"`
	EndDate           time.Time       `json:"end_date"`
	IsActive          bool            `json:"is_active"`
	ScoringMultiplier float64         `json:"scoring_multiplier"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}
