package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/saas-multi-database-api/internal/config"
)

type Client struct {
	Client *redis.Client // Expor publicamente para uso no TenantService
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{Client: client}, nil
}

// Get retrieves a value from Redis
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Set sets a value in Redis with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// Delete removes a key from Redis
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in Redis
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.Client.Exists(ctx, key).Result()
	return count > 0, err
}

// GetDBCode retrieves the db_code for a given url_code from cache
func (c *Client) GetDBCode(ctx context.Context, urlCode string) (string, error) {
	key := fmt.Sprintf("tenant:urlcode:%s", urlCode)
	return c.Get(ctx, key)
}

// SetDBCode caches the db_code for a url_code
func (c *Client) SetDBCode(ctx context.Context, urlCode, dbCode string, expiration time.Duration) error {
	key := fmt.Sprintf("tenant:urlcode:%s", urlCode)
	return c.Set(ctx, key, dbCode, expiration)
}

// InvalidateTenantCache removes all cached data for a tenant
func (c *Client) InvalidateTenantCache(ctx context.Context, urlCode string) error {
	key := fmt.Sprintf("tenant:urlcode:%s", urlCode)
	return c.Delete(ctx, key)
}

// Publish publishes a message to a Redis channel
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.Client.Publish(ctx, channel, message).Err()
}

// Close closes the Redis client
func (c *Client) Close() error {
	return c.Client.Close()
}
