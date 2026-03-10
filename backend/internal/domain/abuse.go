package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Abuse flag type constants.
const (
	AbuseFlagMultiAccount = "multi_account"
	AbuseFlagPerfectScore = "perfect_score"
	AbuseFlagVelocity     = "velocity"
	AbuseFlagPattern      = "pattern"
)

// Abuse flag status constants.
const (
	AbuseFlagStatusPending   = "pending"
	AbuseFlagStatusReviewed  = "reviewed"
	AbuseFlagStatusDismissed = "dismissed"
)

// AbuseFlag represents a flagged suspicious activity for admin review.
type AbuseFlag struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	FlagType   string          `json:"flag_type"`
	Details    json.RawMessage `json:"details"`
	Status     string          `json:"status"`
	ReviewedBy *uuid.UUID      `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
