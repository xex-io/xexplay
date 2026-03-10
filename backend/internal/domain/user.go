package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	XexUserID    uuid.UUID `json:"xex_user_id"`
	DisplayName  string    `json:"display_name"`
	Email        string    `json:"email,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Role         string    `json:"role"`
	ReferralCode string    `json:"referral_code"`
	ReferredBy   *uuid.UUID `json:"referred_by,omitempty"`
	Language     string    `json:"language"`
	TotalPoints  int       `json:"total_points"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
