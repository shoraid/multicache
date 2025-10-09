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

func (m *managerImpl) GetBoolOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue bool) (bool, error) {
	val, err := m.GetBool(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// If value not found or cast error, store default
		if err := m.Set(ctx, key, defaultValue, ttl); err != nil {
			return defaultValue, err
		}
		return defaultValue, nil
	}

	return false, err
}
