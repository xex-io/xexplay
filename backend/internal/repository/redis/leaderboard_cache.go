package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/metrics"
)

type LeaderboardCache struct {
	client *Client
}

func NewLeaderboardCache(client *Client) *LeaderboardCache {
	return &LeaderboardCache{client: client}
}

// LeaderboardKey builds a Redis sorted set key for a leaderboard period.
func LeaderboardKey(periodType, periodKey string) string {
	return fmt.Sprintf("lb:%s:%s", periodType, periodKey)
}

// UpdateScore increments a user's score in the sorted set using ZINCRBY.
func (c *LeaderboardCache) UpdateScore(ctx context.Context, key string, userID string, points float64) error {
	rdb := c.client.Underlying()
	if err := rdb.ZIncrBy(ctx, key, points, userID).Err(); err != nil {
		return fmt.Errorf("leaderboard cache update score: %w", err)
	}
	return nil
}

// GetTopN returns the top N entries from the sorted set (highest scores first).
func (c *LeaderboardCache) GetTopN(ctx context.Context, key string, n int64) ([]domain.LeaderboardRow, error) {
	rdb := c.client.Underlying()
	results, err := rdb.ZRevRangeWithScores(ctx, key, 0, n-1).Result()
	if err != nil {
		if err == goredis.Nil {
			metrics.CacheMissesTotal.WithLabelValues("leaderboard").Inc()
			return nil, nil
		}
		return nil, fmt.Errorf("leaderboard cache get top n: %w", err)
	}

	if len(results) == 0 {
		metrics.CacheMissesTotal.WithLabelValues("leaderboard").Inc()
	} else {
		metrics.CacheHitsTotal.WithLabelValues("leaderboard").Inc()
	}

	rows := make([]domain.LeaderboardRow, 0, len(results))
	for i, z := range results {
		userID := z.Member.(string)
		uid, err := uuid.Parse(userID)
		if err != nil {
			continue
		}
		rows = append(rows, domain.LeaderboardRow{
			Rank:        i + 1,
			UserID:      uid,
			TotalPoints: int(z.Score),
		})
	}
	return rows, nil
}

// GetUserRank returns the 0-based rank of a user (ZREVRANK: rank 0 = highest score).
// Returns -1 if the user is not found in the set.
func (c *LeaderboardCache) GetUserRank(ctx context.Context, key string, userID string) (int64, error) {
	rdb := c.client.Underlying()
	rank, err := rdb.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		if err == goredis.Nil {
			metrics.CacheMissesTotal.WithLabelValues("leaderboard").Inc()
			return -1, nil
		}
		return -1, fmt.Errorf("leaderboard cache get user rank: %w", err)
	}
	metrics.CacheHitsTotal.WithLabelValues("leaderboard").Inc()
	return rank, nil
}

// GetUserScore returns the user's score in the sorted set.
// Returns 0 if the user is not found.
func (c *LeaderboardCache) GetUserScore(ctx context.Context, key string, userID string) (float64, error) {
	rdb := c.client.Underlying()
	score, err := rdb.ZScore(ctx, key, userID).Result()
	if err != nil {
		if err == goredis.Nil {
			metrics.CacheMissesTotal.WithLabelValues("leaderboard").Inc()
			return 0, nil
		}
		return 0, fmt.Errorf("leaderboard cache get user score: %w", err)
	}
	metrics.CacheHitsTotal.WithLabelValues("leaderboard").Inc()
	return score, nil
}

// SetExpiry sets a TTL on a leaderboard sorted set key.
func (c *LeaderboardCache) SetExpiry(ctx context.Context, key string, duration time.Duration) error {
	rdb := c.client.Underlying()
	if err := rdb.Expire(ctx, key, duration).Err(); err != nil {
		return fmt.Errorf("leaderboard cache set expiry: %w", err)
	}
	return nil
}

// DeleteKey removes a leaderboard sorted set key.
func (c *LeaderboardCache) DeleteKey(ctx context.Context, key string) error {
	rdb := c.client.Underlying()
	if err := rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("leaderboard cache delete key: %w", err)
	}
	return nil
}
