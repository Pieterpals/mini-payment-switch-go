package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisDistributedLock is the Redis implementation of port.DistributedLock.
// Uses SETNX (SET if Not eXists) for atomic lock acquisition.
type RedisDistributedLock struct {
	client *redis.Client
}

// NewRedisDistributedLock creates a new Redis-backed distributed lock.
func NewRedisDistributedLock(client *redis.Client) *RedisDistributedLock {
	return &RedisDistributedLock{client: client}
}

// Acquire attempts to acquire a lock using Redis SETNX with TTL.
// Returns true if the lock was acquired, false if it's already held.
func (l *RedisDistributedLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	acquired, err := l.client.SetNX(ctx, key, "locked", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis: failed to acquire lock: %w", err)
	}
	return acquired, nil
}

// Release explicitly deletes the lock key from Redis.
func (l *RedisDistributedLock) Release(ctx context.Context, key string) error {
	if err := l.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis: failed to release lock: %w", err)
	}
	return nil
}
