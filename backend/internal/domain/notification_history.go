package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationHistory represents a record of a sent notification.
type NotificationHistory struct {
	ID             uuid.UUID `json:"id"`
	AdminUserID    uuid.UUID `json:"admin_user_id"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	TargetType     string    `json:"target_type"`
	RecipientCount int       `json:"recipient_count"`
	CreatedAt      time.Time `json:"created_at"`
}
