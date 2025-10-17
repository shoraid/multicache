package multicache

import (
	"context"
	"errors"
	"time"
)

// GetString retrieves a string value from the cache.
// Returns ErrTypeMismatch if the cached value is not a string.
func (m *Manager) GetString(ctx context.Context, key string) (string, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return "", err
	}

	stringVal, err := toString(val)
	if err != nil {
		return "", ErrTypeMismatch
	}

	return stringVal, nil
}

// GetStrings retrieves a []string value from the cache.
// Returns ErrTypeMismatch if the cached value is not a []string.
func (m *Manager) GetStrings(ctx context.Context, key string) ([]string, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	stringsVal, err := toStrings(val)
	if err != nil {
		return nil, ErrTypeMismatch
	}

	return stringsVal, nil
}

// GetStringOrSet works like GetOrSet but for string values.
func (m *Manager) GetStringOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (string, error)) (string, error) {
	val, err := m.GetString(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrTypeMismatch) {
		return "", err
	}

	return getOrSetDefault(ctx, m, key, ttl, defaultFn)
}

// GetStringsOrSet works like GetOrSet but for []string values.
func (m *Manager) GetStringsOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() ([]string, error)) ([]string, error) {
	val, err := m.GetStrings(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrTypeMismatch) {
		return nil, err
	}

	return getOrSetDefault(ctx, m, key, ttl, defaultFn)
}
