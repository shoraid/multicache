package multicache

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (m *managerImpl) Clear(ctx context.Context) error {
	return m.store.Clear(ctx)
}

func (m *managerImpl) Delete(ctx context.Context, key string) error {
	return m.store.Delete(ctx, key)
}

func (m *managerImpl) DeleteByPattern(ctx context.Context, pattern string) error {
	return m.store.DeleteByPattern(ctx, pattern)
}

func (m *managerImpl) DeleteMany(ctx context.Context, keys ...string) error {
	return m.store.DeleteMany(ctx, keys...)
}

func (m *managerImpl) DeleteManyByPattern(ctx context.Context, patterns ...string) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, pattern := range patterns {
		p := pattern
		g.Go(func() error {
			return m.store.DeleteByPattern(ctx, p)
		})
	}

	return g.Wait() // returns first non-nil error (or nil if all succeeded)
}
