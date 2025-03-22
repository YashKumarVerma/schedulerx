package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/yashkumarverma/schedulerx/src/utils"
)

// Client wraps the Redis client
type Client struct {
	client *redis.Client
}

// Ping tests the Redis connection
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func NewClient(ctx context.Context, config *utils.Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", config.CacheClusterURL, "6379")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.CachePassword,
		Username: config.CacheUsername,
		DB:       0,
	})

	// Test the connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		client: rdb,
	}, nil
}
