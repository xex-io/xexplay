package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/xex-exchange/xexplay-api/internal/pkg/jwt"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

const (
	ContextKeyUserID    = "user_id"
	ContextKeyXexUserID = "xex_user_id"
	ContextKeyEmail     = "email"
	ContextKeyRole      = "role"
)

// Auth validates the Exchange-issued JWT using the shared secret.
// It resolves the Exchange user ID (xex_user_id) to the internal Play user ID
// so that all downstream handlers can use FindByID with the internal ID.
func Auth(jwtSecret string, userRepo *postgres.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtpkg.Parse(parts[1], jwtSecret)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// Resolve xex_user_id to internal Play user ID
		if userRepo != nil {
			user, err := userRepo.FindByXexUserID(c.Request.Context(), claims.UserID)
			if err != nil {
				response.InternalError(c, "failed to resolve user")
				c.Abort()
				return
			}
			if user == nil {
				response.Unauthorized(c, "user not registered")
				c.Abort()
				return
			}
			c.Set(ContextKeyUserID, user.ID)
		} else {
			c.Set(ContextKeyUserID, claims.UserID)
		}
		c.Set(ContextKeyXexUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}
