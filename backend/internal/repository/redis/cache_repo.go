package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/metrics"
)

const cacheTTL = 24 * time.Hour

type CacheRepo struct {
	client *Client
}

func NewCacheRepo(client *Client) *CacheRepo {
	return &CacheRepo{client: client}
}

func sessionKey(userID, date string) string {
	return fmt.Sprintf("session:%s:%s", userID, date)
}

func basketKey(date string) string {
	return fmt.Sprintf("basket:%s", date)
}

func (r *CacheRepo) GetSession(ctx context.Context, userID, date string) (*domain.UserSession, error) {
	rdb := r.client.Underlying()
	data, err := rdb.Get(ctx, sessionKey(userID, date)).Bytes()
	if err != nil {
		if err == goredis.Nil {
			metrics.CacheMissesTotal.WithLabelValues("session").Inc()
			return nil, nil
		}
		return nil, fmt.Errorf("cache get session: %w", err)
	}

	metrics.CacheHitsTotal.WithLabelValues("session").Inc()

	var session domain.UserSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("cache unmarshal session: %w", err)
	}
	return &session, nil
}

func (r *CacheRepo) SetSession(ctx context.Context, userID, date string, session *domain.UserSession) error {
	rdb := r.client.Underlying()
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("cache marshal session: %w", err)
	}

	if err := rdb.Set(ctx, sessionKey(userID, date), data, cacheTTL).Err(); err != nil {
		return fmt.Errorf("cache set session: %w", err)
	}
	return nil
}

func (r *CacheRepo) DeleteSession(ctx context.Context, userID, date string) error {
	rdb := r.client.Underlying()
	if err := rdb.Del(ctx, sessionKey(userID, date)).Err(); err != nil {
		return fmt.Errorf("cache delete session: %w", err)
	}
	return nil
}

func (r *CacheRepo) GetBasket(ctx context.Context, date string) (string, error) {
	rdb := r.client.Underlying()
	data, err := rdb.Get(ctx, basketKey(date)).Result()
	if err != nil {
		if err == goredis.Nil {
			metrics.CacheMissesTotal.WithLabelValues("basket").Inc()
			return "", nil
		}
		return "", fmt.Errorf("cache get basket: %w", err)
	}

	metrics.CacheHitsTotal.WithLabelValues("basket").Inc()
	return data, nil
}

func (r *CacheRepo) SetBasket(ctx context.Context, date, jsonData string, ttl time.Duration) error {
	rdb := r.client.Underlying()
	if err := rdb.Set(ctx, basketKey(date), jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("cache set basket: %w", err)
	}
	return nil
}
