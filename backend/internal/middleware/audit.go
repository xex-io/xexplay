package middleware

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/service"
)

// AuditLog is a middleware that auto-logs admin actions to the audit trail.
// It captures the request method, path, and the admin user who performed the action.
func AuditLog(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only audit mutating requests
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
			c.Next()
			return
		}

		// Read the body for audit details (re-buffer so handlers can read it)
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Process the request first
		c.Next()

		// Only log if the request was successful (2xx status)
		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		adminUserID, exists := c.Get(ContextKeyUserID)
		if !exists {
			return
		}

		uid, ok := adminUserID.(uuid.UUID)
		if !ok {
			return
		}

		action := method + " " + c.FullPath()
		entityType := resolveEntityType(c.FullPath())
		entityID := c.Param("id")

		details := map[string]interface{}{
			"method":      method,
			"path":        c.Request.URL.Path,
			"status_code": status,
		}

		ipAddress := c.ClientIP()

		auditService.LogAction(c.Request.Context(), uid, action, entityType, entityID, details, ipAddress)
	}
}

// resolveEntityType extracts a human-readable entity type from the route path.
func resolveEntityType(path string) string {
	switch {
	case contains(path, "/cards"):
		return "card"
	case contains(path, "/events"):
		return "event"
	case contains(path, "/matches"):
		return "match"
	case contains(path, "/baskets"):
		return "basket"
	case contains(path, "/users"):
		return "user"
	case contains(path, "/rewards"):
		return "reward"
	case contains(path, "/notifications"):
		return "notification"
	case contains(path, "/abuse-flags"):
		return "abuse_flag"
	default:
		return "unknown"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
