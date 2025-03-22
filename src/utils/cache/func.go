package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get retrieves a value from Redis by key
// If the key doesn't exist, it returns nil
// If there's an error, it returns the error
func (c *Client) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// GetJSON retrieves a JSON value from Redis by key and unmarshals it into the provided interface
// If the key doesn't exist, it returns nil
// If there's an error, it returns the error
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}
	return nil
}

// Set stores a value in Redis with the given key
// If there's an error, it returns the error
func (c *Client) Set(ctx context.Context, key string, value interface{}) error {
	if err := c.client.Set(ctx, key, value, 0).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// SetJSON stores a JSON value in Redis with the given key
// If there's an error, it returns the error
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}
	if err := c.client.Set(ctx, key, jsonData, 0).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// SetWithExpiry stores a value in Redis with the given key and expiration time
// If there's an error, it returns the error
func (c *Client) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	if err := c.client.Set(ctx, key, value, expiry).Err(); err != nil {
		return fmt.Errorf("failed to set key %s with expiry: %w", key, err)
	}
	return nil
}

// SetJSONWithExpiry stores a JSON value in Redis with the given key and expiration time
// If there's an error, it returns the error
func (c *Client) SetJSONWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}
	if err := c.client.Set(ctx, key, jsonData, expiry).Err(); err != nil {
		return fmt.Errorf("failed to set key %s with expiry: %w", key, err)
	}
	return nil
}
