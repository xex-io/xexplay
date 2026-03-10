package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	redisclient "github.com/xex-exchange/xexplay-api/internal/repository/redis"
)

// RateLimiter limits requests per user using Redis counters.
func RateLimiter(rdb *redisclient.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get(ContextKeyUserID)
		if !exists {
			c.Next()
			return
		}

		userID, _ := userIDVal.(uuid.UUID)
		key := fmt.Sprintf("rate_limit:%s:%s", userID.String(), c.FullPath())

		ctx := context.Background()
		client := rdb.Underlying()

		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			// If Redis is down, allow the request
			c.Next()
			return
		}

		if count == 1 {
			client.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			response.TooManyRequests(c, "rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}
