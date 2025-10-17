package multicache

import (
	"context"
	"errors"
	"time"
)

// Get retrieves a raw cached value by key. It returns ErrCacheMiss
// if the key is not found or has expired.
func (m *Manager) Get(ctx context.Context, key string) (any, error) {
	return m.store.Get(ctx, key)
}

// GetOrSet retrieves a value from the cache if present; otherwise,
// it computes the value lazily by calling defaultFn, stores it with
// the given TTL, and returns it. If storing fails, it still returns
// the computed value along with the store error.
func (m *Manager) GetOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (any, error)) (any, error) {
	val, err := m.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) {
		return nil, err
	}

	return getOrSetDefault(ctx, m, key, ttl, defaultFn)
}

// Has reports whether the given key exists and is not expired.
// It should not return an error if the key simply doesn't exist.
func (m *Manager) Has(ctx context.Context, key string) (bool, error) {
	return m.store.Has(ctx, key)
}

// Set stores a value in the cache under the given key with the specified TTL.
// A ttl <= 0 should be treated as "no expiration" by convention, but this
// behavior is driver-dependent.
func (m *Manager) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return m.store.Set(ctx, key, value, ttl)
}
