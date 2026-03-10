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
	TotalReferrals    int `json:"total_referrals"`
	CompletedReferrals int `json:"completed_referrals"`
}
