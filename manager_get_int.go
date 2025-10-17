package omnicache

import (
	"context"
	"errors"
	"time"
)

// GetInt retrieves an int value from the cache.
// Returns ErrTypeMismatch if the cached value is not an int.
func (m *Manager) GetInt(ctx context.Context, key string) (int, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	intVal, err := toInt(val)
	if err != nil {
		return 0, ErrTypeMismatch
	}

	return intVal, nil
}

// GetInt64 retrieves an int64 value from the cache.
// Returns ErrTypeMismatch if the cached value is not an int64.
func (m *Manager) GetInt64(ctx context.Context, key string) (int64, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	intVal, err := toInt64(val)
	if err != nil {
		return 0, ErrTypeMismatch
	}

	return intVal, nil
}

// GetInts retrieves a slice of int values from the cache.
// Returns ErrTypeMismatch if the cached value is not a []int.
func (m *Manager) GetInts(ctx context.Context, key string) ([]int, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	intsVal, err := toInts(val)
	if err != nil {
		return nil, ErrTypeMismatch
	}

	return intsVal, nil
}

// GetIntOrSet works like GetOrSet but for int values.
func (m *Manager) GetIntOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (int, error)) (int, error) {
	val, err := m.GetInt(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrTypeMismatch) {
		return 0, err
	}

	return getOrSetDefault(ctx, m, key, ttl, defaultFn)
}

// GetInt64OrSet works like GetOrSet but for int64 values.
func (m *Manager) GetInt64OrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (int64, error)) (int64, error) {
	val, err := m.GetInt64(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrTypeMismatch) {
		return 0, err
	}

	return getOrSetDefault(ctx, m, key, ttl, defaultFn)
}

// GetIntsOrSet works like GetOrSet but for []int values.
func (m *Manager) GetIntsOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() ([]int, error)) ([]int, error) {
	val, err := m.GetInts(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrTypeMismatch) {
		return nil, err
	}

	return getOrSetDefault(ctx, m, key, ttl, defaultFn)
}
