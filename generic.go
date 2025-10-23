package omnicache

import (
	"context"
	"errors"
	"time"
)

// GenericManager provides a generic way to interact with the cache for any type T.
// It wraps the Manager and provides type-safe Get and GetOrSet methods.
type GenericManager[T any] struct {
	m *Manager
}

// G returns a new GenericManager for the specified type T,
// bound to the given Manager instance. This allows for type-safe
// Get and GetOrSet operations.
func G[T any](m *Manager) *GenericManager[T] {
	return &GenericManager[T]{m: m}
}

// Get retrieves a value of type T from the cache.
// Returns ErrTypeMismatch if the cached value cannot be converted to T.
func (g *GenericManager[T]) Get(ctx context.Context, key string) (T, error) {
	val, err := g.m.Get(ctx, key)
	if err != nil {
		var zero T
		return zero, err
	}

	result, err := convertAnyToType[T](val)
	if err != nil {
		var zero T
		return zero, ErrTypeMismatch
	}

	return result, nil
}

// GetOrSet retrieves a value of type T from the cache if present; otherwise,
// it computes the value lazily by calling defaultFn, stores it with
// the given TTL, and returns it. If storing fails, it still returns
// the computed value along with the store error.
func (g *GenericManager[T]) GetOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (T, error)) (T, error) {
	val, err := g.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrTypeMismatch) {
		var zero T
		return zero, err
	}

	defaultValue, err := defaultFn()
	if err != nil {
		var zero T
		return zero, err
	}

	if err := g.m.Set(ctx, key, defaultValue, ttl); err != nil {
		return defaultValue, err
	}

	return defaultValue, nil
}
