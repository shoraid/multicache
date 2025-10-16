package multicache

import (
	"context"
	"errors"
	"time"
)

func (m *managerImpl) GetInt(ctx context.Context, key string) (int, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	intVal, err := toInt(val)
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

	intVal, err := toInt64(val)
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

	intsVal, err := toInts(val)
	if err != nil {
		return nil, ErrTypeMismatch
	}

	return intsVal, nil
}

func (m *managerImpl) GetIntOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (int, error)) (int, error) {
	val, err := m.GetInt(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		return getOrSetDefault(ctx, m, key, ttl, defaultFn)
	}

	return 0, err
}

func (m *managerImpl) GetInt64OrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (int64, error)) (int64, error) {
	val, err := m.GetInt64(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		return getOrSetDefault(ctx, m, key, ttl, defaultFn)
	}

	return 0, err
}

func (m *managerImpl) GetIntsOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() ([]int, error)) ([]int, error) {
	val, err := m.GetInts(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		return getOrSetDefault(ctx, m, key, ttl, defaultFn)
	}

	return nil, err
}
