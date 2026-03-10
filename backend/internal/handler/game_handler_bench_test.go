//go:build benchmark

// Package handler benchmarks for game session API handlers.
//
// These benchmarks exercise the HTTP handler layer directly using httptest,
// bypassing network overhead but still requiring a running GameService with
// database and cache backends.
//
// Run with:
//
//	cd backend && go test -tags benchmark -bench=. -benchmem ./internal/handler/ \
//	    -count=3 -benchtime=5s
//
// Prerequisites:
//   - PostgreSQL with the xexplay database and a published basket for today
//   - Redis cache instance
//   - Environment variables: DB_DSN, REDIS_URL (or defaults)
//   - At least one user in the database
//
// If the environment is not available, each benchmark will skip with b.Skip().
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/repository/redis"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

// testEnv holds shared resources across benchmarks.
type testEnv struct {
	handler *GameHandler
	userID  uuid.UUID
}

// setupBenchEnv initialises the GameService and handler from environment.
// Returns nil if the required infrastructure is unavailable.
func setupBenchEnv(b *testing.B) *testEnv {
	b.Helper()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/xexplay?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	db, err := postgres.NewConnection(dsn)
	if err != nil {
		b.Skipf("skipping benchmark: cannot connect to PostgreSQL: %v", err)
		return nil
	}

	redisClient, err := redis.NewConnection(redisURL)
	if err != nil {
		b.Skipf("skipping benchmark: cannot connect to Redis: %v", err)
		return nil
	}
	cache := redis.NewCacheRepo(redisClient)

	sessionRepo := postgres.NewSessionRepo(db)
	answerRepo := postgres.NewAnswerRepo(db)
	basketRepo := postgres.NewBasketRepo(db)
	cardRepo := postgres.NewCardRepo(db)
	userRepo := postgres.NewUserRepo(db)
	streakRepo := postgres.NewStreakRepo(db)

	shuffleService := service.NewShuffleService()
	streakService := service.NewStreakService(streakRepo)
	// achievementService left nil — not needed for handler benchmarks

	gameService := service.NewGameService(
		sessionRepo,
		answerRepo,
		basketRepo,
		cardRepo,
		userRepo,
		cache,
		shuffleService,
		streakService,
		nil, // achievementService
	)

	h := NewGameHandler(gameService)

	// Use a deterministic test user — override via BENCH_USER_ID env var.
	userIDStr := os.Getenv("BENCH_USER_ID")
	var uid uuid.UUID
	if userIDStr != "" {
		uid, err = uuid.Parse(userIDStr)
		if err != nil {
			b.Fatalf("invalid BENCH_USER_ID: %v", err)
		}
	} else {
		uid = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	return &testEnv{handler: h, userID: uid}
}

// newGinContext creates a minimal *gin.Context with the given method, path, body,
// and the user_id key set in the context.
func newGinContext(method, path string, body []byte, userID uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	return c, w
}

// BenchmarkStartSession measures POST /sessions/start handler throughput.
func BenchmarkStartSession(b *testing.B) {
	env := setupBenchEnv(b)
	if env == nil {
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c, _ := newGinContext(http.MethodPost, "/sessions/start", nil, env.userID)
		env.handler.StartSession(c)
	}
}

// BenchmarkGetCurrentCard measures GET /sessions/current/card handler throughput.
func BenchmarkGetCurrentCard(b *testing.B) {
	env := setupBenchEnv(b)
	if env == nil {
		return
	}

	// Ensure a session exists before benchmarking card retrieval.
	c, _ := newGinContext(http.MethodPost, "/sessions/start", nil, env.userID)
	env.handler.StartSession(c)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c, _ := newGinContext(http.MethodGet, "/sessions/current/card", nil, env.userID)
		env.handler.GetCurrentCard(c)
	}
}

// BenchmarkSubmitAnswer measures POST /sessions/current/answer handler throughput.
func BenchmarkSubmitAnswer(b *testing.B) {
	env := setupBenchEnv(b)
	if env == nil {
		return
	}

	// Ensure a session exists.
	c, _ := newGinContext(http.MethodPost, "/sessions/start", nil, env.userID)
	env.handler.StartSession(c)

	payload, _ := json.Marshal(submitAnswerRequest{Answer: true})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c, _ := newGinContext(http.MethodPost, "/sessions/current/answer", payload, env.userID)
		env.handler.SubmitAnswer(c)
	}
}

// BenchmarkSessionView measures GET /sessions/current (session summary) handler throughput.
func BenchmarkSessionView(b *testing.B) {
	env := setupBenchEnv(b)
	if env == nil {
		return
	}

	// Ensure a session exists.
	c, _ := newGinContext(http.MethodPost, "/sessions/start", nil, env.userID)
	env.handler.StartSession(c)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c, _ := newGinContext(http.MethodGet, "/sessions/current", nil, env.userID)
		env.handler.GetCurrentSession(c)
	}
}
