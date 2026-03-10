package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestDuration tracks the duration of HTTP requests.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "xexplay",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestsTotal counts total HTTP requests.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xexplay",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	// ActiveWebSocketConnections tracks current WebSocket connections.
	ActiveWebSocketConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "xexplay",
			Subsystem: "ws",
			Name:      "active_connections",
			Help:      "Number of active WebSocket connections.",
		},
	)

	// ActiveGameSessions tracks current game sessions.
	ActiveGameSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "xexplay",
			Subsystem: "game",
			Name:      "active_sessions",
			Help:      "Number of active game sessions.",
		},
	)

)

// Metrics returns a Gin middleware that records HTTP request metrics.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		HTTPRequestDuration.WithLabelValues(method, path, status).Observe(time.Since(start).Seconds())
		HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	}
}
