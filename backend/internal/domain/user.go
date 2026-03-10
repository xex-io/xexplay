package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Trading tier constants for Exchange-linked users.
const (
	TradingTierNone   = ""       // Not an active trader
	TradingTierBasic  = "basic"  // Low-volume trader
	TradingTierSilver = "silver" // Mid-volume trader
	TradingTierGold   = "gold"   // High-volume trader
	TradingTierVIP    = "vip"    // Top-tier trader
)

// Exchange account status constants.
const (
	ExchangeStatusUnlinked = ""         // No Exchange account linked
	ExchangeStatusActive   = "active"   // Account in good standing
	ExchangeStatusSuspended = "suspended" // Account suspended
	ExchangeStatusRestricted = "restricted" // Account restricted
)

type User struct {
	ID             uuid.UUID  `json:"id"`
	XexUserID      uuid.UUID  `json:"xex_user_id"`
	DisplayName    string     `json:"display_name"`
	Email          string     `json:"email,omitempty"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	Role           string     `json:"role"`
	ReferralCode   string     `json:"referral_code"`
	ReferredBy     *uuid.UUID `json:"referred_by,omitempty"`
	Language       string     `json:"language"`
	TotalPoints    int        `json:"total_points"`
	IsActive       bool       `json:"is_active"`
	TradingTier    string     `json:"trading_tier,omitempty"`
	ExchangeStatus string     `json:"exchange_status,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// IsActiveTrader returns true if the user has any active trading tier.
func (u *User) IsActiveTrader() bool {
	return u.TradingTier != "" && u.TradingTier != TradingTierNone
}

// IsExchangeAccountGoodStanding returns true if the user's Exchange account is active
// and their Play account is at least 7 days old.
func (u *User) IsExchangeAccountGoodStanding() bool {
	if u.ExchangeStatus != ExchangeStatusActive {
		return false
	}
	if time.Since(u.CreatedAt) < 7*24*time.Hour {
		return false
	}
	return true
}

type UserStats struct {
	TotalPoints     int `json:"total_points"`
	TotalSessions   int `json:"total_sessions"`
	TotalAnswers    int `json:"total_answers"`
	CorrectAnswers  int `json:"correct_answers"`
	CurrentStreak   int `json:"current_streak"`
	LongestStreak   int `json:"longest_streak"`
	DailyRank       int `json:"daily_rank,omitempty"`
	WeeklyRank      int `json:"weekly_rank,omitempty"`
}
