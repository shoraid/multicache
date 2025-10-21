package omnicache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shoraid/omnicache/internal/assert"
	omnicachemock "github.com/shoraid/omnicache/mock"
)

func TestManager_Get(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name        string
		key         string
		mockVal     any
		mockErr     error
		expectedVal any
		expectedErr error
	}{
		{
			name:        "should successfully get the value when it exists and is not expired",
			key:         key,
			mockVal:     "test-value",
			mockErr:     nil,
			expectedVal: "test-value",
			expectedErr: nil,
		},
		{
			name:        "should return a cache miss error when the key does not exist",
			key:         key,
			mockVal:     nil,
			mockErr:     ErrCacheMiss,
			expectedVal: nil,
			expectedErr: ErrCacheMiss,
		},
		{
			name:        "should return the same error when the store returns an error other than cache miss",
			key:         key,
			mockVal:     nil,
			mockErr:     errors.New("some other error"),
			expectedVal: nil,
			expectedErr: errors.New("some other error"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)
			mockStore.Mock.On("Get", ctx, tt.key).Return(tt.mockVal, tt.mockErr)

			manager := &Manager{store: mockStore}

			// --- Act ---
			value, err := manager.Get(ctx, tt.key)

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "Get", ctx, tt.key)
			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return an error when Get fails")
				assert.EqualError(t, tt.expectedErr, err, "error returned by Get must match the expected error")
				assert.Nil(t, value, "value must be nil when Get returns an error")
				return
			}

			assert.NoError(t, err, "must not return an error when Get succeeds")
			assert.Equal(t, tt.expectedVal, value, "value returned by Get must match the expected value")
		})
	}
}

func TestManager_GetOrSet(t *testing.T) {
	t.Parallel()

	key := "test-key"
	ttl := 5 * time.Minute
	defaultValue := "default-value"
	defaultFn := func() (any, error) { return defaultValue, nil }
	defaultFnErr := func() (any, error) { return nil, errors.New("default function error") }

	tests := []struct {
		name              string
		key               string
		getMockVal        any
		getMockErr        error
		setMockErr        error
		defaultFn         func() (any, error)
		expectedVal       any
		expectedErr       error
		expectedSetCalled bool
	}{
		{
			name:              "should return existing value when cache hit",
			key:               key,
			getMockVal:        "cached-value",
			getMockErr:        nil,
			defaultFn:         defaultFn,
			expectedVal:       "cached-value",
			expectedErr:       nil,
			expectedSetCalled: false,
		},
		{
			name:              "should set and return default value when cache miss",
			key:               key,
			getMockVal:        nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        nil,
			defaultFn:         defaultFn,
			expectedVal:       defaultValue,
			expectedErr:       nil,
			expectedSetCalled: true,
		},
		{
			name:              "should return error when Get returns a non-cache-miss error",
			key:               key,
			getMockVal:        nil,
			getMockErr:        errors.New("get error"),
			defaultFn:         defaultFn,
			expectedVal:       nil,
			expectedErr:       errors.New("get error"),
			expectedSetCalled: false,
		},
		{
			name:              "should return error when defaultFn fails after cache miss",
			key:               key,
			getMockVal:        nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        nil,
			defaultFn:         defaultFnErr,
			expectedVal:       nil,
			expectedErr:       errors.New("default function error"),
			expectedSetCalled: false,
		},
		{
			name:              "should return default value but also return error when Set fails after cache miss",
			key:               key,
			getMockVal:        nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        errors.New("set error"),
			defaultFn:         defaultFn,
			expectedVal:       defaultValue,
			expectedErr:       errors.New("set error"),
			expectedSetCalled: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)

			// Mock Get
			mockStore.Mock.On("Get", ctx, tt.key).Return(tt.getMockVal, tt.getMockErr)

			// Mock Set only when relevant
			if tt.expectedSetCalled {
				expectedVal, _ := tt.defaultFn()
				mockStore.Mock.On("Set", ctx, tt.key, expectedVal, ttl).Return(tt.setMockErr)
			}

			manager := &Manager{store: mockStore}

			// --- Act ---
			result, err := manager.GetOrSet(ctx, tt.key, ttl, tt.defaultFn)

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "Get", ctx, tt.key)

			if tt.expectedSetCalled {
				mockStore.Mock.AssertCalled(t, "Set", ctx, tt.key, tt.expectedVal, ttl)
			} else {
				mockStore.Mock.AssertNotCalled(t, "Set", ctx, tt.key, tt.expectedVal, ttl)
			}

			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return an error when GetOrSet fails")
				assert.EqualError(t, tt.expectedErr, err, "error returned by GetOrSet must match the expected error")
				assert.Equal(t, tt.expectedVal, result, "returned value on error must match the expected value")
				return
			}

			assert.NoError(t, err, "must not return an error when GetOrSet succeeds")
			assert.Equal(t, tt.expectedVal, result, "returned value must match the expected value")
		})
	}
}

func TestManager_Has(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name        string
		key         string
		mockVal     bool
		mockErr     error
		expectedVal bool
		expectedErr error
	}{
		{
			name:        "should return true when the key exists",
			key:         key,
			mockVal:     true,
			mockErr:     nil,
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name:        "should return false when the key does not exist",
			key:         key,
			mockVal:     false,
			mockErr:     nil,
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        "should return false when the key exists but is expired",
			key:         key,
			mockVal:     false,
			mockErr:     nil,
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name:        "should return an error when the store returns an error other than cache miss",
			key:         key,
			mockVal:     false,
			mockErr:     errors.New("some other error"),
			expectedVal: false,
			expectedErr: errors.New("some other error"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)
			mockStore.Mock.On("Has", ctx, tt.key).Return(tt.mockVal, tt.mockErr)

			manager := &Manager{store: mockStore}

			// --- Act ---
			result, err := manager.Has(ctx, tt.key)

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "Has", ctx, tt.key)
			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return an error when Has fails")
				assert.EqualError(t, tt.expectedErr, err, "error returned by Has must match the expected error")
				assert.Equal(t, tt.expectedVal, result, "result returned on error must match the expected value")
				return
			}

			assert.NoError(t, err, "must not return an error when Has succeeds")
			assert.Equal(t, tt.expectedVal, result, "result returned by Has must match the expected value")
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
		expectedErr error
	}{
		{
			name:        "should set the value successfully when the operation succeeds",
			key:         key,
			value:       value,
			ttl:         ttl,
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return an error when the Set operation fails",
			key:         key,
			value:       value,
			ttl:         ttl,
			mockErr:     errors.New("set failed"),
			expectedErr: errors.New("set failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)
			mockStore.Mock.On("Set", ctx, tt.key, value, ttl).Return(tt.mockErr)

			manager := &Manager{store: mockStore}

			// --- Act ---
			err := manager.Set(ctx, tt.key, tt.value, tt.ttl)

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "Set", ctx, tt.key, value, ttl)
			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return an error when Set fails")
				assert.EqualError(t, tt.expectedErr, err, "error returned by Set must match the expected error")
				return
			}

			assert.NoError(t, err, "must not return an error when Set succeeds")
		})
	}
}
