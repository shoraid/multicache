package multicache

import (
	"context"
	"errors"
	"time"

	"github.com/shoraid/multicache/contract"
)

type GenericManager[T any] struct {
	m contract.Manager
}

func G[T any](m contract.Manager) *GenericManager[T] {
	return &GenericManager[T]{m: m}
}

func (g *GenericManager[T]) Get(ctx context.Context, key string) (T, error) {
	val, err := g.m.Get(ctx, key)
	if err != nil {
		var zero T
		return zero, err
	}

	typed, ok := val.(T)
	if !ok {
		var zero T
		return zero, ErrTypeMismatch
	}

	return typed, nil
}

func (g *GenericManager[T]) GetOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (T, error)) (T, error) {
	val, err := g.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// Compute default lazily via callback
		defVal, err := defaultFn()
		if err != nil {
			var zero T
			return zero, err
		}

		// Try storing into cache
		if err := g.m.Set(ctx, key, defVal, ttl); err != nil {
			return defVal, err // still return computed value even if caching fails
		}

		return defVal, nil
	}

	var zero T
	return zero, err
}
