package multicache

import (
	"context"
	"errors"
	"testing"
	"time"

	multicachemock "github.com/shoraid/multicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestManager_GetInt(t *testing.T) {
	t.Parallel()

	key := "test-int-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectedValue int
		expectedErr   error
	}{
		{
			name:          "should get int successfully",
			key:           key,
			mockValue:     123,
			mockReturnErr: nil,
			expectedValue: 123,
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectedValue: 0,
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-int value",
			key:           key,
			mockValue:     "not an int",
			mockReturnErr: nil,
			expectedValue: 0,
			expectedErr:   ErrTypeMismatch, // cast.ToIntE error
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

			value, err := manager.GetInt(ctx, tt.key)

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

func TestManager_GetInt64(t *testing.T) {
	t.Parallel()

	key := "test-int64-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectedValue int64
		expectedErr   error
	}{
		{
			name:          "should get int64 successfully",
			key:           key,
			mockValue:     int64(123456789012345),
			mockReturnErr: nil,
			expectedValue: int64(123456789012345),
			expectedErr:   nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectedValue: 0,
			expectedErr:   ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-int64 value",
			key:           key,
			mockValue:     "not an int64",
			mockReturnErr: nil,
			expectedValue: 0,
			expectedErr:   ErrTypeMismatch, // cast.ToInt64E error
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

			value, err := manager.GetInt64(ctx, tt.key)

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

func TestManager_GetInts(t *testing.T) {
	t.Parallel()

	key := "test-ints-key"

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectedValue []int
		expectedErr   error
	}{
		{
			name:          "should get []int successfully from []int",
			key:           key,
			mockValue:     []int{1, 2, 3},
			mockReturnErr: nil,
			expectedValue: []int{1, 2, 3},
			expectedErr:   nil,
		},
		{
			name:          "should get []int successfully from JSON string",
			key:           key,
			mockValue:     "[4,5,6]",
			mockReturnErr: nil,
			expectedValue: []int{4, 5, 6},
			expectedErr:   nil,
		},
		{
			name:          "should get []int successfully from JSON []byte",
			key:           key,
			mockValue:     []byte("[7,8,9]"),
			mockReturnErr: nil,
			expectedValue: []int{7, 8, 9},
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
			name:          "should return type mismatch error for non-[]int value",
			key:           key,
			mockValue:     "not an int array",
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
		{
			name:          "should return type mismatch error for non-[]int value",
			key:           key,
			mockValue:     true,
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
				stores: map[string]Store{"default": mockStore},
				store:  mockStore,
			}

			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetInts(ctx, tt.key)

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

func TestManager_GetIntOrSet(t *testing.T) {
	t.Parallel()

	key := "test-int-key"
	defaultValue := 99

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		expectedReturnVal int
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        123,
			mockGetErr:        nil,
			mockSetErr:        nil,
			expectedReturnVal: 123,
			expectedErr:       nil,
			expectSetCall:     false,
		},
		{
			name:              "should set and return default if cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        nil,
			expectedReturnVal: defaultValue,
			expectedErr:       nil,
			expectSetCall:     true,
		},
		{
			name:              "should set and return default if type mismatch on get",
			key:               key,
			mockGetVal:        "not an int",
			mockGetErr:        nil, // GetInt will return ErrTypeMismatch
			mockSetErr:        nil,
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
			expectedReturnVal: 0, // Default value for int
			expectedErr:       errors.New("network error"),
			expectSetCall:     false},
		{
			name:              "should return error if set fails after type mismatch",
			key:               key,
			mockGetVal:        "not an int",
			mockGetErr:        nil, // GetInt will return ErrTypeMismatch
			mockSetErr:        errors.New("set operation failed"),
			expectedReturnVal: defaultValue,
			expectedErr:       errors.New("set operation failed"),
			expectSetCall:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ttl := 5 * time.Minute
			ctx := context.Background()
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

			value, err := manager.GetIntOrSet(ctx, tt.key, ttl, defaultValue)

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

func TestManager_GetInt64OrSet(t *testing.T) {
	t.Parallel()

	key := "test-int64-key"
	defaultValue := int64(999999999999)

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		expectedReturnVal int64
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        int64(123456789012345),
			mockGetErr:        nil,
			mockSetErr:        nil,
			expectedReturnVal: int64(123456789012345),
			expectedErr:       nil,
			expectSetCall:     false,
		},
		{
			name:              "should set and return default if cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        nil,
			expectedReturnVal: defaultValue,
			expectedErr:       nil,
			expectSetCall:     true,
		},
		{
			name:              "should set and return default if type mismatch on get",
			key:               key,
			mockGetVal:        "not an int64",
			mockGetErr:        nil, // GetInt64 will return ErrTypeMismatch
			mockSetErr:        nil,
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
			expectedReturnVal: 0, // Default value for int64
			expectedErr:       errors.New("network error"),
			expectSetCall:     false,
		},
		{
			name:              "should return error if set fails after type mismatch",
			key:               key,
			mockGetVal:        "not an int64",
			mockGetErr:        nil, // GetInt64 will return ErrTypeMismatch
			mockSetErr:        errors.New("set operation failed"),
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

			value, err := manager.GetInt64OrSet(ctx, tt.key, ttl, defaultValue)

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

func TestManager_GetIntsOrSet(t *testing.T) {
	t.Parallel()

	key := "test-ints-key"
	defaultValue := []int{10, 20, 30}

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		expectedReturnVal []int
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        []int{1, 2, 3},
			mockGetErr:        nil,
			mockSetErr:        nil,
			expectedReturnVal: []int{1, 2, 3},
			expectedErr:       nil,
			expectSetCall:     false,
		},
		{
			name:              "should set and return default if cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        nil,
			expectedReturnVal: defaultValue,
			expectedErr:       nil,
			expectSetCall:     true,
		},
		{
			name:              "should set and return default if type mismatch on get",
			key:               key,
			mockGetVal:        "not an int array",
			mockGetErr:        nil, // GetInts will return ErrTypeMismatch
			mockSetErr:        nil,
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
			expectedReturnVal: nil, // Default
			expectedErr:       errors.New("network error"),
			expectSetCall:     false,
		},
		{
			name:              "should return error if set fails after type mismatch",
			key:               key,
			mockGetVal:        "not an int array",
			mockGetErr:        nil, // GetInts will return ErrTypeMismatch
			mockSetErr:        errors.New("set operation failed"),
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

			value, err := manager.GetIntsOrSet(ctx, tt.key, ttl, defaultValue)

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
