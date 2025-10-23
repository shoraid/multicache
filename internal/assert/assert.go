package assert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Error asserts that err is not nil.
func Error(t *testing.T, err error, msg ...any) {
	t.Helper()

	message := "expected error but got nil"
	if len(msg) > 0 {
		message = fmt.Sprint(msg...)
	}

	if err == nil {
		t.Fatal(message)
	}
}

// NoError asserts that err is nil.
func NoError(t *testing.T, err error, msg ...any) {
	t.Helper()

	message := "expected no error, but got error"
	if len(msg) > 0 {
		message = fmt.Sprint(msg...)
	}

	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// Equal compares two comparable values and fails if they differ.
func Equal(t *testing.T, expected, got any, msg ...any) {
	t.Helper()

	// build message
	message := fmt.Sprintf("expected %v, got %v", expected, got)
	if len(msg) > 0 {
		message = fmt.Sprint(msg...)
	}

	// handle nils
	if expected == nil && got == nil {
		return
	}

	ev := reflect.ValueOf(expected)
	gv := reflect.ValueOf(got)

	// handle slice/map nil vs empty equivalence
	if ev.IsValid() && gv.IsValid() &&
		ev.Kind() == gv.Kind() &&
		(ev.Kind() == reflect.Slice || ev.Kind() == reflect.Map) {

		if ev.Len() == 0 && gv.IsNil() || gv.Len() == 0 && ev.IsNil() {
			return // treat nil and empty as equal
		}
	}

	// use DeepEqual for everything else
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("%s\nexpected: %#v\n     got: %#v", message, expected, got)
	}
}

// Empty asserts that the given value is empty, nil, or has zero length.
// It fails the test if the value is not empty.
// Optionally, you can provide a custom message.
func Empty(t *testing.T, v any, msg ...any) {
	t.Helper()

	if isEmpty(v) {
		return
	}

	message := fmt.Sprintf("expected value to be empty, but got: %#v", v)
	if len(msg) > 0 {
		message = fmt.Sprint(msg...)
	}
	t.Fatal(message)
}

// isEmpty checks whether a value is considered "empty".
func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	default:
		// For structs, compare against zero value
		zero := reflect.Zero(rv.Type())
		return reflect.DeepEqual(v, zero.Interface())
	}
}

// NotNil asserts that value is not nil.
func NotNil(t *testing.T, v any, msg ...any) {
	t.Helper()

	message := "expected non-nil value"
	if len(msg) > 0 {
		message = fmt.Sprint(msg...)
	}

	if v == nil {
		t.Fatal(message)
	}
}

// Nil asserts that value is nil.
func Nil(t *testing.T, v any, msg ...any) {
	t.Helper()

	message := "expected nil value"
	if len(msg) > 0 {
		message = fmt.Sprint(msg...)
	}

	if v != nil {
		t.Fatalf("%s: got %#v", message, v)
	}
}

// Contains asserts that the given container contains the specified key.
// It supports string (substring check) and map types.
func Contains(t *testing.T, container any, key any, msg ...any) {
	t.Helper()

	// Default message
	message := fmt.Sprintf("expected container to contain %v", key)
	if len(msg) > 0 {
		message = fmt.Sprint(msg...)
	}

	switch c := container.(type) {
	case string:
		substr, ok := key.(string)
		if !ok {
			t.Fatalf("key must be a string when container is a string")
		}
		if !strings.Contains(c, substr) {
			t.Fatal(message)
		}

	case map[string]any:
		if _, ok := c[key.(string)]; !ok {
			t.Fatal(message)
		}

	default:
		// Generic map type using reflection (works with map[K]V of any comparable K)
		val := reflect.ValueOf(container)
		if val.Kind() == reflect.Map {
			keyVal := reflect.ValueOf(key)
			if !val.MapIndex(keyVal).IsValid() {
				t.Fatal(message)
			}
		} else {
			t.Fatalf("Contains only supports string or map types, got %T", container)
		}
	}
}

// EqualError asserts that err is not nil and its Error() matches expected.
func EqualError(t *testing.T, expected, err error, msg ...any) {
	t.Helper()

	if err == nil {
		message := fmt.Sprintf("expected error %q, got nil", expected.Error())
		if len(msg) > 0 {
			message = fmt.Sprint(msg...)
		}
		t.Fatal(message)
	}

	if expected.Error() != err.Error() {
		message := fmt.Sprintf("expected error %q, got %q", expected.Error(), err.Error())
		if len(msg) > 0 {
			message = fmt.Sprint(msg...)
		}
		t.Fatal(message)
	}
}
