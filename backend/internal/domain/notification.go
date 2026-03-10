package domain

import (
	"time"

	"github.com/google/uuid"
)

// DeviceType constants.
const (
	DeviceIOS     = "ios"
	DeviceAndroid = "android"
	DeviceWeb     = "web"
)

// NotificationTarget constants.
const (
	TargetAll     = "all"
	TargetUser    = "user"
	TargetSegment = "segment"
)

// FCMToken represents a device's push notification token.
type FCMToken struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Token      string    `json:"token"`
	DeviceType string    `json:"device_type"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Notification represents a push notification payload.
type Notification struct {
	Title      string            `json:"title"`
	Body       string            `json:"body"`
	Data       map[string]string `json:"data,omitempty"`
	TargetType string            `json:"target_type"`
}
