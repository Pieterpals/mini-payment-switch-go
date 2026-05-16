package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

// RedisCache implements the port.CacheStore using go-redis.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new RedisCache instance.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Set stores a value in Redis with the given TTL.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "Redis.Set")
	defer span.End()

	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := c.client.Set(ctx, key, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache key %s: %w", key, err)
	}
	return nil
}

// Get retrieves a value from Redis and unmarshals it into dest.
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "Redis.Get")
	defer span.End()

	bytes, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("cache key %s not found", key) // or return a specific not found error
		}
		return fmt.Errorf("failed to get cache key %s: %w", key, err)
	}

	if err := json.Unmarshal(bytes, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}
	return nil
}

// Delete removes a key from Redis.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "Redis.Delete")
	defer span.End()

	return c.client.Del(ctx, key).Err()
}
