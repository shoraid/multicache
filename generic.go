package multicache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// GenericManager provides a generic way to interact with the cache for any type T.
// It wraps the Manager and provides type-safe Get and GetOrSet methods.
type GenericManager[T any] struct {
	m *Manager
}

// G returns a new GenericManager for the specified type T,
// bound to the given Manager instance. This allows for type-safe
// Get and GetOrSet operations.
func G[T any](m *Manager) *GenericManager[T] {
	return &GenericManager[T]{m: m}
}

// Get retrieves a value of type T from the cache.
// Returns ErrTypeMismatch if the cached value cannot be converted to T.
func (g *GenericManager[T]) Get(ctx context.Context, key string) (T, error) {
	val, err := g.m.Get(ctx, key)
	if err != nil {
		var zero T
		return zero, err
	}

	return convertAnyToType[T](val)
}

// GetOrSet retrieves a value of type T from the cache if present; otherwise,
// it computes the value lazily by calling defaultFn, stores it with
// the given TTL, and returns it. If storing fails, it still returns
// the computed value along with the store error.
func (g *GenericManager[T]) GetOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (T, error)) (T, error) {
	val, err := g.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrTypeMismatch) {
		var zero T
		return zero, err
	}

	return getOrSetDefault(ctx, g.m, key, ttl, defaultFn)
}

// convertAnyToType attempts to convert an `any` type to a specific generic type `T`.
// It handles direct type assertion, byte slices, strings, and falls back to JSON marshaling/unmarshaling.
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

// parseFromBytes attempts to convert a byte slice to a specific generic type T.
// It handles common primitive types and falls back to JSON unmarshaling.
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

// parseFromString attempts to convert a string to a specific generic type T.
// It handles common primitive types and falls back to JSON unmarshaling.
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
