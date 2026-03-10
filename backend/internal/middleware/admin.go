package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
)

// Admin checks that the authenticated user has admin or super_admin role.
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			response.Forbidden(c, "access denied")
			c.Abort()
			return
		}

		roleStr, _ := role.(string)
		if roleStr != domain.RoleAdmin && roleStr != "super_admin" {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
