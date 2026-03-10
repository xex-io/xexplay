package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Client struct {
	rdb *goredis.Client
}

func NewConnection(redisURL string) (*Client, error) {
	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	rdb := goredis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Info().Msg("connected to Redis")
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	log.Info().Msg("Redis connection closed")
	return c.rdb.Close()
}

func (c *Client) Underlying() *goredis.Client {
	return c.rdb
}
