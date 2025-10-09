package multicache

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

func (m *managerImpl) GetString(ctx context.Context, key string) (string, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return "", err
	}

	switch v := val.(type) {
	case string:
		return v, nil

	case []byte:
		return string(v), nil

	default:
		return "", ErrTypeMismatch
	}
}

func (m *managerImpl) GetStrings(ctx context.Context, key string) ([]string, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case string:
		var data []string
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return nil, ErrTypeMismatch
		}
		return data, nil

	case []byte:
		var data []string
		if err := json.Unmarshal(v, &data); err != nil {
			return nil, ErrTypeMismatch
		}
		return data, nil

	case []string:
		return v, nil

	default:
		return nil, ErrTypeMismatch
	}
}

func (m *managerImpl) GetStringOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue string) (string, error) {
	val, err := m.GetString(ctx, key)
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

	return "", err
}

func (m *managerImpl) GetStringsOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue []string) ([]string, error) {
	val, err := m.GetStrings(ctx, key)
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
