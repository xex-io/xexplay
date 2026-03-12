package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AutomationLog struct {
	ID             uuid.UUID       `json:"id"`
	JobName        string          `json:"job_name"`
	Status         string          `json:"status"` // "success", "error"
	Details        json.RawMessage `json:"details,omitempty"`
	ItemsProcessed int             `json:"items_processed"`
	ErrorMessage   string          `json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}
