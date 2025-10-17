package omnicache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// toBool attempts to convert various types to bool.
// It supports bool, string, []byte, int, int64, float64, and float32.
// Returns an error if the value cannot be reasonably converted.
func toBool(val any) (bool, error) {
	switch v := val.(type) {
	case bool:
		return v, nil

	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return false, err
		}
		return parsed, nil

	case []byte:
		parsed, err := strconv.ParseBool(string(v))
		if err != nil {
			return false, err
		}
		return parsed, nil

	case int:
		return v != 0, nil

	case int64:
		return v != 0, nil

	case float64:
		return v != 0, nil

	case float32:
		return v != 0, nil

	default:
		return false, fmt.Errorf("convert: cannot cast %T to bool", val)
	}
}

// toInt attempts to convert various types to int.
// It supports int, int64, float64, string, []byte, and bool.
// Returns an error if the value cannot be reasonably converted.
func toInt(val any) (int, error) {
	switch v := val.(type) {
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("convert: failed to parse int from string: %w", err)
		}
		return n, nil

	case []byte:
		n, err := strconv.Atoi(string(v))
		if err != nil {
			return 0, fmt.Errorf("convert: failed to parse int from bytes: %w", err)
		}
		return n, nil

	case int:
		return v, nil

	case int8:
		return int(v), nil

	case int16:
		return int(v), nil

	case int32:
		return int(v), nil

	case int64:
		return int(v), nil

	case uint:
		return int(v), nil

	case uint8:
		return int(v), nil

	case uint16:
		return int(v), nil

	case uint32:
		return int(v), nil

	case uint64:
		return int(v), nil

	case float32:
		return int(v), nil

	case float64:
		return int(v), nil

	case bool:
		if v {
			return 1, nil
		}
		return 0, nil

	default:
		return 0, fmt.Errorf("convert: cannot cast %T to int", val)
	}
}

// toInt64 attempts to convert various types to int64.
// It supports int, uint, float, bool, string, and []byte.
// Returns an error if the type cannot be reasonably converted.
func toInt64(val any) (int64, error) {
	switch v := val.(type) {
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("convert: failed to parse int64 from string: %w", err)
		}
		return n, nil

	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("convert: failed to parse int64 from bytes: %w", err)
		}
		return n, nil

	case int:
		return int64(v), nil

	case int8:
		return int64(v), nil

	case int16:
		return int64(v), nil

	case int32:
		return int64(v), nil

	case int64:
		return v, nil

	case uint:
		return int64(v), nil

	case uint8:
		return int64(v), nil

	case uint16:
		return int64(v), nil

	case uint32:
		return int64(v), nil

	case uint64:
		return int64(v), nil

	case float32:
		return int64(v), nil

	case float64:
		return int64(v), nil

	case bool:
		if v {
			return 1, nil
		}
		return 0, nil

	default:
		return 0, fmt.Errorf("convert: cannot cast %T to int64", val)
	}
}

// toInts attempts to convert any value into a []int.
// It handles multiple common representations returned by different cache backends.
func toInts(val any) ([]int, error) {
	switch v := val.(type) {
	case []int:
		return v, nil

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

	case []float64:
		ints := make([]int, len(v))
		for i, num := range v {
			ints[i] = int(num)
		}
		return ints, nil

	case []any:
		ints := make([]int, len(v))
		for i, elem := range v {
			switch num := elem.(type) {
			case float64:
				ints[i] = int(num)
			case int:
				ints[i] = num
			default:
				return nil, ErrTypeMismatch
			}
		}
		return ints, nil

	default:
		return nil, ErrTypeMismatch
	}
}

// toString attempts to convert any value into a string.
// It handles various primitive types and falls back to JSON marshaling for complex types.
// Returns an error if the value cannot be reasonably converted or marshaled.
func toString(val any) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil

	case []byte:
		return string(v), nil

	case int:
		return strconv.Itoa(v), nil

	case int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10), nil

	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10), nil

	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil

	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil

	case bool:
		return strconv.FormatBool(v), nil

	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("convert: cannot cast %T to string and failed to marshal to JSON: %w", val, err)
		}
		return string(jsonBytes), nil
	}
}

// toStrings attempts to convert any value into a []string.
// It handles multiple common representations returned by different cache backends.
func toStrings(val any) ([]string, error) {
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
		// Attempt a generic fallback via JSON round-trip
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, ErrTypeMismatch
		}

		var data []string
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, ErrTypeMismatch
		}

		return data, nil
	}
}

// getOrSetDefault is a helper function to handle the common pattern of
// getting a value from cache, and if it's a cache miss or type mismatch,
// computing a default value and setting it in the cache.
func getOrSetDefault[T any](ctx context.Context, m *Manager, key string, ttl time.Duration, defaultFn func() (T, error)) (T, error) {
	val, err := defaultFn()
	if err != nil {
		var zero T
		return zero, err
	}

	if err := m.Set(ctx, key, val, ttl); err != nil {
		return val, err
	}

	return val, nil
}
