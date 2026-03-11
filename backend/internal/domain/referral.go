package domain

import (
	"time"

	"github.com/google/uuid"
)

// Referral status constants.
const (
	ReferralStatusSignedUp     = "signed_up"
	ReferralStatusFirstSession = "first_session"
)

// Referral represents a referral relationship between two users.
type Referral struct {
	ID            uuid.UUID `json:"id"`
	ReferrerID    uuid.UUID `json:"referrer_id"`
	ReferredID    uuid.UUID `json:"referred_id"`
	Status        string    `json:"status"`
	RewardGranted bool      `json:"reward_granted"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ReferralStats holds aggregated referral statistics for a user.
type ReferralStats struct {
	TotalReferrals     int `json:"total_referrals"`
	CompletedReferrals int `json:"completed_referrals"`
}

// TopReferrer represents a top referrer entry for admin reporting.
type TopReferrer struct {
	UserID         uuid.UUID `json:"user_id"`
	DisplayName    string    `json:"display_name"`
	Email          string    `json:"email"`
	ReferralCount  int       `json:"referral_count"`
	ConvertedCount int       `json:"converted_count"`
}
