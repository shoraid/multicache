package multicache

import (
	"context"
	"errors"
	"time"
)

func (m *managerImpl) Get(ctx context.Context, key string) (any, error) {
	return m.store.Get(ctx, key)
}

func (m *managerImpl) GetOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (any, error)) (any, error) {
	val, err := m.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) {
		// Compute default lazily via callback
		defVal, defErr := defaultFn()
		if defErr != nil {
			return nil, defErr
		}

		// Try storing into cache
		if setErr := m.Set(ctx, key, defVal, ttl); setErr != nil {
			return defVal, setErr // still return computed value even if caching fails
		}

		return defVal, nil
	}

	return nil, err
}

func (m *managerImpl) Has(ctx context.Context, key string) (bool, error) {
	return m.store.Has(ctx, key)
}

func (m *managerImpl) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return m.store.Set(ctx, key, value, ttl)
}
