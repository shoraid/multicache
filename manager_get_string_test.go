package omnicache

import (
	"context"
	"errors"
	"testing"
	"time"

	omnicachemock "github.com/shoraid/omnicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestManager_GetString(t *testing.T) {
	t.Parallel()

	key := "test-string-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockErr       error
		expectedValue string
		expectedErr   error
	}{
		{
			name:          "should get string successfully",
			key:           key,
			mockValue:     "hello world",
			mockErr:       nil,
			expectedValue: "hello world",
			expectedErr:   nil,
		},
		{
			name:          "should get string successfully from []byte",
			key:           key,
			mockValue:     []byte("hello byte world"),
			mockErr:       nil,
			expectedValue: "hello byte world",
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockErr:       ErrCacheMiss,
			expectedValue: "",
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-string value",
			key:           key,
			mockValue:     make(chan int),
			mockErr:       nil,
			expectedValue: "",
			expectedErr:   ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.GetFunc = func(_ context.Context, key string) (any, error) {
				assert.Equal(t, tt.key, key, "expected correct key to be used")
				return tt.mockValue, tt.mockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			value, err := manager.GetString(ctx, tt.key)

			// Assert
			mockStore.CalledOnce(t, "Get")

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedValue, value, "expected default value on error")
				return
			}

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, tt.expectedValue, value, "expected correct value")
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
		mockErr       error
		expectedValue []string
		expectedErr   error
	}{
		{
			name:          "should get []string successfully from []string",
			key:           key,
			mockValue:     []string{"a", "b", "c"},
			mockErr:       nil,
			expectedValue: []string{"a", "b", "c"},
			expectedErr:   nil,
		},
		{
			name:          "should get []string successfully from JSON string",
			key:           key,
			mockValue:     `["d","e","f"]`,
			mockErr:       nil,
			expectedValue: []string{"d", "e", "f"},
			expectedErr:   nil,
		},
		{
			name:          "should get []string successfully from JSON []byte",
			key:           key,
			mockValue:     []byte(`["g","h","i"]`),
			mockErr:       nil,
			expectedValue: []string{"g", "h", "i"},
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockErr:       ErrCacheMiss,
			expectedValue: nil,
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-[]string value",
			key:           key,
			mockValue:     123,
			mockErr:       nil,
			expectedValue: nil,
			expectedErr:   ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch error for invalid JSON string",
			key:           key,
			mockValue:     "invalid json",
			mockErr:       nil,
			expectedValue: nil,
			expectedErr:   ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch error for invalid JSON []byte",
			key:           key,
			mockValue:     []byte("invalid json"),
			mockErr:       nil,
			expectedValue: nil,
			expectedErr:   ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.GetFunc = func(_ context.Context, key string) (any, error) {
				assert.Equal(t, tt.key, key, "expected correct key to be used")
				return tt.mockValue, tt.mockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			value, err := manager.GetStrings(ctx, tt.key)

			// Assert
			mockStore.CalledOnce(t, "Get")

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Nil(t, value, "expected nil value on error")
				return
			}

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, tt.expectedValue, value, "expected correct value")
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
		name          string
		key           string
		mockGetVal    any
		mockGetErr    error
		mockSetErr    error
		defaultFunc   func() (string, error)
		expectedVal   string
		expectedErr   error
		expectSetCall bool
	}{
		{
			name:          "should return existing value if found",
			key:           key,
			mockGetVal:    "existing-string",
			mockGetErr:    nil,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   "existing-string",
			expectedErr:   nil,
			expectSetCall: false,
		},
		{
			name:          "should set and return default if cache miss",
			key:           key,
			mockGetVal:    nil,
			mockGetErr:    ErrCacheMiss,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   defaultValue,
			expectedErr:   nil,
			expectSetCall: true,
		},
		{
			name:          "should set and return default if type mismatch on get",
			key:           key,
			mockGetVal:    make(chan int), // GetString will return ErrTypeMismatch
			mockGetErr:    nil,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   defaultValue,
			expectedErr:   nil,
			expectSetCall: true,
		},
		{
			name:          "should return error if get fails with other error",
			key:           key,
			mockGetVal:    nil,
			mockGetErr:    errors.New("network error"),
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   "", // Default value for string
			expectedErr:   errors.New("network error"),
			expectSetCall: false,
		},
		{
			name:          "should return error if default function fails",
			key:           key,
			mockGetVal:    nil,
			mockGetErr:    ErrCacheMiss,
			mockSetErr:    nil,
			defaultFunc:   errorFn,
			expectedVal:   "",
			expectedErr:   errors.New("default function error"),
			expectSetCall: false, // Set should not be called if defaultFn fails
		},
		{
			name:          "should return error if set fails after cache miss",
			key:           key,
			mockGetVal:    nil,
			mockGetErr:    ErrCacheMiss,
			mockSetErr:    errors.New("set operation failed"),
			defaultFunc:   defaultFn,
			expectedVal:   defaultValue,
			expectedErr:   errors.New("set operation failed"),
			expectSetCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			ttl := 5 * time.Minute

			mockStore := new(omnicachemock.MockStore)
			mockStore.GetFunc = func(_ context.Context, key string) (any, error) {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Get")
				if tt.mockGetVal != nil && errors.Is(tt.mockGetErr, ErrTypeMismatch) {
					// Simulate GetString's internal type conversion failure
					return tt.mockGetVal, nil
				}
				return tt.mockGetVal, tt.mockGetErr
			}
			mockStore.SetFunc = func(_ context.Context, key string, value any, duration time.Duration) error {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Set")
				assert.Equal(t, defaultValue, value, "expected correct value to be set")
				assert.Equal(t, ttl, duration, "expected correct ttl to be used in Set")
				return tt.mockSetErr
			}

			manager := &Manager{store: mockStore}

			// Act
			value, err := manager.GetStringOrSet(ctx, tt.key, ttl, tt.defaultFunc)

			// Assert
			mockStore.CalledOnce(t, "Get")

			if tt.expectSetCall {
				mockStore.CalledOnce(t, "Set")
			} else {
				mockStore.NotCalled(t, "Set")
			}

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedVal, value, "expected default value on error")
				return
			}

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, tt.expectedVal, value, "expected correct value")
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
		name          string
		key           string
		mockGetVal    any
		mockGetErr    error
		mockSetErr    error
		defaultFunc   func() ([]string, error)
		expectedVal   []string
		expectedErr   error
		expectSetCall bool
	}{
		{
			name:          "should return existing value if found",
			key:           key,
			mockGetVal:    []string{"existing1", "existing2"},
			mockGetErr:    nil,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   []string{"existing1", "existing2"},
			expectedErr:   nil,
			expectSetCall: false,
		},
		{
			name:          "should set and return default if cache miss",
			key:           key,
			mockGetVal:    nil,
			mockGetErr:    ErrCacheMiss,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   defaultValue,
			expectedErr:   nil,
			expectSetCall: true,
		},
		{
			name:          "should set and return default if type mismatch on get",
			key:           key,
			mockGetVal:    123, // GetStrings will return ErrTypeMismatch
			mockGetErr:    nil,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   defaultValue,
			expectedErr:   nil,
			expectSetCall: true,
		},
		{
			name:          "should return error if get fails with other error",
			key:           key,
			mockGetVal:    nil,
			mockGetErr:    errors.New("network error"),
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   nil, // Default value for []string
			expectedErr:   errors.New("network error"),
			expectSetCall: false,
		},
		{
			name:          "should return error if default function fails",
			key:           key,
			mockGetVal:    nil,
			mockGetErr:    ErrCacheMiss,
			mockSetErr:    nil,
			defaultFunc:   errorFn,
			expectedVal:   nil,
			expectedErr:   errors.New("default function error"),
			expectSetCall: false, // Set should not be called if defaultFn fails
		},
		{
			name:          "should return error if set fails after cache miss",
			key:           key,
			mockGetVal:    nil,
			mockGetErr:    ErrCacheMiss,
			mockSetErr:    errors.New("set operation failed"),
			defaultFunc:   defaultFn,
			expectedVal:   defaultValue,
			expectedErr:   errors.New("set operation failed"),
			expectSetCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			ttl := 5 * time.Minute

			mockStore := new(omnicachemock.MockStore)
			mockStore.GetFunc = func(_ context.Context, key string) (any, error) {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Get")
				if tt.mockGetVal != nil && errors.Is(tt.mockGetErr, ErrTypeMismatch) {
					// Simulate GetStrings' internal type conversion failure
					return tt.mockGetVal, nil
				}
				return tt.mockGetVal, tt.mockGetErr
			}
			mockStore.SetFunc = func(_ context.Context, key string, value any, duration time.Duration) error {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Set")
				assert.Equal(t, defaultValue, value, "expected correct value to be set")
				assert.Equal(t, ttl, duration, "expected correct ttl to be used in Set")
				return tt.mockSetErr
			}

			manager := &Manager{store: mockStore}

			// Act
			value, err := manager.GetStringsOrSet(ctx, tt.key, ttl, tt.defaultFunc)

			// Assert
			mockStore.CalledOnce(t, "Get")

			if tt.expectSetCall {
				mockStore.CalledOnce(t, "Set")
			} else {
				mockStore.NotCalled(t, "Set")
			}

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedVal, value, "expected default value on error")
				return
			}

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, tt.expectedVal, value, "expected correct value")
		})
	}
}
