package omnicache

import (
	"context"
	"errors"
	"testing"
	"time"

	omnicachemock "github.com/shoraid/omnicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestManager_GetBool(t *testing.T) {
	t.Parallel()

	key := "test-bool-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockErr       error
		expectedValue bool
		expectedErr   error
	}{
		{
			name:          "should get true bool successfully",
			key:           key,
			mockValue:     true,
			mockErr:       nil,
			expectedValue: true,
			expectedErr:   nil,
		},
		{
			name:          "should get false bool successfully",
			key:           key,
			mockValue:     false,
			mockErr:       nil,
			expectedValue: false,
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockErr:       ErrCacheMiss,
			expectedValue: false,
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-bool value",
			key:           key,
			mockValue:     "not a bool",
			mockErr:       nil,
			expectedValue: false,
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
			value, err := manager.GetBool(ctx, tt.key)

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

func TestManager_GetBoolOrSet(t *testing.T) {
	t.Parallel()

	key := "test-bool-key"
	defaultValue := true
	defaultFn := func() (bool, error) {
		return defaultValue, nil
	}
	errorFn := func() (bool, error) {
		return false, errors.New("default function error")
	}

	tests := []struct {
		name          string
		key           string
		mockGetVal    any
		mockGetErr    error
		mockSetErr    error
		defaultFunc   func() (bool, error)
		expectedVal   bool
		expectedErr   error
		expectSetCall bool
	}{
		{
			name:          "should return existing value if found",
			key:           key,
			mockGetVal:    false,
			mockGetErr:    nil,
			mockSetErr:    nil,
			defaultFunc:   defaultFn,
			expectedVal:   false,
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
			mockGetVal:    "not a bool", // GetBool will return ErrTypeMismatch
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
			expectedVal:   false, // Default value for bool
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
			expectedVal:   false,
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
					// Simulate GetBool's internal type conversion failure
					return tt.mockGetVal, nil
				}
				return tt.mockGetVal, tt.mockGetErr
			}
			mockStore.SetFunc = func(_ context.Context, key string, value any, duration time.Duration) error {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Set")
				assert.Equal(t, tt.expectedVal, value, "expected correct value to be set")
				assert.Equal(t, ttl, duration, "expected correct ttl to be used in Set")
				return tt.mockSetErr
			}

			manager := &Manager{store: mockStore}

			// Act
			value, err := manager.GetBoolOrSet(ctx, tt.key, ttl, tt.defaultFunc)

			// Assert
			mockStore.CalledOnce(t, "Get")

			if tt.expectSetCall {
				mockStore.CalledOnce(t, "Set")
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
