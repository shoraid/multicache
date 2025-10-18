package omnicache

import (
	"context"
	"sync"
)

// Clear removes all keys and values from the current store.
// It should not return an error if the store is already empty.
func (m *Manager) Clear(ctx context.Context) error {
	return m.store.Clear(ctx)
}

// Delete removes a single entry by key. If the key does not exist,
// it should return nil (no error).
func (m *Manager) Delete(ctx context.Context, key string) error {
	return m.store.Delete(ctx, key)
}

// DeleteByPattern removes all keys matching the provided pattern.
// The pattern syntax depends on the underlying driver (e.g. glob
// for Redis, regex for memory). Drivers should document their behavior.
func (m *Manager) DeleteByPattern(ctx context.Context, pattern string) error {
	return m.store.DeleteByPattern(ctx, pattern)
}

// DeleteMany removes multiple entries by their keys. If some keys
// do not exist, they are skipped without returning an error.
func (m *Manager) DeleteMany(ctx context.Context, keys ...string) error {
	return m.store.DeleteMany(ctx, keys...)
}

// DeleteManyByPattern removes entries matching any of the given patterns.
// Behavior follows DeleteByPattern for each pattern.
func (m *Manager) DeleteManyByPattern(ctx context.Context, patterns ...string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(patterns))

	for _, pattern := range patterns {
		p := pattern
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.store.DeleteByPattern(ctx, p); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}
	return nil
}
