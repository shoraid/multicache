package multicache

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cast"
)

func (m *managerImpl) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return false, err
	}

	boolVal, err := cast.ToBoolE(val)
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
		// Compute default lazily via callback
		defVal, defErr := defaultFn()
		if defErr != nil {
			return false, defErr
		}

		// Try storing into cache
		if setErr := m.Set(ctx, key, defVal, ttl); setErr != nil {
			return defVal, setErr // still return computed value even if caching fails
		}

		return defVal, nil
	}

	return false, err
}
