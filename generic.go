package multicache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shoraid/multicache/contract"
)

type GenericManager[T any] struct {
	m contract.Manager
}

func G[T any](m contract.Manager) *GenericManager[T] {
	return &GenericManager[T]{m: m}
}

func (g *GenericManager[T]) Get(ctx context.Context, key string) (T, error) {
	val, err := g.m.Get(ctx, key)
	if err != nil {
		var zero T
		return zero, err
	}

	return convertAnyToType[T](val)
}

func (g *GenericManager[T]) GetOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (T, error)) (T, error) {
	val, err := g.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// Compute default lazily via callback
		defVal, err := defaultFn()
		if err != nil {
			var zero T
			return zero, err
		}

		// Try storing into cache
		if err := g.m.Set(ctx, key, defVal, ttl); err != nil {
			return defVal, err // still return computed value even if caching fails
		}

		return defVal, nil
	}

	var zero T
	return zero, err
}

func convertAnyToType[T any](val any) (T, error) {
	var zero T

	if v, ok := val.(T); ok {
		return v, nil
	}

	if bytes, ok := val.([]byte); ok {
		return parseFromBytes[T](bytes)
	}

	if str, ok := val.(string); ok {
		return parseFromString[T](str)
	}

	raw, err := json.Marshal(val)
	if err != nil {
		return zero, fmt.Errorf("convert: cannot marshal fallback (%T): %w", val, err)
	}

	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return zero, fmt.Errorf("convert: cannot unmarshal fallback to T: %w", err)
	}

	return result, nil
}

func parseFromBytes[T any](b []byte) (T, error) {
	var zero T
	var t T

	switch any(t).(type) {
	case string:
		return any(string(b)).(T), nil

	case int:
		n, err := strconv.Atoi(string(b))
		if err != nil {
			return zero, fmt.Errorf("convert: failed to parse int from bytes: %w", err)
		}
		return any(n).(T), nil

	case int64:
		n, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			return zero, fmt.Errorf("convert: failed to parse int64 from bytes: %w", err)
		}
		return any(n).(T), nil

	case bool:
		v, err := strconv.ParseBool(string(b))
		if err != nil {
			return zero, fmt.Errorf("convert: failed to parse bool from bytes: %w", err)
		}
		return any(v).(T), nil

	default:
		if err := json.Unmarshal(b, &t); err != nil {
			return zero, fmt.Errorf("convert: failed to unmarshal bytes to T: %w", err)
		}
		return t, nil
	}
}

func parseFromString[T any](s string) (T, error) {
	var zero T
	var t T

	switch any(t).(type) {
	case string:
		return any(s).(T), nil

	case int:
		n, err := strconv.Atoi(s)
		if err != nil {
			return zero, fmt.Errorf("convert: failed to parse int from string: %w", err)
		}
		return any(n).(T), nil

	case int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("convert: failed to parse int64 from string: %w", err)
		}
		return any(n).(T), nil

	case bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return zero, fmt.Errorf("convert: failed to parse bool from string: %w", err)
		}
		return any(v).(T), nil

	default:
		if err := json.Unmarshal([]byte(s), &t); err != nil {
			return zero, fmt.Errorf("convert: failed to unmarshal string to T: %w", err)
		}
		return t, nil
	}
}
