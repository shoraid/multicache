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

func TestManager_GetBool(t *testing.T) {
	t.Parallel()

	key := "test-bool-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectedValue bool
		expectedErr   error
	}{
		{
			name:          "should get true bool successfully",
			key:           "test-bool-key",
			mockValue:     true,
			mockReturnErr: nil,
			expectedValue: true,
			expectedErr:   nil,
		},
		{
			name:          "should get false bool successfully",
			key:           key,
			mockValue:     false,
			mockReturnErr: nil,
			expectedValue: false,
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectedValue: false,
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-bool value",
			key:           key,
			mockValue:     "not a bool",
			mockReturnErr: nil,
			expectedValue: false,
			expectedErr:   ErrTypeMismatch, // cast.ToBoolE error
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

			value, err := manager.GetBool(ctx, tt.key)

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
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		defaultFunc       func() (bool, error)
		expectedReturnVal bool
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing true value if found",
			key:               key,
			mockGetVal:        true,
			mockGetErr:        nil,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: true,
			expectedErr:       nil,
			expectSetCall:     false,
		},
		{
			name:              "should return existing false value if found",
			key:               key,
			mockGetVal:        false,
			mockGetErr:        nil,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: false,
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
			mockGetVal:        "not a bool",
			mockGetErr:        nil, // GetBool will return ErrTypeMismatch
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
			expectedReturnVal: false, // Default value for bool
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
			expectedReturnVal: false,
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

			value, err := manager.GetBoolOrSet(ctx, tt.key, ttl, tt.defaultFunc)
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
