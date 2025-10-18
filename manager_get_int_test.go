package omnicache

import (
	"context"
	"errors"
	"testing"
	"time"

	omnicachemock "github.com/shoraid/omnicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestManager_GetInt(t *testing.T) {
	t.Parallel()

	key := "test-int-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockErr       error
		expectedValue int
		expectedErr   error
	}{
		{
			name:          "should get int successfully",
			key:           key,
			mockValue:     123,
			mockErr:       nil,
			expectedValue: 123,
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockErr:       ErrCacheMiss,
			expectedValue: 0,
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-int value",
			key:           key,
			mockValue:     "not an int",
			mockErr:       nil,
			expectedValue: 0,
			expectedErr:   ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		tt := tt

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
			value, err := manager.GetInt(ctx, tt.key)

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

func TestManager_GetInt64(t *testing.T) {
	t.Parallel()

	key := "test-int64-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockErr       error
		expectedValue int64
		expectedErr   error
	}{
		{
			name:          "should get int64 successfully",
			key:           key,
			mockValue:     int64(123456789012345),
			mockErr:       nil,
			expectedValue: int64(123456789012345),
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockErr:       ErrCacheMiss,
			expectedValue: 0,
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-int64 value",
			key:           key,
			mockValue:     "not an int64",
			mockErr:       nil,
			expectedValue: 0,
			expectedErr:   ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		tt := tt

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
			value, err := manager.GetInt64(ctx, tt.key)

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

func TestManager_GetInts(t *testing.T) {
	t.Parallel()

	key := "test-ints-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockErr       error
		expectedValue []int
		expectedErr   error
	}{
		{
			name:          "should get []int successfully from []int",
			key:           key,
			mockValue:     []int{1, 2, 3},
			mockErr:       nil,
			expectedValue: []int{1, 2, 3},
			expectedErr:   nil,
		},
		{
			name:          "should get []int successfully from JSON string",
			key:           key,
			mockValue:     "[4,5,6]",
			mockErr:       nil,
			expectedValue: []int{4, 5, 6},
			expectedErr:   nil,
		},
		{
			name:          "should get []int successfully from JSON []byte",
			key:           key,
			mockValue:     []byte("[7,8,9]"),
			mockErr:       nil,
			expectedValue: []int{7, 8, 9},
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
			name:          "should return type mismatch error for non-[]int value",
			key:           key,
			mockValue:     "not an int array",
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
		{
			name:          "should return type mismatch error for non-[]int value",
			key:           key,
			mockValue:     true,
			mockErr:       nil,
			expectedValue: nil,
			expectedErr:   ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		tt := tt

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
			value, err := manager.GetInts(ctx, tt.key)

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

func TestManager_GetIntOrSet(t *testing.T) {
	t.Parallel()

	key := "test-int-key"
	defaultValue := 99
	defaultFn := func() (int, error) {
		return defaultValue, nil
	}
	errorFn := func() (int, error) {
		return 0, errors.New("default function error")
	}

	tests := []struct {
		name          string
		key           string
		mockGetVal    any
		mockGetErr    error
		mockSetErr    error
		defaultFunc   func() (int, error)
		expectedVal   int
		expectedErr   error
		expectSetCall bool
	}{
		{
			name:          "should return existing value if found",
			key:           key,
			mockGetVal:    123,
			mockGetErr:    nil,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   123,
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
			mockGetVal:    "not an int", // GetInt will return ErrTypeMismatch
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
			expectedVal:   0, // Default value for int
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
			expectedVal:   0,
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
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			ttl := 5 * time.Minute

			mockStore := new(omnicachemock.MockStore)
			mockStore.GetFunc = func(_ context.Context, key string) (any, error) {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Get")
				if tt.mockGetVal != nil && errors.Is(tt.mockGetErr, ErrTypeMismatch) {
					// Simulate GetInt's internal type conversion failure
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
			value, err := manager.GetIntOrSet(ctx, tt.key, ttl, tt.defaultFunc)

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

func TestManager_GetInt64OrSet(t *testing.T) {
	t.Parallel()

	key := "test-int64-key"
	defaultValue := int64(999999999999)
	defaultFn := func() (int64, error) {
		return defaultValue, nil
	}
	errorFn := func() (int64, error) {
		return 0, errors.New("default function error")
	}

	tests := []struct {
		name          string
		key           string
		mockGetVal    any
		mockGetErr    error
		mockSetErr    error
		defaultFunc   func() (int64, error)
		expectedVal   int64
		expectedErr   error
		expectSetCall bool
	}{
		{
			name:          "should return existing value if found",
			key:           key,
			mockGetVal:    int64(123456789012345),
			mockGetErr:    nil,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   int64(123456789012345),
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
			mockGetVal:    "not an int64", // GetInt64 will return ErrTypeMismatch
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
			expectedVal:   0, // Default value for int64
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
			expectedVal:   0,
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
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			ttl := 5 * time.Minute

			mockStore := new(omnicachemock.MockStore)
			mockStore.GetFunc = func(_ context.Context, key string) (any, error) {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Get")
				if tt.mockGetVal != nil && errors.Is(tt.mockGetErr, ErrTypeMismatch) {
					// Simulate GetInt64's internal type conversion failure
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
			value, err := manager.GetInt64OrSet(ctx, tt.key, ttl, tt.defaultFunc)

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

func TestManager_GetIntsOrSet(t *testing.T) {
	t.Parallel()

	key := "test-ints-key"
	defaultValue := []int{10, 20, 30}
	defaultFn := func() ([]int, error) {
		return defaultValue, nil
	}
	errorFn := func() ([]int, error) {
		return nil, errors.New("default function error")
	}

	tests := []struct {
		name          string
		key           string
		mockGetVal    any
		mockGetErr    error
		mockSetErr    error
		defaultFunc   func() ([]int, error)
		expectedVal   []int
		expectedErr   error
		expectSetCall bool
	}{
		{
			name:          "should return existing value if found",
			key:           key,
			mockGetVal:    []int{1, 2, 3},
			mockGetErr:    nil,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   []int{1, 2, 3},
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
			mockGetVal:    "not an int array", // GetInts will return ErrTypeMismatch
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
			expectedVal:   nil, // Default value for []int
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
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			ttl := 5 * time.Minute

			mockStore := new(omnicachemock.MockStore)
			mockStore.GetFunc = func(_ context.Context, key string) (any, error) {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Get")
				if tt.mockGetVal != nil && errors.Is(tt.mockGetErr, ErrTypeMismatch) {
					// Simulate GetInts' internal type conversion failure
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
			value, err := manager.GetIntsOrSet(ctx, tt.key, ttl, tt.defaultFunc)

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
