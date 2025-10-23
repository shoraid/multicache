package omnicache

import (
	"testing"

	"github.com/shoraid/omnicache/internal/assert"
)

func TestHelper_convertAnyToType(t *testing.T) {
	t.Parallel()

	type dummy struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		input     any
		expected  any
		expectErr bool
	}{
		// --- Direct type match ---
		{name: "direct string", input: "hello", expected: "hello", expectErr: false},
		{name: "direct int", input: 123, expected: 123, expectErr: false},
		{name: "direct bool", input: true, expected: true, expectErr: false},
		{name: "direct []byte", input: []byte("bytes"), expected: []byte("bytes"), expectErr: false},
		{name: "direct struct", input: dummy{Name: "Test", Age: 1}, expected: dummy{Name: "Test", Age: 1}, expectErr: false},

		// --- From string conversions ---
		{name: "string to int", input: "123", expected: 123, expectErr: false},
		{name: "string to int64", input: "9876543210", expected: int64(9876543210), expectErr: false},
		{name: "string to float64", input: "123.45", expected: 123.45, expectErr: false},
		{name: "string to bool (true)", input: "true", expected: true, expectErr: false},
		{name: "string to bool (false)", input: "false", expected: false, expectErr: false},
		{name: "string to []byte", input: "byte_string", expected: []byte("byte_string"), expectErr: false},
		{name: "string to struct (JSON)", input: `{"Name":"Test","Age":30}`, expected: dummy{Name: "Test", Age: 30}, expectErr: false},
		{name: "string to []string (JSON)", input: `["a","b"]`, expected: []string{"a", "b"}, expectErr: false},

		// --- From []byte conversions ---
		{name: "[]byte to string", input: []byte("hello"), expected: "hello", expectErr: false},
		{name: "[]byte to int", input: []byte("123"), expected: 123, expectErr: false},
		{name: "[]byte to int64", input: []byte("9876543210"), expected: int64(9876543210), expectErr: false},
		{name: "[]byte to float64", input: []byte("123.45"), expected: 123.45, expectErr: false},
		{name: "[]byte to bool (true)", input: []byte("true"), expected: true, expectErr: false},
		{name: "[]byte to bool (false)", input: []byte("false"), expected: false, expectErr: false},
		{name: "[]byte to struct (JSON)", input: []byte(`{"Name":"Test","Age":30}`), expected: dummy{Name: "Test", Age: 30}, expectErr: false},
		{name: "[]byte to []string (JSON)", input: []byte(`["a","b"]`), expected: []string{"a", "b"}, expectErr: false},

		// --- From number conversions ---
		{name: "int to string", input: 123, expected: "123", expectErr: false},
		{name: "int to float64", input: 123, expected: 123.0, expectErr: false},
		{name: "int64 to string", input: int64(9876543210), expected: "9876543210", expectErr: false},
		{name: "float64 to string", input: 123.45, expected: "123.45", expectErr: false},
		{name: "bool to string", input: true, expected: "true", expectErr: false},

		// --- JSON fallback conversions (for types not directly handled) ---
		{name: "struct to string (JSON)", input: dummy{Name: "Test", Age: 30}, expected: `{"Name":"Test","Age":30}`, expectErr: false},
		{name: "[]int to string (JSON)", input: []int{1, 2, 3}, expected: `[1,2,3]`, expectErr: false},

		// --- Error cases ---
		{name: "nil input", input: nil, expected: nil, expectErr: true},
		{name: "unsupported conversion (string to struct with bad JSON)", input: `{"Name":"Test",Age:30}`, expected: dummy{}, expectErr: true},
		{name: "unsupported conversion (int to struct)", input: 123, expected: dummy{}, expectErr: true},
		{name: "unsupported conversion (string to int with bad string)", input: "abc", expected: 0, expectErr: true},
		{name: "unsupported conversion ([]byte to int with bad bytes)", input: []byte("abc"), expected: 0, expectErr: true},

		// --- Force marshal error branch ---
		{name: "marshal error (unsupported type like channel)", input: make(chan int), expected: "", expectErr: true},
		{name: "string to struct via JSON with invalid syntax", input: `{"Name":"Test",Age:30}`, expected: dummy{}, expectErr: true},
		{name: "string to []string via JSON with invalid syntax", input: `["a","b"`, expected: []string{}, expectErr: true},
		{
			name:      "struct to map[string]any via JSON round-trip",
			input:     dummy{Name: "Test", Age: 30},
			expected:  map[string]any{"Name": "Test", "Age": float64(30)},
			expectErr: false,
		},
		{
			name:      "map[string]any to struct via JSON round-trip",
			input:     map[string]any{"Name": "Test", "Age": 30},
			expected:  dummy{Name: "Test", Age: 30},
			expectErr: false,
		},
		{
			name:      "fallback round-trip marshal error (unsupported type func)",
			input:     struct{ F func() }{F: func() {}},
			expected:  dummy{}, // zero value of target type
			expectErr: true,
		},
		{
			name:      "fallback round-trip unmarshal error (type mismatch)",
			input:     map[string]any{"Name": "Test", "Age": "not-a-number"}, // valid JSON but invalid for dummy.Age (expects int)
			expected:  dummy{},                                               // zero value for T
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Act ---
			var result any
			var err error

			switch tt.expected.(type) {
			case string:
				result, err = convertAnyToType[string](tt.input)
			case int:
				result, err = convertAnyToType[int](tt.input)
			case int64:
				result, err = convertAnyToType[int64](tt.input)
			case float64:
				result, err = convertAnyToType[float64](tt.input)
			case bool:
				result, err = convertAnyToType[bool](tt.input)
			case []byte:
				result, err = convertAnyToType[[]byte](tt.input)
			case dummy:
				result, err = convertAnyToType[dummy](tt.input)
			case []string:
				result, err = convertAnyToType[[]string](tt.input)
			case []int:
				result, err = convertAnyToType[[]int](tt.input)
			case nil:
				result, err = convertAnyToType[any](tt.input)
			case map[string]any:
				result, err = convertAnyToType[map[string]any](tt.input)
			default:
				t.Fatalf("unsupported type in test: %T", tt.expected)
			}

			if tt.expectErr {
				assert.Error(t, err, "must return an error when conversion fails")
				assert.Equal(t, tt.expected, result, "result on error must be the zero value of the target type")
				return
			}

			assert.NoError(t, err, "must not return an error when conversion succeeds")
			assert.Equal(t, tt.expected, result, "converted value must match the expected value")
		})
	}
}

func TestHelper_fromString(t *testing.T) {
	t.Parallel()

	type dummy struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		input     string
		expected  any
		expectErr bool
	}{
		// Valid
		{name: "to string", input: "hello", expected: "hello", expectErr: false},
		{name: "to int", input: "123", expected: 123, expectErr: false},
		{name: "to int64", input: "9876543210", expected: int64(9876543210), expectErr: false},
		{name: "to float64", input: "123.45", expected: 123.45, expectErr: false},
		{name: "to bool true", input: "true", expected: true, expectErr: false},
		{name: "1 to bool true", input: "1", expected: true, expectErr: false},
		{name: "to bool false", input: "false", expected: false, expectErr: false},
		{name: "0 to bool false", input: "0", expected: false, expectErr: false},
		{name: "to []byte", input: "byte_string", expected: []byte("byte_string"), expectErr: false},
		{name: "to []string from JSON", input: `["a", "b"]`, expected: []string{"a", "b"}, expectErr: false},
		{name: "to []int from JSON", input: `[1, 2]`, expected: []int{1, 2}, expectErr: false},
		{name: "to struct from JSON", input: `{"Name":"Test","Age":30}`, expected: dummy{Name: "Test", Age: 30}, expectErr: false},
		{name: "to uint", input: "123", expected: uint(123), expectErr: false},
		{name: "to uint64", input: "9876543210", expected: uint64(9876543210), expectErr: false},

		// Invalid
		{name: "invalid string to int", input: "abc", expected: 0, expectErr: true},
		{name: "invalid string to int64", input: "def", expected: int64(0), expectErr: true},
		{name: "invalid string to float64", input: "xyz", expected: 0.0, expectErr: true},
		{name: "invalid string to bool", input: "not_a_bool", expected: false, expectErr: true},
		{name: "invalid string to uint", input: "-123", expected: uint(0), expectErr: true},
		{name: "invalid string to uint64", input: "-9876543210", expected: uint64(0), expectErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result any
			var err error

			switch tt.expected.(type) {
			case string:
				result, err = fromString[string](tt.input)
			case int:
				result, err = fromString[int](tt.input)
			case int64:
				result, err = fromString[int64](tt.input)
			case float64:
				result, err = fromString[float64](tt.input)
			case bool:
				result, err = fromString[bool](tt.input)
			case []byte:
				result, err = fromString[[]byte](tt.input)
			case []string:
				result, err = fromString[[]string](tt.input)
			case []int:
				result, err = fromString[[]int](tt.input)
			case dummy:
				result, err = fromString[dummy](tt.input)
			case uint:
				result, err = fromString[uint](tt.input)
			case uint64:
				result, err = fromString[uint64](tt.input)
			default:
				t.Fatalf("unsupported type in test: %T", tt.expected)
			}

			if tt.expectErr {
				assert.Error(t, err, "must return an error when conversion fails")
				assert.Equal(t, tt.expected, result, "result on error must be the zero value of the target type")
				return
			}

			assert.NoError(t, err, "must not return an error when conversion succeeds")
			assert.Equal(t, tt.expected, result, "converted value must match the expected value")
		})
	}
}

func TestHelper_fromBytes(t *testing.T) {
	t.Parallel()

	type dummy struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		input     []byte
		expected  any
		expectErr bool
	}{
		// Valid
		{name: "to string", input: []byte("hello"), expected: "hello", expectErr: false},
		{name: "to int", input: []byte("123"), expected: 123, expectErr: false},
		{name: "to int64", input: []byte("9876543210"), expected: int64(9876543210), expectErr: false},
		{name: "to float64", input: []byte("123.45"), expected: 123.45, expectErr: false},
		{name: "to bool true", input: []byte("true"), expected: true, expectErr: false},
		{name: "1 to bool true", input: []byte("1"), expected: true, expectErr: false},
		{name: "to bool false", input: []byte("false"), expected: false, expectErr: false},
		{name: "0 to bool false", input: []byte("0"), expected: false, expectErr: false},
		{name: "to []string from JSON", input: []byte(`["a", "b"]`), expected: []string{"a", "b"}, expectErr: false},
		{name: "to []int from JSON", input: []byte(`[1, 2]`), expected: []int{1, 2}, expectErr: false},
		{name: "to struct from JSON", input: []byte(`{"Name":"Test","Age":30}`), expected: dummy{Name: "Test", Age: 30}, expectErr: false},
		{name: "to uint", input: []byte("123"), expected: uint(123), expectErr: false},
		{name: "to uint64", input: []byte("9876543210"), expected: uint64(9876543210), expectErr: false},

		// Invalid
		{name: "invalid bytes to int", input: []byte("abc"), expected: 0, expectErr: true},
		{name: "invalid bytes to int64", input: []byte("def"), expected: int64(0), expectErr: true},
		{name: "invalid bytes to float64", input: []byte("xyz"), expected: 0.0, expectErr: true},
		{name: "invalid bytes to bool", input: []byte("not_a_bool"), expected: false, expectErr: true},
		{name: "invalid string to uint", input: []byte("-123"), expected: uint(0), expectErr: true},
		{name: "invalid string to uint64", input: []byte("-9876543210"), expected: uint64(0), expectErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result any
			var err error

			switch tt.expected.(type) {
			case string:
				result, err = fromBytes[string](tt.input)
			case int:
				result, err = fromBytes[int](tt.input)
			case int64:
				result, err = fromBytes[int64](tt.input)
			case float64:
				result, err = fromBytes[float64](tt.input)
			case bool:
				result, err = fromBytes[bool](tt.input)
			case []string:
				result, err = fromBytes[[]string](tt.input)
			case []int:
				result, err = fromBytes[[]int](tt.input)
			case dummy:
				result, err = fromBytes[dummy](tt.input)
			case uint:
				result, err = fromBytes[uint](tt.input)
			case uint64:
				result, err = fromBytes[uint64](tt.input)
			default:
				t.Fatalf("unsupported type in test: %T", tt.expected)
			}

			if tt.expectErr {
				assert.Error(t, err, "must return an error when conversion fails")
				assert.Equal(t, tt.expected, result, "result on error must be the zero value of the target type")
				return
			}

			assert.NoError(t, err, "must not return an error when conversion succeeds")
			assert.Equal(t, tt.expected, result, "converted value must match the expected value")
		})
	}
}

func TestHelper_fromStringOrNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{name: "to string", input: "hello", expected: "hello"},
		{name: "to int", input: "123", expected: 123},
		{name: "to int64", input: "9876543210", expected: int64(9876543210)},
		{name: "to float64", input: "123.45", expected: 123.45},
		{name: "to bool true", input: "true", expected: true},
		{name: "to bool false", input: "false", expected: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result any
			var err error

			switch tt.expected.(type) {
			case string:
				result, err = fromStringOrNumber[string](tt.input)
			case int:
				result, err = fromStringOrNumber[int](tt.input)
			case int64:
				result, err = fromStringOrNumber[int64](tt.input)
			case float64:
				result, err = fromStringOrNumber[float64](tt.input)
			case bool:
				result, err = fromStringOrNumber[bool](tt.input)
			default:
				t.Fatalf("unsupported type in test: %T", tt.expected)
			}

			assert.NoError(t, err, "must not return an error when conversion succeeds")
			assert.Equal(t, tt.expected, result, "converted value must match the expected value")
		})
	}
}
