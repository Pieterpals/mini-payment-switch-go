package port

import (
	"context"
	"time"
)

// DistributedLock defines the contract for distributed locking (idempotency & concurrency control).
// Implementations live in the adapter layer (e.g., Redis SETNX).
type DistributedLock interface {
	// Acquire attempts to acquire a distributed lock with the given key and TTL.
	// Returns true if the lock was successfully acquired, false if already held by another process.
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Release explicitly releases a distributed lock.
	// This is optional if relying on TTL-based expiration, but recommended for fast release.
	Release(ctx context.Context, key string) error
}

// CacheStore defines the contract for key-value caching operations.
type CacheStore interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
}
