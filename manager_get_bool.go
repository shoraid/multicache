package multicache

import (
	"context"
	"errors"
	"time"
)

func (m *managerImpl) GetBool(ctx context.Context, key string) (bool, error) {
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

func (m *managerImpl) GetBoolOrSet(
	ctx context.Context,
	key string,
	ttl time.Duration,
	defaultFn func() (bool, error),
) (bool, error) {
	val, err := m.GetBool(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		return getOrSetDefault(ctx, m, key, ttl, defaultFn)
	}

	return false, err
}
