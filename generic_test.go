package multicache

import (
	"context"
	"errors"
	"testing"
	"time"

	multicachemock "github.com/shoraid/multicache/mock"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	ID   int
	Name string
}

func TestGenericManager_Get(t *testing.T) {
	t.Parallel()

	key := "test-generic-key"
	expectedValue := TestStruct{ID: 1, Name: "Test"}

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectedValue TestStruct
		expectedErr   error
	}{
		{
			name:          "should get generic value successfully",
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
			expectedValue: TestStruct{},
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-generic value",
			key:           key,
			mockValue:     "not a TestStruct",
			mockReturnErr: nil,
			expectedValue: TestStruct{},
			expectedErr:   ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch when store returns nil, nil",
			key:           key,
			mockValue:     nil,
			mockReturnErr: nil,
			expectedValue: TestStruct{},
			expectedErr:   ErrTypeMismatch,
		},
		{
			name:          "should return other error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: errors.New("some other error"),
			expectedValue: TestStruct{},
			expectedErr:   errors.New("some other error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockStore := new(multicachemock.MockStore)
			manager := &managerImpl{
				stores: map[string]Store{"default": mockStore},
				store:  mockStore,
			}

			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			genericManager := G[TestStruct](manager)

			value, err := genericManager.Get(ctx, tt.key)

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

func TestGenericManager_GetOrSet(t *testing.T) {
	t.Parallel()

	key := "test-generic-key"
	defaultValue := TestStruct{ID: 2, Name: "Default"}
	defaultFn := func() (TestStruct, error) {
		return defaultValue, nil
	}
	errorFn := func() (TestStruct, error) {
		return TestStruct{}, errors.New("default function error")
	}

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		defaultFunc       func() (TestStruct, error)
		expectedReturnVal TestStruct
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        TestStruct{ID: 1, Name: "Existing"},
			mockGetErr:        nil,
			mockSetErr:        nil,
			defaultFunc:       defaultFn,
			expectedReturnVal: TestStruct{ID: 1, Name: "Existing"},
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
			mockGetVal:        "not a TestStruct",
			mockGetErr:        nil, // Get will return ErrTypeMismatch
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
			expectedReturnVal: TestStruct{},
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
			expectedReturnVal: TestStruct{},
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
				stores: map[string]Store{"default": mockStore},
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

			genericManager := G[TestStruct](manager)

			value, err := genericManager.GetOrSet(ctx, tt.key, ttl, tt.defaultFunc)
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
