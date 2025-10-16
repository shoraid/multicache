package multicache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shoraid/multicache/contract"
	multicachemock "github.com/shoraid/multicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestHelper_toBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    bool
		expectedErr error
	}{
		// True
		{
			name:        "from bool true",
			input:       true,
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "from string 'true'",
			input:       "true",
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "from []byte 'true'",
			input:       []byte("true"),
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "from string '1'",
			input:       "1",
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "from int 1",
			input:       1,
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "from int64 1",
			input:       int64(1),
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "from float64 1.0",
			input:       1.0,
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "from float32 1.0",
			input:       float32(1.0),
			expected:    true,
			expectedErr: nil,
		},

		// False
		{
			name:        "from bool false",
			input:       false,
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "from string 'false'",
			input:       "false",
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "from []byte 'false'",
			input:       []byte("false"),
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "from string '0'",
			input:       "0",
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "from int 0",
			input:       0,
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "from int64 0",
			input:       int64(0),
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "from float64 0.0",
			input:       0.0,
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "from float32 0.0",
			input:       float32(0.0),
			expected:    false,
			expectedErr: nil,
		},

		// Invalid
		{
			name:        "from string invalid",
			input:       "not a bool",
			expected:    false,
			expectedErr: errors.New("strconv.ParseBool: parsing \"not a bool\": invalid syntax"),
		},
		{
			name:        "from []byte invalid",
			input:       []byte("not a bool"),
			expected:    false,
			expectedErr: errors.New("strconv.ParseBool: parsing \"not a bool\": invalid syntax"),
		},
		{
			name:        "from unsupported type",
			input:       struct{}{},
			expected:    false,
			expectedErr: errors.New("convert: cannot cast struct {} to bool"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := toBool(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHelper_toInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    int
		expectedErr error
	}{
		// Valid conversions
		{
			name:        "from string valid int",
			input:       "456",
			expected:    456,
			expectedErr: nil,
		},
		{
			name:        "from []byte valid int",
			input:       []byte("789"),
			expected:    789,
			expectedErr: nil,
		},
		{
			name:        "from int",
			input:       123,
			expected:    123,
			expectedErr: nil,
		},
		{
			name:        "from int8",
			input:       int8(12),
			expected:    12,
			expectedErr: nil,
		},
		{
			name:        "from int16",
			input:       int16(1234),
			expected:    1234,
			expectedErr: nil,
		},
		{
			name:        "from int32",
			input:       int32(123456),
			expected:    123456,
			expectedErr: nil,
		},
		{
			name:        "from int64",
			input:       int64(1234567890),
			expected:    1234567890,
			expectedErr: nil,
		},
		{
			name:        "from uint",
			input:       uint(123),
			expected:    123,
			expectedErr: nil,
		},
		{
			name:        "from uint8",
			input:       uint8(12),
			expected:    12,
			expectedErr: nil,
		},
		{
			name:        "from uint16",
			input:       uint16(1234),
			expected:    1234,
			expectedErr: nil,
		},
		{
			name:        "from uint32",
			input:       uint32(123456),
			expected:    123456,
			expectedErr: nil,
		},
		{
			name:        "from uint64",
			input:       uint64(1234567890),
			expected:    1234567890,
			expectedErr: nil,
		},
		{
			name:        "from float32",
			input:       float32(123.45),
			expected:    123,
			expectedErr: nil,
		},
		{
			name:        "from float64",
			input:       123.45,
			expected:    123,
			expectedErr: nil,
		},
		{
			name:        "from bool true",
			input:       true,
			expected:    1,
			expectedErr: nil,
		},
		{
			name:        "from bool false",
			input:       false,
			expected:    0,
			expectedErr: nil,
		},

		// Invalid conversions
		{
			name:        "from string invalid int",
			input:       "abc",
			expected:    0,
			expectedErr: errors.New("convert: failed to parse int from string: strconv.Atoi: parsing \"abc\": invalid syntax"),
		},
		{
			name:        "from []byte invalid int",
			input:       []byte("xyz"),
			expected:    0,
			expectedErr: errors.New("convert: failed to parse int from bytes: strconv.Atoi: parsing \"xyz\": invalid syntax"),
		},
		{
			name:        "from unsupported type",
			input:       struct{}{},
			expected:    0,
			expectedErr: errors.New("convert: cannot cast struct {} to int"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := toInt(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHelper_toInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    int64
		expectedErr error
	}{
		// Valid conversions
		{
			name:        "from string valid int64",
			input:       "9876543210",
			expected:    9876543210,
			expectedErr: nil,
		},
		{
			name:        "from []byte valid int64",
			input:       []byte("1234567890"),
			expected:    1234567890,
			expectedErr: nil,
		},
		{
			name:        "from int",
			input:       123,
			expected:    123,
			expectedErr: nil,
		},
		{
			name:        "from int8",
			input:       int8(12),
			expected:    12,
			expectedErr: nil,
		},
		{
			name:        "from int16",
			input:       int16(1234),
			expected:    1234,
			expectedErr: nil,
		},
		{
			name:        "from int32",
			input:       int32(123456),
			expected:    123456,
			expectedErr: nil,
		},
		{
			name:        "from int64",
			input:       int64(987654321098765),
			expected:    987654321098765,
			expectedErr: nil,
		},
		{
			name:        "from uint",
			input:       uint(123),
			expected:    123,
			expectedErr: nil,
		},
		{
			name:        "from uint8",
			input:       uint8(12),
			expected:    12,
			expectedErr: nil,
		},
		{
			name:        "from uint16",
			input:       uint16(1234),
			expected:    1234,
			expectedErr: nil,
		},
		{
			name:        "from uint32",
			input:       uint32(123456),
			expected:    123456,
			expectedErr: nil,
		},
		{
			name:        "from uint64",
			input:       uint64(987654321098765),
			expected:    987654321098765,
			expectedErr: nil,
		},
		{
			name:        "from float32",
			input:       float32(123.45),
			expected:    123,
			expectedErr: nil,
		},
		{
			name:        "from float64",
			input:       123.45,
			expected:    123,
			expectedErr: nil,
		},
		{
			name:        "from bool true",
			input:       true,
			expected:    1,
			expectedErr: nil,
		},
		{
			name:        "from bool false",
			input:       false,
			expected:    0,
			expectedErr: nil,
		},

		// Invalid conversions
		{
			name:        "from string invalid int64",
			input:       "abc",
			expected:    0,
			expectedErr: errors.New("convert: failed to parse int64 from string: strconv.ParseInt: parsing \"abc\": invalid syntax"),
		},
		{
			name:        "from []byte invalid int64",
			input:       []byte("xyz"),
			expected:    0,
			expectedErr: errors.New("convert: failed to parse int64 from bytes: strconv.ParseInt: parsing \"xyz\": invalid syntax"),
		},
		{
			name:        "from unsupported type",
			input:       struct{}{},
			expected:    0,
			expectedErr: errors.New("convert: cannot cast struct {} to int64"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := toInt64(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHelper_toInts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    []int
		expectedErr error
	}{
		{
			name:        "from []int",
			input:       []int{1, 2, 3},
			expected:    []int{1, 2, 3},
			expectedErr: nil,
		},
		{
			name:        "from JSON string",
			input:       "[4,5,6]",
			expected:    []int{4, 5, 6},
			expectedErr: nil,
		},
		{
			name:        "from JSON []byte",
			input:       []byte("[7,8,9]"),
			expected:    []int{7, 8, 9},
			expectedErr: nil,
		},
		{
			name:        "from []float64",
			input:       []float64{1.1, 2.9, 3.0},
			expected:    []int{1, 2, 3},
			expectedErr: nil,
		},
		{
			name:        "from []any with floats and ints",
			input:       []any{1.1, 2, 3.9},
			expected:    []int{1, 2, 3},
			expectedErr: nil,
		},
		{
			name:        "from empty []int",
			input:       []int{},
			expected:    []int{},
			expectedErr: nil,
		},
		{
			name:        "from empty JSON string",
			input:       "[]",
			expected:    []int{},
			expectedErr: nil,
		},
		{
			name:        "from empty JSON []byte",
			input:       []byte("[]"),
			expected:    []int{},
			expectedErr: nil,
		},
		{
			name:        "from invalid JSON string",
			input:       "invalid json",
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "from invalid JSON []byte",
			input:       []byte("invalid json"),
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "from []any with unsupported type",
			input:       []any{1, "two", 3},
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "from unsupported type",
			input:       "not an array",
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "from nil",
			input:       nil,
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := toInts(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHelper_toString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    string
		expectedErr error
	}{
		{
			name:        "from string",
			input:       "hello",
			expected:    "hello",
			expectedErr: nil,
		},
		{
			name:        "from []byte",
			input:       []byte("world"),
			expected:    "world",
			expectedErr: nil,
		},

		// Int
		{
			name:        "from int",
			input:       123,
			expected:    "123",
			expectedErr: nil,
		},
		{
			name:        "from int8",
			input:       int8(12),
			expected:    "12",
			expectedErr: nil,
		},
		{
			name:        "from int16",
			input:       int16(123),
			expected:    "123",
			expectedErr: nil,
		},
		{
			name:        "from int32",
			input:       int32(12345),
			expected:    "12345",
			expectedErr: nil,
		},
		{
			name:        "from int64",
			input:       int64(4567890123),
			expected:    "4567890123",
			expectedErr: nil,
		},

		// Uint
		{
			name:        "from uint",
			input:       uint(123),
			expected:    "123",
			expectedErr: nil,
		},
		{
			name:        "from uint8",
			input:       uint8(12),
			expected:    "12",
			expectedErr: nil,
		},
		{
			name:        "from uint16",
			input:       uint16(123),
			expected:    "123",
			expectedErr: nil,
		},
		{
			name:        "from uint32",
			input:       uint32(12345),
			expected:    "12345",
			expectedErr: nil,
		},
		{
			name:        "from uint64",
			input:       uint64(4567890123),
			expected:    "4567890123",
			expectedErr: nil,
		},

		// Float
		{
			name:        "from float32",
			input:       float32(1.23),
			expected:    "1.23",
			expectedErr: nil,
		},
		{
			name:        "from float64",
			input:       4.567,
			expected:    "4.567",
			expectedErr: nil,
		},

		// Boolean
		{
			name:        "from bool true",
			input:       true,
			expected:    "true",
			expectedErr: nil,
		},
		{
			name:        "from bool false",
			input:       false,
			expected:    "false",
			expectedErr: nil,
		},

		// Struct
		{
			name:        "from struct (JSON marshalable)",
			input:       struct{ Name string }{Name: "Test"},
			expected:    `{"Name":"Test"}`,
			expectedErr: nil,
		},
		{
			name: "from complex struct (JSON marshalable)",
			input: struct {
				A int
				B []string
			}{A: 1, B: []string{"x", "y"}},
			expected:    `{"A":1,"B":["x","y"]}`,
			expectedErr: nil,
		},

		// Invalid
		{
			name:        "from unmarshalable type", // e.g., channel
			input:       make(chan int),
			expected:    "",
			expectedErr: errors.New("convert: cannot cast chan int to string and failed to marshal to JSON: json: unsupported type: chan int"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := toString(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHelper_toStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    []string
		expectedErr error
	}{
		{
			name:        "from []string",
			input:       []string{"a", "b", "c"},
			expected:    []string{"a", "b", "c"},
			expectedErr: nil,
		},
		{
			name:        "from JSON string",
			input:       `["d","e","f"]`,
			expected:    []string{"d", "e", "f"},
			expectedErr: nil,
		},
		{
			name:        "from JSON []byte",
			input:       []byte(`["g","h","i"]`),
			expected:    []string{"g", "h", "i"},
			expectedErr: nil,
		},
		{
			name:        "from empty []string",
			input:       []string{},
			expected:    []string{},
			expectedErr: nil,
		},
		{
			name:        "from empty JSON string",
			input:       "[]",
			expected:    []string{},
			expectedErr: nil,
		},
		{
			name:        "from empty JSON []byte",
			input:       []byte("[]"),
			expected:    []string{},
			expectedErr: nil,
		},
		{
			name:        "from invalid JSON string",
			input:       "invalid json",
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "from invalid JSON []byte",
			input:       []byte("invalid json"),
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "from unsupported type",
			input:       123,
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "from nil",
			input:       nil,
			expected:    nil,
			expectedErr: nil,
		},
		{
			name:        "from []any of strings (json marshal fallback)",
			input:       []any{"str1", "str2"},
			expected:    []string{"str1", "str2"},
			expectedErr: nil,
		},
		{
			name:        "from []any of mixed types (json marshal fallback, should fail)",
			input:       []any{"str1", 123},
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
		{
			name: "from non-marshallable type (should fail marshal)",
			input: func() any {
				ch := make(chan int)
				return ch
			}(),
			expected:    nil,
			expectedErr: ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := toStrings(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHelper_getOrSetDefault(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Value string
	}

	ctx := context.Background()
	key := "test-key"
	ttl := time.Minute

	defaultVal := TestStruct{Value: "default"}
	defaultFn := func() (TestStruct, error) {
		return defaultVal, nil
	}
	errorFn := func() (TestStruct, error) {
		return TestStruct{}, errors.New("default function error")
	}

	tests := []struct {
		name          string
		defaultFunc   func() (TestStruct, error)
		setErr        error
		expectedValue TestStruct
		expectedErr   error
	}{
		{
			name:          "should return default value and set it if no error",
			defaultFunc:   defaultFn,
			setErr:        nil,
			expectedValue: defaultVal,
			expectedErr:   nil,
		},
		{
			name:          "should return default value and set error if set fails",
			defaultFunc:   defaultFn,
			setErr:        errors.New("set failed"),
			expectedValue: defaultVal,
			expectedErr:   errors.New("set failed"),
		},
		{
			name:          "should return error if default function fails",
			defaultFunc:   errorFn,
			setErr:        nil, // Set should not be called
			expectedValue: TestStruct{},
			expectedErr:   errors.New("default function error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(multicachemock.MockStore)
			manager := &managerImpl{
				stores: map[string]contract.Store{"default": mockStore},
				store:  mockStore,
			}

			if tt.expectedErr == nil || tt.expectedErr.Error() == "set failed" {
				mockStore.
					On("Set", ctx, key, defaultVal, ttl).
					Return(tt.setErr).
					Maybe() // Use Maybe because Set might not be called if defaultFn fails
			}

			value, err := getOrSetDefault(ctx, manager, key, ttl, tt.defaultFunc)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Equal(t, tt.expectedValue, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}

			mockStore.AssertExpectations(t)
		})
	}
}
