package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/config"
	"github.com/redis/go-redis/v9"
)

// Cache defines interface for cache operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type redisCache struct {
	client *redis.Client
}

// New creates a new Redis cache instance
func New(cfg *config.Config) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisCache{
		client: client,
	}, nil
}

// Get retrieves a value from cache and unmarshals it into dest
func (c *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			RecordMiss()
			return fmt.Errorf("cache miss")
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	RecordHit()
	return nil
}

// Set stores a value in cache with TTL
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (c *redisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}
	return exists > 0, nil
}

// Close closes the Redis connection
func (c *redisCache) Close() error {
	return c.client.Close()
}
