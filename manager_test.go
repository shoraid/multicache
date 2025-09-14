package multicache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestManager_NewManager(t *testing.T) {
	mockStore := new(MockStore)

	tests := []struct {
		name         string
		defaultStore string
		stores       map[string]Store
		expectErr    error
		expectNil    bool
	}{
		{
			name:         "should return manager when default store exists",
			defaultStore: "default",
			stores:       map[string]Store{"default": mockStore},
			expectErr:    nil,
			expectNil:    false,
		},
		{
			name:         "should return error when default store does not exist",
			defaultStore: "missing",
			stores:       map[string]Store{"default": mockStore},
			expectErr:    ErrInvalidDefaultStore,
			expectNil:    true,
		},
		{
			name:         "should return error when stores map is empty",
			defaultStore: "default",
			stores:       map[string]Store{},
			expectErr:    ErrInvalidDefaultStore,
			expectNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewManager(tt.defaultStore, tt.stores)

			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr, "expected error to match")
				assert.Nil(t, mgr, "expected manager to be nil")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.NotNil(t, mgr, "expected manager to be non-nil")
			}
		})
	}
}

func TestManager_Store(t *testing.T) {
	mockDefault := new(MockStore)
	mockOther := new(MockStore)

	stores := map[string]Store{
		"default": mockDefault,
		"other":   mockOther,
	}

	manager, err := NewManager("default", stores)
	assert.NoError(t, err, "expected no error creating manager")

	tests := []struct {
		name       string
		alias      string
		expectNil  bool
		expectSame bool
	}{
		{
			name:       "should return manager with existing alias store",
			alias:      "other",
			expectNil:  false,
			expectSame: false,
		},
		{
			name:       "should return manager with nil store when alias does not exist",
			alias:      "missing",
			expectNil:  true,
			expectSame: false,
		},
		{
			name:       "should return manager with same store when alias is default",
			alias:      "default",
			expectNil:  false,
			expectSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newMgr := manager.Store(tt.alias)

			impl, ok := newMgr.(*managerImpl)
			assert.True(t, ok, "expected returned Manager to be *managerImpl")

			if tt.expectNil {
				assert.Nil(t, impl.Store(tt.alias).(*managerImpl).store, "expected store to be nil")
			} else {
				assert.NotNil(t, impl.store, "expected store to be non-nil")
			}

			if tt.expectSame {
				assert.Equal(t, manager.(*managerImpl).store, impl.store, "expected same store reference")
			} else if !tt.expectNil {
				assert.NotSame(t, manager.(*managerImpl).store, impl.store, "expected different store reference")
			}
		})
	}
}

func TestManager_Clear(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name       string
		mockReturn error
		expectErr  bool
	}{
		{
			name:       "should clear successfully",
			mockReturn: nil,
			expectErr:  false,
		},
		{
			name:       "should return error when clear fails",
			mockReturn: errors.New("clear failed"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("Clear", ctx).
				Return(tt.mockReturn).
				Once()

			err := manager.Clear(ctx)

			if tt.expectErr {
				assert.Error(t, err, "expected error when clear fails")
				assert.EqualError(t, err, tt.mockReturn.Error(), "expected correct error message")
			} else {
				assert.NoError(t, err, "expected no error when clear succeeds")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_Delete(t *testing.T) {
	ctx := context.Background()
	key := "test-key"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name       string
		key        string
		mockReturn error
		expectErr  bool
	}{
		{
			name:       "should delete key successfully",
			key:        key,
			mockReturn: nil,
			expectErr:  false,
		},
		{
			name:       "should return error when delete fails",
			key:        key,
			mockReturn: errors.New("delete failed"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("Delete", ctx, tt.key).
				Return(tt.mockReturn).
				Once()

			err := manager.Delete(ctx, tt.key)

			if tt.expectErr {
				assert.Error(t, err, "expected error when delete fails")
				assert.EqualError(t, err, tt.mockReturn.Error(), "expected correct error message")
			} else {
				assert.NoError(t, err, "expected no error when delete succeeds")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_DeleteByPattern(t *testing.T) {
	ctx := context.Background()
	pattern := "user:*"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name       string
		pattern    string
		mockReturn error
		expectErr  bool
	}{
		{
			name:       "should delete by pattern successfully",
			pattern:    pattern,
			mockReturn: nil,
			expectErr:  false,
		},
		{
			name:       "should return error when delete by pattern fails",
			pattern:    pattern,
			mockReturn: errors.New("delete by pattern failed"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // reset calls between cases
			mockStore.On("DeleteByPattern", ctx, tt.pattern).Return(tt.mockReturn).Once()

			err := manager.DeleteByPattern(ctx, tt.pattern)

			if tt.expectErr {
				assert.Error(t, err, "expected error when delete by pattern fails")
				assert.EqualError(t, err, tt.mockReturn.Error(), "expected correct error message")
			} else {
				assert.NoError(t, err, "expected no error when delete by pattern succeeds")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_DeleteMany(t *testing.T) {
	ctx := context.Background()
	keys := []string{"key1", "key2", "key3"}
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name       string
		keys       []string
		mockReturn error
		expectErr  bool
	}{
		{
			name:       "should delete many keys successfully",
			keys:       keys,
			mockReturn: nil,
			expectErr:  false,
		},
		{
			name:       "should return error when delete many fails",
			keys:       keys,
			mockReturn: errors.New("delete many failed"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("DeleteMany", ctx, tt.keys).
				Return(tt.mockReturn).
				Once()

			err := manager.DeleteMany(ctx, tt.keys...)

			if tt.expectErr {
				assert.Error(t, err, "expected error when delete many fails")
				assert.EqualError(t, err, tt.mockReturn.Error(), "expected correct error message")
			} else {
				assert.NoError(t, err, "expected no error when delete many succeeds")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_DeleteManyByPattern(t *testing.T) {
	ctx := context.Background()
	patterns := []string{"user:*", "product:*"}
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name        string
		patterns    []string
		mockReturns []error // one error per pattern deletion
		expectErr   bool
	}{
		{
			name:        "should delete many by pattern successfully",
			patterns:    patterns,
			mockReturns: []error{nil, nil},
			expectErr:   false,
		},
		{
			name:        "should return error if one deletion fails",
			patterns:    patterns,
			mockReturns: []error{errors.New("delete user pattern failed"), nil},
			expectErr:   true,
		},
		{
			name:        "should return error if all deletions fail",
			patterns:    patterns,
			mockReturns: []error{errors.New("delete user pattern failed"), errors.New("delete product pattern failed")},
			expectErr:   true,
		},
		{
			name:        "should handle empty patterns list",
			patterns:    []string{},
			mockReturns: []error{},
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // reset calls for isolation

			for i, pattern := range tt.patterns {
				var returnErr error
				if i < len(tt.mockReturns) {
					returnErr = tt.mockReturns[i]
				}
				mockStore.
					On("DeleteByPattern", mock.Anything, pattern).
					Return(returnErr).
					Once()
			}

			err := manager.DeleteManyByPattern(ctx, tt.patterns...)

			if tt.expectErr {
				assert.Error(t, err, "expected error when delete many by pattern fails")
			} else {
				assert.NoError(t, err, "expected no error when delete many by pattern succeeds")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_Get(t *testing.T) {
	ctx := context.Background()
	key := "test-key"
	expectedValue := "test-value"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectValue   any
		expectErr     error
	}{
		{
			name:          "should get value successfully",
			key:           key,
			mockValue:     expectedValue,
			mockReturnErr: nil,
			expectValue:   expectedValue,
			expectErr:     nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectValue:   nil,
			expectErr:     ErrCacheMiss,
		},
		{
			name:          "should return other error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: errors.New("some other error"),
			expectValue:   nil,
			expectErr:     errors.New("some other error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.Get(ctx, tt.key)

			if tt.expectErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectErr.Error(), err.Error(), "expected correct error message")
				assert.Nil(t, value, "expected nil value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetBool(t *testing.T) {
	ctx := context.Background()
	key := "test-bool-key"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectValue   bool
		expectErr     error
	}{
		{
			name:          "should get true bool successfully",
			key:           key,
			mockValue:     true,
			mockReturnErr: nil,
			expectValue:   true,
			expectErr:     nil,
		},
		{
			name:          "should get false bool successfully",
			key:           key,
			mockValue:     false,
			mockReturnErr: nil,
			expectValue:   false,
			expectErr:     nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectValue:   false,
			expectErr:     ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-bool value",
			key:           key,
			mockValue:     "not a bool",
			mockReturnErr: nil,
			expectValue:   false,
			expectErr:     ErrTypeMismatch, // cast.ToBoolE error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetBool(ctx, tt.key)

			if tt.expectErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectValue, value, "expected default value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetBoolOrSet(t *testing.T) {
	ctx := context.Background()
	key := "test-bool-key"
	ttl := 5 * time.Minute
	defaultValue := true
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
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
			expectedReturnVal: false, // Default value for bool
			expectedErr:       errors.New("network error"),
			expectSetCall:     false,
		},
		{
			name:              "should return error if set fails after cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        errors.New("set operation failed"),
			expectedReturnVal: defaultValue,
			expectedErr:       errors.New("set operation failed"),
			expectSetCall:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			value, err := manager.GetBoolOrSet(ctx, tt.key, ttl, defaultValue)

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

func TestManager_GetInt(t *testing.T) {
	ctx := context.Background()
	key := "test-int-key"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectValue   int
		expectErr     error
	}{
		{
			name:          "should get int successfully",
			key:           key,
			mockValue:     123,
			mockReturnErr: nil,
			expectValue:   123,
			expectErr:     nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectValue:   0,
			expectErr:     ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-int value",
			key:           key,
			mockValue:     "not an int",
			mockReturnErr: nil,
			expectValue:   0,
			expectErr:     ErrTypeMismatch, // cast.ToIntE error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetInt(ctx, tt.key)

			if tt.expectErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectValue, value, "expected default value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetIntOrSet(t *testing.T) {
	ctx := context.Background()
	key := "test-int-key"
	ttl := 5 * time.Minute
	defaultValue := 99
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

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

func TestManager_GetInt64(t *testing.T) {
	ctx := context.Background()
	key := "test-int64-key"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectValue   int64
		expectErr     error
	}{
		{
			name:          "should get int64 successfully",
			key:           key,
			mockValue:     int64(123456789012345),
			mockReturnErr: nil,
			expectValue:   int64(123456789012345),
			expectErr:     nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectValue:   0,
			expectErr:     ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-int64 value",
			key:           key,
			mockValue:     "not an int64",
			mockReturnErr: nil,
			expectValue:   0,
			expectErr:     ErrTypeMismatch, // cast.ToInt64E error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetInt64(ctx, tt.key)

			if tt.expectErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectValue, value, "expected default value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetInt64OrSet(t *testing.T) {
	ctx := context.Background()
	key := "test-int64-key"
	ttl := 5 * time.Minute
	defaultValue := int64(999999999999)
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

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

func TestManager_GetInts(t *testing.T) {
	ctx := context.Background()
	key := "test-ints-key"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectValue   []int
		expectErr     error
	}{
		{
			name:          "should get []int successfully from []int",
			key:           key,
			mockValue:     []int{1, 2, 3},
			mockReturnErr: nil,
			expectValue:   []int{1, 2, 3},
			expectErr:     nil,
		},
		{
			name:          "should get []int successfully from JSON string",
			key:           key,
			mockValue:     "[4,5,6]",
			mockReturnErr: nil,
			expectValue:   []int{4, 5, 6},
			expectErr:     nil,
		},
		{
			name:          "should get []int successfully from JSON []byte",
			key:           key,
			mockValue:     []byte("[7,8,9]"),
			mockReturnErr: nil,
			expectValue:   []int{7, 8, 9},
			expectErr:     nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectValue:   nil,
			expectErr:     ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-[]int value",
			key:           key,
			mockValue:     "not an int array",
			mockReturnErr: nil,
			expectValue:   nil,
			expectErr:     ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch error for invalid JSON string",
			key:           key,
			mockValue:     "invalid json",
			mockReturnErr: nil,
			expectValue:   nil,
			expectErr:     ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch error for invalid JSON []byte",
			key:           key,
			mockValue:     []byte("invalid json"),
			mockReturnErr: nil,
			expectValue:   nil,
			expectErr:     ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetInts(ctx, tt.key)

			if tt.expectErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectErr.Error(), err.Error(), "expected correct error message")
				assert.Nil(t, value, "expected nil value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetIntsOrSet(t *testing.T) {
	ctx := context.Background()
	key := "test-ints-key"
	ttl := 5 * time.Minute
	defaultValue := []int{10, 20, 30}
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

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

func TestManager_GetString(t *testing.T) {
	ctx := context.Background()
	key := "test-string-key"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectValue   string
		expectErr     error
	}{
		{
			name:          "should get string successfully",
			key:           key,
			mockValue:     "hello world",
			mockReturnErr: nil,
			expectValue:   "hello world",
			expectErr:     nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectValue:   "",
			expectErr:     ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-string value",
			key:           key,
			mockValue:     123,
			mockReturnErr: nil,
			expectValue:   "",
			expectErr:     ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetString(ctx, tt.key)

			if tt.expectErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectValue, value, "expected default value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetStringOrSet(t *testing.T) {
	ctx := context.Background()
	key := "test-string-key"
	ttl := 5 * time.Minute
	defaultValue := "default-string"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		expectedReturnVal string
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        "existing-string",
			mockGetErr:        nil,
			mockSetErr:        nil,
			expectedReturnVal: "existing-string",
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
			mockGetVal:        123, // GetString will return ErrTypeMismatch
			mockGetErr:        nil,
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
			expectedReturnVal: "", // Default value for string
			expectedErr:       errors.New("network error"),
			expectSetCall:     false},
		{
			name:              "should return error if set fails after cache miss",
			key:               key,
			mockGetVal:        nil,
			mockGetErr:        ErrCacheMiss,
			mockSetErr:        errors.New("set operation failed"),
			expectedReturnVal: defaultValue,
			expectedErr:       errors.New("set operation failed"),
			expectSetCall:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			value, err := manager.GetStringOrSet(ctx, tt.key, ttl, defaultValue)

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

func TestManager_GetStrings(t *testing.T) {
	ctx := context.Background()
	key := "test-strings-key"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockReturnErr error
		expectValue   []string
		expectErr     error
	}{
		{
			name:          "should get []string successfully from []string",
			key:           key,
			mockValue:     []string{"a", "b", "c"},
			mockReturnErr: nil,
			expectValue:   []string{"a", "b", "c"},
			expectErr:     nil,
		},
		{
			name:          "should get []string successfully from JSON string",
			key:           key,
			mockValue:     `["d","e","f"]`,
			mockReturnErr: nil,
			expectValue:   []string{"d", "e", "f"},
			expectErr:     nil,
		},
		{
			name:          "should get []string successfully from JSON []byte",
			key:           key,
			mockValue:     []byte(`["g","h","i"]`),
			mockReturnErr: nil,
			expectValue:   []string{"g", "h", "i"},
			expectErr:     nil,
		},
		{
			name:          "should return cache miss error",
			key:           key,
			mockValue:     nil,
			mockReturnErr: ErrCacheMiss,
			expectValue:   nil,
			expectErr:     ErrCacheMiss,
		},
		{
			name:          "should return type mismatch error for non-[]string value",
			key:           key,
			mockValue:     123,
			mockReturnErr: nil,
			expectValue:   nil,
			expectErr:     ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch error for invalid JSON string",
			key:           key,
			mockValue:     "invalid json",
			mockReturnErr: nil,
			expectValue:   nil,
			expectErr:     ErrTypeMismatch,
		},
		{
			name:          "should return type mismatch error for invalid JSON []byte",
			key:           key,
			mockValue:     []byte("invalid json"),
			mockReturnErr: nil,
			expectValue:   nil,
			expectErr:     ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil
			mockStore.
				On("Get", ctx, tt.key).
				Return(tt.mockValue, tt.mockReturnErr).
				Once()

			value, err := manager.GetStrings(ctx, tt.key)

			if tt.expectErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectErr.Error(), err.Error(), "expected correct error message")
				assert.Nil(t, value, "expected nil value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectValue, value, "expected correct value")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_GetStringsOrSet(t *testing.T) {
	ctx := context.Background()
	key := "test-strings-key"
	ttl := 5 * time.Minute
	defaultValue := []string{"default1", "default2"}
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name              string
		key               string
		mockGetVal        any
		mockGetErr        error
		mockSetErr        error
		expectedReturnVal []string
		expectedErr       error
		expectSetCall     bool
	}{
		{
			name:              "should return existing value if found",
			key:               key,
			mockGetVal:        []string{"existing1", "existing2"},
			mockGetErr:        nil,
			mockSetErr:        nil,
			expectedReturnVal: []string{"existing1", "existing2"},
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
			mockGetVal:        123, // GetStrings will return ErrTypeMismatch
			mockGetErr:        nil,
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
			mockGetVal:        123, // GetStrings will return ErrTypeMismatch
			mockGetErr:        nil,
			mockSetErr:        errors.New("set operation failed"),
			expectedReturnVal: defaultValue,
			expectedErr:       errors.New("set operation failed"),
			expectSetCall:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			value, err := manager.GetStringsOrSet(ctx, tt.key, ttl, defaultValue)

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

func TestManager_GetOrSet(t *testing.T) {
	ctx := context.Background()
	key := "test-key"
	valueToSet := "new-value"
	ttl := 5 * time.Minute
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

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
	ctx := context.Background()
	key := "test-key"
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name          string
		key           string
		mockReturnVal bool
		mockReturnErr error
		expectResult  bool
		expectErr     error
	}{
		{
			name:          "should return true if key exists",
			key:           key,
			mockReturnVal: true,
			mockReturnErr: nil,
			expectResult:  true,
			expectErr:     nil,
		},
		{
			name:          "should return false if key does not exist",
			key:           key,
			mockReturnVal: false,
			mockReturnErr: nil,
			expectResult:  false,
			expectErr:     nil,
		},
		{
			name:          "should return false if key exists but is expired",
			key:           key,
			mockReturnVal: false,
			mockReturnErr: nil,
			expectResult:  false,
			expectErr:     nil,
		},
		{
			name:          "should return error if store returns an error other than cache miss",
			key:           key,
			mockReturnVal: false,
			mockReturnErr: errors.New("some other error"),
			expectResult:  false,
			expectErr:     errors.New("some other error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("Has", ctx, tt.key).
				Return(tt.mockReturnVal, tt.mockReturnErr).
				Once()

			result, err := manager.Has(ctx, tt.key)

			if tt.expectErr != nil {
				assert.Error(t, err, "expected error")
				assert.Equal(t, tt.expectErr.Error(), err.Error(), "expected correct error message")
				assert.Equal(t, tt.expectResult, result, "expected correct result on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectResult, result, "expected correct result")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestManager_Set(t *testing.T) {
	ctx := context.Background()
	key := "test-key"
	value := "test-value"
	ttl := 5 * time.Minute
	mockStore := new(MockStore)

	manager := &managerImpl{
		stores: map[string]Store{"default": mockStore},
		store:  mockStore,
	}

	tests := []struct {
		name       string
		key        string
		value      any
		ttl        time.Duration
		mockReturn error
		expectErr  bool
	}{
		{
			name:       "should set value successfully",
			key:        key,
			value:      value,
			ttl:        ttl,
			mockReturn: nil,
			expectErr:  false,
		},
		{
			name:       "should return error when set fails",
			key:        key,
			value:      value,
			ttl:        ttl,
			mockReturn: errors.New("set failed"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // reset calls for isolation
			mockStore.
				On("Set", ctx, tt.key, tt.value, tt.ttl).
				Return(tt.mockReturn).
				Once()

			err := manager.Set(ctx, tt.key, tt.value, tt.ttl)

			if tt.expectErr {
				assert.Error(t, err, "expected error when set fails")
				assert.EqualError(t, err, tt.mockReturn.Error(), "expected correct error message")
			} else {
				assert.NoError(t, err, "expected no error when set succeeds")
			}

			mockStore.AssertExpectations(t)
		})
	}
}
