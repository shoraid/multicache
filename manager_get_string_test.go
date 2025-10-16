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

func TestManager_GetString(t *testing.T) {
	t.Parallel()

	key := "test-string-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectedValue string
		expectedErr   error
	}{
		{
			name:          "should get string successfully",
			key:           key,
			mockValue:     "hello world",
			mockReturnErr: nil,
			expectedValue: "hello world",
			expectedErr:   nil,
		},
		{
			name:          "should get string successfully from []byte",
			key:           key,
			mockValue:     []byte("hello byte world"),
			mockReturnErr: nil,
			expectedValue: "hello byte world",
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectedValue: "",
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-string value",
			key:           key,
			mockValue:     make(chan int),
			mockReturnErr: nil,
			expectedValue: "",
			expectedErr:   ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockStore := new(multicachemock.MockStore)

			manager := &managerImpl{
				stores: map[string]contract.Store{"default": mockStore},
				store:  mockStore,
			}

			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetString(ctx, tt.key)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedValue, value, "expected default value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetStrings(t *testing.T) {
	t.Parallel()

	key := "test-strings-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectedValue []string
		expectedErr   error
	}{
		{
			name:          "should get []string successfully from []string",
			key:           key,
			mockValue:     []string{"a", "b", "c"},
			mockReturnErr: nil,
			expectedValue: []string{"a", "b", "c"},
			expectedErr:   nil,
		},
		{
			name:          "should get []string successfully from JSON string",
			key:           key,
			mockValue:     `["d","e","f"]`,
			mockReturnErr: nil,
			expectedValue: []string{"d", "e", "f"},
			expectedErr:   nil,
		},
		{
			name:          "should get []string successfully from JSON []byte",
			key:           key,
			mockValue:     []byte(`["g","h","i"]`),
			mockReturnErr: nil,
			expectedValue: []string{"g", "h", "i"},
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectedValue: nil,
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-[]string value",
			key:           key,
			mockValue:     123,
			mockReturnErr: nil,
			expectedValue: nil,
			expectedErr:   ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch error for invalid JSON string",
			key:           key,
			mockValue:     "invalid json",
			mockReturnErr: nil,
			expectedValue: nil,
			expectedErr:   ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch error for invalid JSON []byte",
			key:           key,
			mockValue:     []byte("invalid json"),
			mockReturnErr: nil,
			expectedValue: nil,
			expectedErr:   ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockStore := new(multicachemock.MockStore)

			manager := &managerImpl{
				stores: map[string]contract.Store{"default": mockStore},
				store:  mockStore,
			}

			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetStrings(ctx, tt.key)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Nil(t, value, "expected nil value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetStringOrSet(t *testing.T) {
	t.Parallel()

	key := "test-string-key"
	defaultValue := "default-string"
	defaultFn := func() (string, error) {
		return defaultValue, nil
	}
	errorFn := func() (string, error) {
		return "", errors.New("default function error")
	}

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		defaultFunc       func() (string, error)
		expectedReturnVal string
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        "existing-string",
			mockGetErr:        nil,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: "existing-string",
			expectedErr:       nil,
			expectSetCall:     false,
		},
		{
			name:              "should set and return default if cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: defaultValue,
			expectedErr:       nil,
			expectSetCall:     true,
		},
		{
			name:              "should set and return default if type mismatch on get",
			key:               key,
			mockGetVal:        make(chan int), // GetString will return ErrTypeMismatch
			mockGetErr:        nil,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: defaultValue,
			expectedErr:       nil,
			expectSetCall:     true,
		},
		{
			name:              "should return error if get fails with other error",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        errors.New("network error"),
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: "", // Default value for string
			expectedErr:       errors.New("network error"),
			expectSetCall:     false,
		},
		{
			name:              "should return error if default function fails",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        nil,
			defaultFunc:       errorFn,
			expectedReturnVal: "",
			expectedErr:       errors.New("default function error"),
			expectSetCall:     false, // Set should not be called if defaultFn fails
		},
		{
			name:              "should return error if set fails after cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        errors.New("set operation failed"),
			defaultFunc:       defaultFn,
			expectedReturnVal: defaultValue,
			expectedErr:       errors.New("set operation failed"),
			expectSetCall:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ttl := 5 * time.Minute
			mockStore := new(multicachemock.MockStore)

			manager := &managerImpl{
				stores: map[string]contract.Store{"default": mockStore},
				store:  mockStore,
			}

			mockStore.ExpectedCalls = nil // Reset calls for isolation

			// Mock the initial Get call
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockGetVal, tt.mockGetErr).
				Once()

			// Mock the Set call if expected
			if tt.expectSetCall {
				mockStore.
					On("Set", ctx, tt.key, defaultValue, ttl).
					Return(tt.mockSetErr).
					Once()
			}

			value, err := manager.GetStringOrSet(ctx, tt.key, ttl, tt.defaultFunc)
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedReturnVal, value, "expected default value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedReturnVal, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetStringsOrSet(t *testing.T) {
	t.Parallel()

	key := "test-strings-key"
	defaultValue := []string{"default1", "default2"}
	defaultFn := func() ([]string, error) {
		return defaultValue, nil
	}
	errorFn := func() ([]string, error) {
		return nil, errors.New("default function error")
	}

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		defaultFunc       func() ([]string, error)
		expectedReturnVal []string
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        []string{"existing1", "existing2"},
			mockGetErr:        nil,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: []string{"existing1", "existing2"},
			expectedErr:       nil,
			expectSetCall:     false,
		},
		{
			name:              "should set and return default if cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: defaultValue,
			expectedErr:       nil,
			expectSetCall:     true,
		},
		{
			name:              "should set and return default if type mismatch on get",
			key:               key,
			mockGetVal:        123, // GetStrings will return ErrTypeMismatch
			mockGetErr:        nil,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: defaultValue,
			expectedErr:       nil,
			expectSetCall:     true,
		},
		{
			name:              "should return error if get fails with other error",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        errors.New("network error"),
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: nil, // Default value for []string
			expectedErr:       errors.New("network error"),
			expectSetCall:     false,
		},
		{
			name:              "should return error if default function fails",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        nil,
			defaultFunc:       errorFn,
			expectedReturnVal: nil,
			expectedErr:       errors.New("default function error"),
			expectSetCall:     false, // Set should not be called if defaultFn fails
		},
		{
			name:              "should return error if set fails after cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        errors.New("set operation failed"),
			defaultFunc:       defaultFn,
			expectedReturnVal: defaultValue,
			expectedErr:       errors.New("set operation failed"),
			expectSetCall:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ttl := 5 * time.Minute
			mockStore := new(multicachemock.MockStore)

			manager := &managerImpl{
				stores: map[string]contract.Store{"default": mockStore},
				store:  mockStore,
			}

			mockStore.ExpectedCalls = nil // Reset calls for isolation

			// Mock the initial Get call
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockGetVal, tt.mockGetErr).
				Once()

			// Mock the Set call if expected
			if tt.expectSetCall {
				mockStore.
					On("Set", ctx, tt.key, defaultValue, ttl).
					Return(tt.mockSetErr).
					Once()
			}

			value, err := manager.GetStringsOrSet(ctx, tt.key, ttl, tt.defaultFunc)
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedReturnVal, value, "expected default value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedReturnVal, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}
