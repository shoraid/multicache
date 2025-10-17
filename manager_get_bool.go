package multicache

import (
	"context"
	"errors"
	"time"
)

// GetBool retrieves a boolean value from the cache.
// Returns ErrTypeMismatch if the cached value is not a bool.
func (m *Manager) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return false, err
	}

	boolVal, err := toBool(val)
	if err != nil {
		return false, ErrTypeMismatch
	}

	return boolVal, nil
}

// GetBoolOrSet works like GetOrSet but for boolean values.
func (m *Manager) GetBoolOrSet(
	ctx context.Context,
	key string,
	ttl time.Duration,
	defaultFn func() (bool, error),
) (bool, error) {
	val, err := m.GetBool(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrTypeMismatch) {
		return false, err
	}

	return getOrSetDefault(ctx, m, key, ttl, defaultFn)
}
