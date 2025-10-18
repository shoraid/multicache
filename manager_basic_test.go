package omnicache

import (
	"context"
	"errors"
	"testing"
	"time"

	omnicachemock "github.com/shoraid/omnicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestManager_Get(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockErr       error
		expectedValue any
		expectedErr   error
	}{
		{
			name:          "should get value successfully",
			key:           key,
			mockValue:     "test-value",
			mockErr:       nil,
			expectedValue: "test-value",
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
			name:          "should return other error",
			key:           key,
			mockValue:     nil,
			mockErr:       errors.New("some other error"),
			expectedValue: nil,
			expectedErr:   errors.New("some other error"),
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
			value, err := manager.Get(ctx, tt.key)

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

func TestManager_GetOrSet(t *testing.T) {
	t.Parallel()

	key := "test-key"
	ttl := 5 * time.Minute
	defaultValue := "default-value"
	defaultFn := func() (any, error) {
		return defaultValue, nil
	}
	defaultFnErr := func() (any, error) {
		return nil, errors.New("default function error")
	}

	tests := []struct {
		name              string
		key               string
		getMockValue      any
		getMockErr        error
		setMockErr        error
		defaultFn         func() (any, error)
		expectedValue     any
		expectedErr       error
		expectedSetCalled bool
	}{
		{
			name:              "should return existing value if cache hit",
			key:               key,
			getMockValue:      "cached-value",
			getMockErr:        nil,
			defaultFn:         defaultFn,
			expectedValue:     "cached-value",
			expectedErr:       nil,
			expectedSetCalled: false,
		},
		{
			name:              "should set and return default value if cache miss",
			key:               key,
			getMockValue:      nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        nil,
			defaultFn:         defaultFn,
			expectedValue:     defaultValue,
			expectedErr:       nil,
			expectedSetCalled: true,
		},
		{
			name:              "should return error from Get if not cache miss",
			key:               key,
			getMockValue:      nil,
			getMockErr:        errors.New("get error"),
			defaultFn:         defaultFn,
			expectedValue:     nil,
			expectedErr:       errors.New("get error"),
			expectedSetCalled: false,
		},
		{
			name:              "should return error from defaultFn if cache miss and defaultFn fails",
			key:               key,
			getMockValue:      nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        nil,
			defaultFn:         defaultFnErr,
			expectedValue:     nil,
			expectedErr:       errors.New("default function error"),
			expectedSetCalled: false,
		},
		{
			name:              "should return default value but also error if Set fails after cache miss",
			key:               key,
			getMockValue:      nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        errors.New("set error"),
			defaultFn:         defaultFn,
			expectedValue:     defaultValue,
			expectedErr:       errors.New("set error"),
			expectedSetCalled: true,
		},
		{
			name:              "should return error if Get returns non-cache-miss error",
			key:               key,
			getMockValue:      nil,
			getMockErr:        errors.New("some random error"),
			defaultFn:         defaultFn,
			expectedValue:     nil,
			expectedErr:       errors.New("some random error"),
			expectedSetCalled: false,
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
				assert.Equal(t, tt.key, key, "expected correct key to be used in Get")
				return tt.getMockValue, tt.getMockErr
			}
			mockStore.SetFunc = func(_ context.Context, key string, value any, duration time.Duration) error {
				assert.Equal(t, tt.key, key, "expected correct key to be used in Set")
				assert.Equal(t, tt.expectedValue, value, "expected correct value to be set")
				assert.Equal(t, ttl, duration, "expected correct ttl to be used in Set")
				return tt.setMockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			result, err := manager.GetOrSet(ctx, tt.key, ttl, tt.defaultFn)

			// Assert
			mockStore.CalledOnce(t, "Get")

			if tt.expectedSetCalled {
				mockStore.CalledOnce(t, "Set")
			}

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedValue, result, "expected correct value on error")
				return
			}

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, tt.expectedValue, result, "expected correct value")
		})
	}
}

func TestManager_Has(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name          string
		key           string
		mockValue     bool
		mockErr       error
		expectedValue bool
		expectedErr   error
	}{
		{
			name:          "should return true if key exists",
			key:           key,
			mockValue:     true,
			mockErr:       nil,
			expectedValue: true,
			expectedErr:   nil,
		},
		{
			name:          "should return false if key does not exist",
			key:           key,
			mockValue:     false,
			mockErr:       nil,
			expectedValue: false,
			expectedErr:   nil,
		},
		{
			name:          "should return false if key exists but is expired",
			key:           key,
			mockValue:     false,
			mockErr:       nil,
			expectedValue: false,
			expectedErr:   nil,
		},
		{
			name:          "should return error if store returns an error other than cache miss",
			key:           key,
			mockValue:     false,
			mockErr:       errors.New("some other error"),
			expectedValue: false,
			expectedErr:   errors.New("some other error"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.HasFunc = func(_ context.Context, key string) (bool, error) {
				assert.Equal(t, tt.key, key, "expected correct key to be used")
				return tt.mockValue, tt.mockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			result, err := manager.Has(ctx, tt.key)

			// Assert
			mockStore.CalledOnce(t, "Has")

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedValue, result, "expected correct result on error")
				return
			}

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, tt.expectedValue, result, "expected correct result")
		})
	}
}

func TestManager_Set(t *testing.T) {
	t.Parallel()

	key := "test-key"
	value := "test-value"
	ttl := 5 * time.Minute

	tests := []struct {
		name        string
		key         string
		value       any
		ttl         time.Duration
		mockErr     error
		expectedErr bool
	}{
		{
			name:        "should set value successfully",
			key:         key,
			value:       value,
			ttl:         ttl,
			mockErr:     nil,
			expectedErr: false,
		},
		{
			name:        "should return error when set fails",
			key:         key,
			value:       value,
			ttl:         ttl,
			mockErr:     errors.New("set failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.SetFunc = func(_ context.Context, key string, value any, ttl time.Duration) error {
				assert.Equal(t, tt.key, key, "expected correct key to be used")
				assert.Equal(t, tt.value, value, "expected correct value to be used")
				assert.Equal(t, tt.ttl, ttl, "expected correct ttl to be used")

				return tt.mockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			err := manager.Set(ctx, tt.key, tt.value, tt.ttl)

			// Assert
			mockStore.CalledOnce(t, "Set")

			if tt.expectedErr {
				assert.ErrorIs(t, err, tt.mockErr, "expected error when set fails")
				return
			}

			assert.NoError(t, err, "expected no error when set succeeds")
		})
	}
}
