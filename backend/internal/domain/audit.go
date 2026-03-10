package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Audit action constants.
const (
	AuditActionCardResolved       = "card_resolved"
	AuditActionUserBanned         = "user_banned"
	AuditActionUserUpdated        = "user_updated"
	AuditActionRewardDistributed  = "reward_distributed"
	AuditActionRewardConfigChange = "reward_config_change"
	AuditActionNotificationSent   = "notification_sent"
	AuditActionEventCreated       = "event_created"
	AuditActionEventUpdated       = "event_updated"
	AuditActionMatchCreated       = "match_created"
	AuditActionMatchUpdated       = "match_updated"
	AuditActionBasketPublished    = "basket_published"
	AuditActionAbuseFlagReviewed  = "abuse_flag_reviewed"
)

// Audit entity type constants.
const (
	EntityTypeUser         = "user"
	EntityTypeCard         = "card"
	EntityTypeEvent        = "event"
	EntityTypeMatch        = "match"
	EntityTypeBasket       = "basket"
	EntityTypeReward       = "reward"
	EntityTypeNotification = "notification"
	EntityTypeAbuseFlag    = "abuse_flag"
)

// AuditLog represents an admin action audit record.
type AuditLog struct {
	ID          uuid.UUID       `json:"id"`
	AdminUserID uuid.UUID       `json:"admin_user_id"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	Details     json.RawMessage `json:"details"`
	IPAddress   string          `json:"ip_address"`
	CreatedAt   time.Time       `json:"created_at"`
}
