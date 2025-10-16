package multicache

import (
	"context"
	"errors"
	"time"
)

func (m *managerImpl) GetString(ctx context.Context, key string) (string, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return "", err
	}

	stringVal, err := toString(val)
	if err != nil {
		return "", ErrTypeMismatch
	}

	return stringVal, nil
}

func (m *managerImpl) GetStrings(ctx context.Context, key string) ([]string, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	stringsVal, err := toStrings(val)
	if err != nil {
		return nil, ErrTypeMismatch
	}

	return stringsVal, nil
}

func (m *managerImpl) GetStringOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (string, error)) (string, error) {
	val, err := m.GetString(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		return getOrSetDefault(ctx, m, key, ttl, defaultFn)
	}

	return "", err
}

func (m *managerImpl) GetStringsOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() ([]string, error)) ([]string, error) {
	val, err := m.GetStrings(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		return getOrSetDefault(ctx, m, key, ttl, defaultFn)
	}

	return nil, err
}
