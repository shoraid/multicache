package contract

import (
	"context"
	"time"
)

// Manager defines the contract for a cache manager that provides
// a unified API to interact with multiple cache stores (e.g. memory,
// Redis, Valkey, file, etc). It supports both untyped and typed
// operations with lazy default value population.
//
// All implementations must be concurrency-safe.
type Manager interface {
	// Store switches the active cache store to the one registered under
	// the given alias. It returns a new Manager instance bound to that
	// store. If the alias does not exist, the implementation may return
	// a no-op manager or panic, depending on configuration.
	Store(alias string) Manager

	// Register adds a new store with the given alias.
	// The first registered store becomes the default.
	// Returns an error if the alias is already registered.
	Register(alias string, store Store) error

	// SetDefault sets the store with the given alias as the default store.
	// Returns an error if the alias has not been registered.
	SetDefault(alias string) error

	// Clear removes all keys and values from the current store.
	// It should not return an error if the store is already empty.
	Clear(ctx context.Context) error

	// Delete removes a single entry by key. If the key does not exist,
	// it should return nil (no error).
	Delete(ctx context.Context, key string) error

	// DeleteByPattern removes all keys matching the provided pattern.
	// The pattern syntax depends on the underlying driver (e.g. glob
	// for Redis, regex for memory). Drivers should document their behavior.
	DeleteByPattern(ctx context.Context, pattern string) error

	// DeleteMany removes multiple entries by their keys. If some keys
	// do not exist, they are skipped without returning an error.
	DeleteMany(ctx context.Context, keys ...string) error

	// DeleteManyByPattern removes entries matching any of the given patterns.
	// Behavior follows DeleteByPattern for each pattern.
	DeleteManyByPattern(ctx context.Context, patterns ...string) error

	// Get retrieves a raw cached value by key. It returns ErrCacheMiss
	// if the key is not found or has expired.
	Get(ctx context.Context, key string) (any, error)

	// GetOrSet retrieves a value from the cache if present; otherwise,
	// it computes the value lazily by calling defaultFn, stores it with
	// the given TTL, and returns it. If storing fails, it still returns
	// the computed value along with the store error.
	GetOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (any, error)) (any, error)

	// Has reports whether the given key exists and is not expired.
	// It should not return an error if the key simply doesn't exist.
	Has(ctx context.Context, key string) (bool, error)

	// Set stores a value in the cache under the given key with the specified TTL.
	// A ttl <= 0 should be treated as "no expiration" by convention, but this
	// behavior is driver-dependent.
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// GetBool retrieves a boolean value from the cache.
	// Returns ErrTypeMismatch if the cached value is not a bool.
	GetBool(ctx context.Context, key string) (bool, error)

	// GetBoolOrSet works like GetOrSet but for boolean values.
	GetBoolOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (bool, error)) (bool, error)

	// GetInt retrieves an int value from the cache.
	// Returns ErrTypeMismatch if the cached value is not an int.
	GetInt(ctx context.Context, key string) (int, error)

	// GetInt64 retrieves an int64 value from the cache.
	// Returns ErrTypeMismatch if the cached value is not an int64.
	GetInt64(ctx context.Context, key string) (int64, error)

	// GetInts retrieves a slice of int values from the cache.
	// Returns ErrTypeMismatch if the cached value is not a []int.
	GetInts(ctx context.Context, key string) ([]int, error)

	// GetIntOrSet works like GetOrSet but for int values.
	GetIntOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (int, error)) (int, error)

	// GetInt64OrSet works like GetOrSet but for int64 values.
	GetInt64OrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (int64, error)) (int64, error)

	// GetIntsOrSet works like GetOrSet but for []int values.
	GetIntsOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() ([]int, error)) ([]int, error)

	// GetString retrieves a string value from the cache.
	// Returns ErrTypeMismatch if the cached value is not a string.
	GetString(ctx context.Context, key string) (string, error)

	// GetStrings retrieves a []string value from the cache.
	// Returns ErrTypeMismatch if the cached value is not a []string.
	GetStrings(ctx context.Context, key string) ([]string, error)

	// GetStringOrSet works like GetOrSet but for string values.
	GetStringOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (string, error)) (string, error)

	// GetStringsOrSet works like GetOrSet but for []string values.
	GetStringsOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() ([]string, error)) ([]string, error)
}
