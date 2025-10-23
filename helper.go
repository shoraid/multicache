package omnicache

import (
	"fmt"
	"strconv"

	"github.com/bytedance/sonic"
)

// convertAnyToType converts any Go value to a target generic type T.
// It supports primitives, []byte, and slices efficiently.
func convertAnyToType[T any](v any) (T, error) {
	var zero T

	if v == nil {
		return zero, ErrTypeMismatch
	}

	switch val := any(v).(type) {

	// --- Fast paths for primitives ---
	case T:
		return val, nil

	case string:
		return fromString[T](val)

	case []byte:
		return fromBytes[T](val)

	case int:
		return fromStringOrNumber[T](strconv.Itoa(val))

	case int64:
		return fromStringOrNumber[T](strconv.FormatInt(val, 10))

	case float64:
		return fromStringOrNumber[T](strconv.FormatFloat(val, 'f', -1, 64))

	case bool:
		return fromStringOrNumber[T](strconv.FormatBool(val))

	default:
		var out T

		// --- convert struct/slice to string
		if _, ok := any(zero).(string); ok {
			b, err := sonic.Marshal(val)
			if err != nil {
				return zero, fmt.Errorf("marshal error: %w", err)
			}
			return any(string(b)).(T), nil
		}

		// --- Fallback: JSON round-trip
		b, err := sonic.Marshal(val)
		if err != nil {
			return zero, fmt.Errorf("marshal error: %w", err)
		}
		if err := sonic.Unmarshal(b, &out); err != nil {
			return zero, fmt.Errorf("unmarshal error: %w", err)
		}
		return out, nil
	}
}

// fromString converts a string to a target generic type T.
func fromString[T any](s string) (T, error) {
	var zero T
	var result any
	var err error

	switch any(zero).(type) {
	case string:
		result = s
	case bool:
		result, err = strconv.ParseBool(s)
	case int:
		var n int64
		n, err = strconv.ParseInt(s, 10, 0)
		result = int(n)
	case int64:
		result, err = strconv.ParseInt(s, 10, 64)
	case float64:
		result, err = strconv.ParseFloat(s, 64)
	case []byte:
		result = []byte(s)
	default:
		var out T
		if err := sonic.Unmarshal([]byte(s), &out); err != nil {
			return zero, fmt.Errorf("unsupported conversion from string to %T: %w", zero, err)
		}
		return out, nil
	}

	if err != nil {
		return zero, err
	}

	return result.(T), nil
}

// fromBytes converts a byte slice to a target generic type T.
func fromBytes[T any](b []byte) (T, error) {
	var zero T
	var result any
	var err error

	switch any(zero).(type) {
	case string:
		result = string(b)
	case bool:
		result, err = strconv.ParseBool(string(b))
	case int:
		var n int64
		n, err = strconv.ParseInt(string(b), 10, 0)
		result = int(n)
	case int64:
		result, err = strconv.ParseInt(string(b), 10, 64)
	case float64:
		result, err = strconv.ParseFloat(string(b), 64)
	default:
		// Fallback for struct or unknown type
		var out T
		if err := sonic.Unmarshal(b, &out); err != nil {
			return zero, fmt.Errorf("unsupported conversion from string to %T: %w", zero, err)
		}
		return out, nil
	}

	if err != nil {
		return zero, err
	}

	return result.(T), nil
}

// fromStringOrNumber converts a string to a target generic type T.
func fromStringOrNumber[T any](s string) (T, error) {
	return fromString[T](s)
}
