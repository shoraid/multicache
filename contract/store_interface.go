package contract

import (
	"context"
	"time"
)

// Store defines the minimal caching operations that must be supported
// by any storage driver (e.g. memory, Redis, Valkey, file, DB).
//
// All methods should be safe for concurrent use.
type Store interface {
	// Clear removes all entries from the store.
	Clear(ctx context.Context) error

	// Close releases any resources held by the store.
	// Should be safe to call multiple times.
	Close(ctx context.Context) error

	// Delete removes a single entry by key.
	// If the key does not exist, it should return nil (no error).
	Delete(ctx context.Context, key string) error

	// DeleteByPattern removes entries matching a specific pattern.
	// The pattern syntax depends on the implementation (e.g. glob, regex).
	DeleteByPattern(ctx context.Context, pattern string) error

	// DeleteMany removes multiple entries by their keys.
	// If some keys do not exist, they are skipped without error.
	DeleteMany(ctx context.Context, keys ...string) error

	// Get retrieves the value associated with the given key.
	// It returns ErrCacheMiss if the key is not found or expired.
	Get(ctx context.Context, key string) (any, error)

	// Has checks whether a key exists and is not expired.
	// It should not return an error if the key simply doesn't exist.
	Has(ctx context.Context, key string) (bool, error)

	// Set stores a value with the given key and TTL.
	// If ttl <= 0, the behavior is driver-dependent (e.g. no expiration).
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}
