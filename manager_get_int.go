package multicache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/spf13/cast"
)

func (m *managerImpl) GetInt(ctx context.Context, key string) (int, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	intVal, err := cast.ToIntE(val)
	if err != nil {
		return 0, ErrTypeMismatch
	}

	return intVal, nil
}

func (m *managerImpl) GetInt64(ctx context.Context, key string) (int64, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	intVal, err := cast.ToInt64E(val)
	if err != nil {
		return 0, ErrTypeMismatch
	}

	return intVal, nil
}

func (m *managerImpl) GetInts(ctx context.Context, key string) ([]int, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case string:
		var data []int
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return nil, ErrTypeMismatch
		}

		return data, nil

	case []byte:
		var data []int
		if err := json.Unmarshal(v, &data); err != nil {
			return nil, ErrTypeMismatch
		}

		return data, nil

	case []int:
		return v, nil

	default:
		return nil, ErrTypeMismatch
	}
}

func (m *managerImpl) GetIntOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue int) (int, error) {
	val, err := m.GetInt(ctx, key)
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

	return 0, err
}

func (m *managerImpl) GetInt64OrSet(ctx context.Context, key string, ttl time.Duration, defaultValue int64) (int64, error) {
	val, err := m.GetInt64(ctx, key)
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

	return 0, err
}

func (m *managerImpl) GetIntsOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue []int) ([]int, error) {
	val, err := m.GetInts(ctx, key)
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

	return nil, err
}
