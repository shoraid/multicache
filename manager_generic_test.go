package multicache

import (
	"context"
	"errors"
	"testing"
	"time"

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
				stores: map[string]Store{"default": mockStore},
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
	valueToSet := "new-value"

	tests := []struct {
		name              string
		key               string
		mockReturnVal     any
		mockReturnErr     error
		expectedReturnVal any
		expectedErr       error
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockReturnVal:     "existing-value",
			mockReturnErr:     nil,
			expectedReturnVal: "existing-value",
			expectedErr:       nil,
		},
		{
			name:              "should set and return new value if not found",
			key:               key,
			mockReturnVal:     valueToSet,
			mockReturnErr:     nil,
			expectedReturnVal: valueToSet,
			expectedErr:       nil,
		},
		{
			name:              "should return error when store returns error",
			key:               key,
			mockReturnVal:     nil,
			mockReturnErr:     errors.New("store error"),
			expectedReturnVal: nil,
			expectedErr:       errors.New("store error"),
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

			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("GetOrSet", ctx, tt.key, ttl, valueToSet).
				Return(tt.mockReturnVal, tt.mockReturnErr).
				Once()

			value, err := manager.GetOrSet(ctx, tt.key, ttl, valueToSet)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "expected correct error message")
				assert.Nil(t, value, "expected nil value on error")
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
				stores: map[string]Store{"default": mockStore},
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
				stores: map[string]Store{"default": mockStore},
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
