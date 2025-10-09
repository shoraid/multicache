package multicache

import (
	"context"
	"time"
)

func (m *managerImpl) Get(ctx context.Context, key string) (any, error) {
	return m.store.Get(ctx, key)
}

func (m *managerImpl) GetOrSet(ctx context.Context, key string, ttl time.Duration, value any) (any, error) {
	return m.store.GetOrSet(ctx, key, ttl, value)
}

func (m *managerImpl) Has(ctx context.Context, key string) (bool, error) {
	return m.store.Has(ctx, key)
}

func (m *managerImpl) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return m.store.Set(ctx, key, value, ttl)
}
