package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/repository/redis"
)

type HealthHandler struct {
	db  *postgres.DB
	rdb *redis.Client
}

func NewHealthHandler(db *postgres.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

// Liveness returns 200 if the server is running.
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness checks DB and Redis connectivity.
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	checks := gin.H{}

	// Check PostgreSQL
	if err := h.db.Pool.Ping(ctx); err != nil {
		checks["postgres"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"checks": checks,
		})
		return
	}
	checks["postgres"] = "healthy"

	// Check Redis
	if err := h.rdb.Underlying().Ping(ctx).Err(); err != nil {
		checks["redis"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"checks": checks,
		})
		return
	}
	checks["redis"] = "healthy"

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": checks,
	})
}
