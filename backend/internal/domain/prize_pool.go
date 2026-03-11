package domain

import (
	"time"

	"github.com/google/uuid"
)

// Prize pool status constants.
const (
	PrizePoolStatusActive    = "active"
	PrizePoolStatusCompleted = "completed"
	PrizePoolStatusCancelled = "cancelled"
)

// PrizePool represents a prize pool for a competition period.
type PrizePool struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	TotalAmount float64    `json:"total_amount"`
	Currency    string     `json:"currency"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PrizePoolDistribution represents a single distribution from a prize pool.
type PrizePoolDistribution struct {
	ID            uuid.UUID `json:"id"`
	PrizePoolID   uuid.UUID `json:"prize_pool_id"`
	UserID        uuid.UUID `json:"user_id"`
	Amount        float64   `json:"amount"`
	Rank          int       `json:"rank"`
	Status        string    `json:"status"`
	DistributedAt time.Time `json:"distributed_at"`
}
