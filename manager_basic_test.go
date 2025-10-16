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

func TestManager_Get(t *testing.T) {
	t.Parallel()

	key := "test-key"
	expectedValue := "test-value"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectedValue any
		expectedErr   error
	}{
		{
			name:          "should get value successfully",
			key:           key,
			mockValue:     expectedValue,
			mockReturnErr: nil,
			expectedValue: expectedValue,
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
			name:          "should return other error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: errors.New("some other error"),
			expectedValue: nil,
			expectedErr:   errors.New("some other error"),
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

			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.Get(ctx, tt.key)

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

func TestManager_GetOrSet(t *testing.T) {
	t.Parallel()

	key := "test-key"
	defaultValue := "default-value"
	defaultFn := func() (any, error) {
		return defaultValue, nil
	}
	errorFn := func() (any, error) {
		return nil, errors.New("default function error")
	}

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		defaultFunc       func() (any, error)
		expectedReturnVal any
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        "existing-value",
			mockGetErr:        nil,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: "existing-value",
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
			name:              "should return error if get fails with other error",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        errors.New("network error"),
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: nil,
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
					On("Set", ctx, tt.key, tt.expectedReturnVal, ttl).
					Return(tt.mockSetErr).
					Once()
			}

			value, err := manager.GetOrSet(ctx, tt.key, ttl, tt.defaultFunc)

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

func TestManager_Has(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name           string
		key            string
		mockReturnVal  bool
		mockReturnErr  error
		expectedResult bool
		expectedErr    error
	}{
		{
			name:           "should return true if key exists",
			key:            key,
			mockReturnVal:  true,
			mockReturnErr:  nil,
			expectedResult: true,
			expectedErr:    nil,
		},
		{
			name:           "should return false if key does not exist",
			key:            key,
			mockReturnVal:  false,
			mockReturnErr:  nil,
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name:           "should return false if key exists but is expired",
			key:            key,
			mockReturnVal:  false,
			mockReturnErr:  nil,
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name:           "should return error if store returns an error other than cache miss",
			key:            key,
			mockReturnVal:  false,
			mockReturnErr:  errors.New("some other error"),
			expectedResult: false,
			expectedErr:    errors.New("some other error"),
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

			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("Has", ctx, tt.key).
				Return(tt.mockReturnVal, tt.mockReturnErr).
				Once()

			result, err := manager.Has(ctx, tt.key)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectedResult, result, "expected correct result on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedResult, result, "expected correct result")
			}

			mockStore.AssertExpectations(t)
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
		mockReturn  error
		expectedErr bool
	}{
		{
			name:        "should set value successfully",
			key:         key,
			value:       value,
			ttl:         ttl,
			mockReturn:  nil,
			expectedErr: false,
		},
		{
			name:        "should return error when set fails",
			key:         key,
			value:       value,
			ttl:         ttl,
			mockReturn:  errors.New("set failed"),
			expectedErr: true,
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

			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("Set", ctx, tt.key, tt.value, tt.ttl).
				Return(tt.mockReturn).
				Once()

			err := manager.Set(ctx, tt.key, tt.value, tt.ttl)

			if tt.expectedErr {
				assert.Error(t, err, "expected error when set fails")
				assert.EqualError(t, err, tt.mockReturn.Error(), "expected correct error message")
			} else {
				assert.NoError(t, err, "expected no error when set succeeds")
			}

			mockStore.AssertExpectations(t)
		})
	}
}
