package multicache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/shoraid/multicache/contract"
	multicachemock "github.com/shoraid/multicache/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericManager_Get(t *testing.T) {
	t.Parallel()

	type SampleStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	ctx := context.Background()
	key := "test-key"
	sampleStruct := SampleStruct{ID: 1, Name: "Alice"}

	// Convert struct to JSON for []byte and string cases
	jsonBytes, _ := json.Marshal(sampleStruct)
	jsonString := string(jsonBytes)

	tests := []struct {
		name        string
		mockVal     any
		mockErr     error
		expectedVal SampleStruct
		expectedErr error
	}{
		{
			name:        "should return struct value directly if type matches",
			mockVal:     sampleStruct,
			mockErr:     nil,
			expectedVal: sampleStruct,
			expectedErr: nil,
		},
		{
			name:        "should return error if underlying store returns cache miss",
			mockVal:     nil,
			mockErr:     ErrCacheMiss,
			expectedVal: SampleStruct{},
			expectedErr: ErrCacheMiss,
		},
		{
			name:        "should convert from JSON string to struct",
			mockVal:     jsonString,
			mockErr:     nil,
			expectedVal: sampleStruct,
			expectedErr: nil,
		},
		{
			name:        "should convert from JSON bytes to struct",
			mockVal:     jsonBytes,
			mockErr:     nil,
			expectedVal: sampleStruct,
			expectedErr: nil,
		},
		{
			name:        "should return error if store returns random error",
			mockVal:     nil,
			mockErr:     errors.New("store error"),
			expectedVal: SampleStruct{},
			expectedErr: errors.New("store error"),
		},
		{
			name:        "should return error if type is not convertible",
			mockVal:     12345, // int cannot be converted to SampleStruct
			mockErr:     nil,
			expectedVal: SampleStruct{},
			expectedErr: errors.New("convert: cannot unmarshal fallback to T"), // prefix check
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

			// Mock Get call
			mockStore.
				On("Get", ctx, key).
				Return(tt.mockVal, tt.mockErr).
				Once()

			g := G[SampleStruct](manager)
			val, err := g.Get(ctx, key)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Contains(t, err.Error(), tt.expectedErr.Error(), "expected error message match")
				assert.Equal(t, SampleStruct{}, val, "expected zero value on error")
			} else {
				require.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedVal, val, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestGenericManager_GetOrSet(t *testing.T) {
	t.Parallel()

	type DummyStruct struct {
		ID   int
		Name string
	}

	ctx := context.Background()
	key := "test-key"
	ttl := 5 * time.Minute

	existingVal := DummyStruct{ID: 1, Name: "Existing"}
	defaultVal := DummyStruct{ID: 2, Name: "Default"}

	defaultFn := func() (DummyStruct, error) {
		return defaultVal, nil
	}

	errorFn := func() (DummyStruct, error) {
		return DummyStruct{}, errors.New("default function error")
	}

	tests := []struct {
		name            string
		getVal          any
		getErr          error
		defaultFn       func() (DummyStruct, error)
		setErr          error
		expectedVal     DummyStruct
		expectedErr     error
		expectSetCalled bool
	}{
		{
			name:            "should return existing value if found in cache",
			getVal:          existingVal,
			getErr:          nil,
			defaultFn:       defaultFn,
			setErr:          nil,
			expectedVal:     existingVal,
			expectedErr:     nil,
			expectSetCalled: false,
		},
		{
			name:            "should compute default and set when cache miss occurs",
			getVal:          nil,
			getErr:          ErrCacheMiss,
			defaultFn:       defaultFn,
			setErr:          nil,
			expectedVal:     defaultVal,
			expectedErr:     nil,
			expectSetCalled: true,
		},
		{
			name:            "should compute default and set when type mismatch occurs",
			getVal:          "not a struct",
			getErr:          ErrTypeMismatch,
			defaultFn:       defaultFn,
			setErr:          nil,
			expectedVal:     defaultVal,
			expectedErr:     nil,
			expectSetCalled: true,
		},
		{
			name:            "should return error if default function fails",
			getVal:          nil,
			getErr:          ErrCacheMiss,
			defaultFn:       errorFn,
			setErr:          nil,
			expectedVal:     DummyStruct{},
			expectedErr:     errors.New("default function error"),
			expectSetCalled: false,
		},
		{
			name:            "should return value and set error if set fails after cache miss",
			getVal:          nil,
			getErr:          ErrCacheMiss,
			defaultFn:       defaultFn,
			setErr:          errors.New("set failed"),
			expectedVal:     defaultVal,
			expectedErr:     errors.New("set failed"),
			expectSetCalled: true,
		},
		{
			name:            "should return error if get returns unexpected error",
			getVal:          nil,
			getErr:          errors.New("network error"),
			defaultFn:       defaultFn,
			setErr:          nil,
			expectedVal:     DummyStruct{},
			expectedErr:     errors.New("network error"),
			expectSetCalled: false,
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

			// Mock Get call
			mockStore.
				On("Get", ctx, key).
				Return(tt.getVal, tt.getErr).
				Once()

			// Mock Set call if expected
			if tt.expectSetCalled {
				mockStore.
					On("Set", ctx, key, defaultVal, ttl).
					Return(tt.setErr).
					Once()
			}

			g := G[DummyStruct](manager)
			val, err := g.GetOrSet(ctx, key, ttl, tt.defaultFn)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Contains(t, err.Error(), tt.expectedErr.Error(), "expected error message")
				assert.Equal(t, tt.expectedVal, val, "expected returned value")
			} else {
				require.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedVal, val, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestGenericManager_convertAnyToType(t *testing.T) {
	t.Parallel()

	type convertTestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	structVal := convertTestStruct{ID: 1, Name: "Alice"}
	structJSON, _ := json.Marshal(structVal)

	tests := []struct {
		name        string
		input       any
		expectedVal any
		expectedErr string
	}{
		// Direct type assertion
		{
			name:        "should return direct value if type matches",
			input:       "hello",
			expectedVal: "hello",
			expectedErr: "",
		},

		// From []byte
		{
			name:        "should parse []byte to string successfully",
			input:       []byte("hello"),
			expectedVal: "hello",
			expectedErr: "",
		},
		{
			name:        "should parse []byte to int successfully",
			input:       []byte("123"),
			expectedVal: 123,
			expectedErr: "",
		},
		{
			name:        "should return error when parsing []byte to int fails",
			input:       []byte("abc"),
			expectedVal: 0,
			expectedErr: `convert: failed to parse int from bytes: strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			name:        "should parse []byte to bool successfully",
			input:       []byte("true"),
			expectedVal: true,
			expectedErr: "",
		},
		{
			name:        "should unmarshal []byte to struct successfully",
			input:       structJSON,
			expectedVal: structVal,
			expectedErr: "",
		},
		{
			name:        "should return error for invalid struct JSON []byte",
			input:       []byte("{invalid_json}"),
			expectedVal: convertTestStruct{},
			expectedErr: "convert: failed to unmarshal bytes to T: invalid character 'i' looking for beginning of object key string",
		},

		// From string
		{
			name:        "should parse string to int successfully",
			input:       "456",
			expectedVal: 456,
			expectedErr: "",
		},
		{
			name:        "should parse string to bool successfully",
			input:       "false",
			expectedVal: false,
			expectedErr: "",
		},
		{
			name:        "should unmarshal string to struct successfully",
			input:       string(structJSON),
			expectedVal: structVal,
			expectedErr: "",
		},
		{
			name:        "should return error for invalid struct string",
			input:       "{invalid_json}",
			expectedVal: convertTestStruct{},
			expectedErr: "convert: failed to unmarshal string to T: invalid character 'i' looking for beginning of object key string",
		},

		// Fallback JSON marshal/unmarshal
		{
			name:        "should fallback marshal/unmarshal map[string]any to struct successfully",
			input:       map[string]any{"id": 1, "name": "Alice"},
			expectedVal: structVal,
			expectedErr: "",
		},
		{
			name:        "should return error if fallback marshal fails",
			input:       make(chan int), // json.Marshal will fail on channel
			expectedVal: convertTestStruct{},
			expectedErr: "convert: cannot marshal fallback (chan int): json: unsupported type: chan int",
		},
		{
			name:        "should return error if fallback unmarshal fails",
			input:       map[string]any{"id": make(chan int)}, // Marshal passes, but unmarshal to struct will fail
			expectedVal: convertTestStruct{},
			expectedErr: "convert: cannot marshal fallback (map[string]interface {}): json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch expected := tt.expectedVal.(type) {
			case string:
				val, err := convertAnyToType[string](tt.input)
				if tt.expectedErr != "" {
					assert.EqualError(t, err, tt.expectedErr, "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case int:
				val, err := convertAnyToType[int](tt.input)
				if tt.expectedErr != "" {
					assert.EqualError(t, err, tt.expectedErr, "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case bool:
				val, err := convertAnyToType[bool](tt.input)
				if tt.expectedErr != "" {
					assert.EqualError(t, err, tt.expectedErr, "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case convertTestStruct:
				val, err := convertAnyToType[convertTestStruct](tt.input)
				if tt.expectedErr != "" {
					assert.EqualError(t, err, tt.expectedErr, "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			default:
				t.Fatalf("unhandled expected type %T", expected)
			}
		})
	}
}

func TestGenericManager_parseFromBytes(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	structValue := testStruct{ID: 10, Name: "Alice"}
	structJSON, _ := json.Marshal(structValue)

	tests := []struct {
		name        string
		input       []byte
		expectedVal any
		expectedErr error
	}{
		// String
		{
			name:        "should parse bytes to string successfully",
			input:       []byte("hello"),
			expectedVal: "hello",
			expectedErr: nil,
		},

		// Int
		{
			name:        "should parse bytes to int successfully",
			input:       []byte("123"),
			expectedVal: 123,
			expectedErr: nil,
		},
		{
			name:        "should return error for invalid int bytes",
			input:       []byte("abc"),
			expectedVal: 0,
			expectedErr: errors.New(`convert: failed to parse int from bytes: strconv.Atoi: parsing "abc": invalid syntax`),
		},

		// Int64
		{
			name:        "should parse bytes to int64 successfully",
			input:       []byte("456"),
			expectedVal: int64(456),
			expectedErr: nil,
		},
		{
			name:        "should return error for invalid int64 bytes",
			input:       []byte("notanumber"),
			expectedVal: int64(0),
			expectedErr: errors.New(`convert: failed to parse int64 from bytes: strconv.ParseInt: parsing "notanumber": invalid syntax`),
		},

		// Boolean - true variants
		{
			name:        `should parse "true" bytes to bool true`,
			input:       []byte("true"),
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        `should parse "TRUE" bytes to bool true`,
			input:       []byte("TRUE"),
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        `should parse "1" bytes to bool true`,
			input:       []byte("1"),
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        `should parse "t" bytes to bool true`,
			input:       []byte("t"),
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        `should parse "T" bytes to bool true`,
			input:       []byte("T"),
			expectedVal: true,
			expectedErr: nil,
		},

		// Boolean - false variants
		{
			name:        `should parse "false" bytes to bool false`,
			input:       []byte("false"),
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        `should parse "FALSE" bytes to bool false`,
			input:       []byte("FALSE"),
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        `should parse "0" bytes to bool false`,
			input:       []byte("0"),
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        `should parse "f" bytes to bool false`,
			input:       []byte("f"),
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        `should parse "F" bytes to bool false`,
			input:       []byte("F"),
			expectedVal: false,
			expectedErr: nil,
		},

		// Boolean - invalid
		{
			name:        "should return error for invalid bool bytes",
			input:       []byte("notabool"),
			expectedVal: false,
			expectedErr: errors.New(`convert: failed to parse bool from bytes: strconv.ParseBool: parsing "notabool": invalid syntax`),
		},

		// Struct
		{
			name:        "should unmarshal bytes to struct successfully",
			input:       structJSON,
			expectedVal: structValue,
			expectedErr: nil,
		},
		{
			name:        "should return error for invalid struct JSON bytes",
			input:       []byte("{invalid_json}"),
			expectedVal: testStruct{},
			expectedErr: fmt.Errorf("convert: failed to unmarshal bytes to T: invalid character 'i' looking for beginning of object key string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch expected := tt.expectedVal.(type) {
			case string:
				val, err := parseFromBytes[string](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case int:
				val, err := parseFromBytes[int](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case int64:
				val, err := parseFromBytes[int64](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case bool:
				val, err := parseFromBytes[bool](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case testStruct:
				val, err := parseFromBytes[testStruct](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			default:
				t.Fatalf("unhandled expected type %T", expected)
			}
		})
	}
}

func TestGenericManager_parseFromString(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	structValue := testStruct{ID: 1, Name: "John"}
	structJSON, _ := json.Marshal(structValue)

	tests := []struct {
		name        string
		input       string
		expectedVal any
		expectedErr error
	}{
		// String
		{
			name:        "should parse string to string successfully",
			input:       "hello",
			expectedVal: "hello",
			expectedErr: nil,
		},

		// Int
		{
			name:        "should parse string to int successfully",
			input:       "123",
			expectedVal: 123,
			expectedErr: nil,
		},
		{
			name:        "should return error for invalid int string",
			input:       "abc",
			expectedVal: 0,
			expectedErr: errors.New("convert: failed to parse int from string: strconv.Atoi: parsing \"abc\": invalid syntax"),
		},

		// Int64
		{
			name:        "should parse string to int64 successfully",
			input:       "456",
			expectedVal: int64(456),
			expectedErr: nil,
		},
		{
			name:        "should return error for invalid int64 string",
			input:       "notanumber",
			expectedVal: int64(0),
			expectedErr: errors.New("convert: failed to parse int64 from string: strconv.ParseInt: parsing \"notanumber\": invalid syntax"),
		},

		// Boolean - true variants
		{
			name:        `should parse "true" string to bool true`,
			input:       "true",
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        `should parse "TRUE" string to bool true`,
			input:       "TRUE",
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        `should parse "1" string to bool true`,
			input:       "1",
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        `should parse "t" string to bool true`,
			input:       "t",
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        `should parse "T" string to bool true`,
			input:       "T",
			expectedVal: true,
			expectedErr: nil,
		},

		// Boolean - false variants
		{
			name:        `should parse "false" string to bool false`,
			input:       "false",
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        `should parse "FALSE" string to bool false`,
			input:       "FALSE",
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        `should parse "0" string to bool false`,
			input:       "0",
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        `should parse "f" string to bool false`,
			input:       "f",
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        `should parse "F" string to bool false`,
			input:       "F",
			expectedVal: false,
			expectedErr: nil,
		},

		// Boolean - invalid
		{
			name:        "should return error for invalid bool string",
			input:       "notabool",
			expectedVal: false,
			expectedErr: errors.New(`convert: failed to parse bool from string: strconv.ParseBool: parsing "notabool": invalid syntax`),
		},

		// Struct
		{
			name:        "should unmarshal string to struct successfully",
			input:       string(structJSON),
			expectedVal: structValue,
			expectedErr: nil,
		},
		{
			name:        "should return error for invalid struct JSON",
			input:       "{invalid_json}",
			expectedVal: testStruct{},
			expectedErr: fmt.Errorf("convert: failed to unmarshal string to T: invalid character 'i' looking for beginning of object key string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch expected := tt.expectedVal.(type) {
			case string:
				val, err := parseFromString[string](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case int:
				val, err := parseFromString[int](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case int64:
				val, err := parseFromString[int64](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case bool:
				val, err := parseFromString[bool](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			case testStruct:
				val, err := parseFromString[testStruct](tt.input)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error")
				} else {
					assert.NoError(t, err, "expected no error")
				}
				assert.Equal(t, expected, val, "expected correct value")

			default:
				t.Fatalf("unhandled expected type %T", expected)
			}
		})
	}
}
