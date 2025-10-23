package omnicache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shoraid/omnicache/internal/assert"
	omnicachemock "github.com/shoraid/omnicache/mock"
)

type TestStruct struct {
	Name string
	Age  int
}

func TestGenericManager_Get(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name        string
		mockVal     any
		mockErr     error
		expectedVal any
		expectedErr error
	}{
		{
			name:        "should successfully get a string value",
			mockVal:     "hello",
			mockErr:     nil,
			expectedVal: "hello",
			expectedErr: nil,
		},
		{
			name:        "should successfully get an int value",
			mockVal:     123,
			mockErr:     nil,
			expectedVal: 123,
			expectedErr: nil,
		},
		{
			name:        "should successfully get a struct value",
			mockVal:     TestStruct{Name: "John", Age: 30},
			mockErr:     nil,
			expectedVal: TestStruct{Name: "John", Age: 30},
			expectedErr: nil,
		},
		{
			name:        "should return ErrCacheMiss when key not found",
			mockVal:     nil,
			mockErr:     ErrCacheMiss,
			expectedVal: "", // Zero value for string
			expectedErr: ErrCacheMiss,
		},
		{
			name:        "should return underlying error if not cache miss or type mismatch",
			mockVal:     nil,
			mockErr:     errors.New("some other error"),
			expectedVal: "",
			expectedErr: errors.New("some other error"),
		},
		{
			name:        "should return ErrTypeMismatch when convertAnyToType fails",
			mockVal:     TestStruct{Name: "John", Age: 30},
			mockErr:     nil, // Get succeeded
			expectedVal: 0,   // zero value for int
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "should return ErrTypeMismatch when value is nil but no error",
			mockVal:     nil,
			mockErr:     nil,
			expectedVal: "",
			expectedErr: ErrTypeMismatch,
		},
		{
			name:        "should return ErrTypeMismatch when cached bool mismatches target type",
			mockVal:     true,
			mockErr:     nil,
			expectedVal: 0,
			expectedErr: ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)
			mockStore.Mock.On("Get", ctx, key).Return(tt.mockVal, tt.mockErr)

			manager := &Manager{store: mockStore}

			var (
				result any
				err    error
			)

			// --- Act ---
			switch expected := tt.expectedVal.(type) {
			case string:
				g := G[string](manager)
				result, err = g.Get(ctx, key)
			case int:
				g := G[int](manager)
				result, err = g.Get(ctx, key)
			case bool:
				g := G[bool](manager)
				result, err = g.Get(ctx, key)
			case []byte:
				g := G[[]byte](manager)
				result, err = g.Get(ctx, key)
			case TestStruct:
				g := G[TestStruct](manager)
				result, err = g.Get(ctx, key)
			default:
				t.Fatalf("unsupported expected type: %T", expected)
			}

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "Get", ctx, key)
			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return an error when Get fails")
				assert.EqualError(t, tt.expectedErr, err, "error returned by Get must match the expected error")
				assert.Equal(t, tt.expectedVal, result, "result on error must be the zero value of the target type")
				return
			}

			assert.NoError(t, err, "must not return an error when Get succeeds")
			assert.Equal(t, tt.expectedVal, result, "value returned by Get must match the expected value")
		})
	}
}

func TestGenericManager_GetOrSet(t *testing.T) {
	t.Parallel()

	key := "test-key"
	ttl := 5 * time.Minute
	defaultString := "default-string"
	defaultInt := 123
	defaultStruct := TestStruct{Name: "Default", Age: 99}

	stringDefaultFn := func() (string, error) { return defaultString, nil }
	intDefaultFn := func() (int, error) { return defaultInt, nil }
	structDefaultFn := func() (TestStruct, error) { return defaultStruct, nil }
	errorDefaultFn := func() (string, error) { return "", errors.New("default fn error") }

	tests := []struct {
		name              string
		getMockVal        any
		getMockErr        error
		setMockErr        error
		defaultFn         any // Can be func() (T, error) for different T
		expectedVal       any
		expectedErr       error
		expectedSetCalled bool
	}{
		// --- Cache Hit Scenarios ---
		{
			name:        "string: should return existing value when cache hit",
			getMockVal:  "cached-string",
			getMockErr:  nil,
			defaultFn:   stringDefaultFn,
			expectedVal: "cached-string",
			expectedErr: nil,
		},
		{
			name:        "int: should return existing value when cache hit",
			getMockVal:  456,
			getMockErr:  nil,
			defaultFn:   intDefaultFn,
			expectedVal: 456,
		},
		{
			name:        "struct: should return existing value when cache hit",
			getMockVal:  TestStruct{Name: "Cached", Age: 10},
			getMockErr:  nil,
			defaultFn:   structDefaultFn,
			expectedVal: TestStruct{Name: "Cached", Age: 10},
		},

		// --- Cache Miss Scenarios ---
		{
			name:              "string: should set and return default value when cache miss",
			getMockVal:        nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        nil,
			defaultFn:         stringDefaultFn,
			expectedVal:       defaultString,
			expectedErr:       nil,
			expectedSetCalled: true,
		},
		{
			name:              "int: should set and return default value when cache miss",
			getMockVal:        nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        nil,
			defaultFn:         intDefaultFn,
			expectedVal:       defaultInt,
			expectedErr:       nil,
			expectedSetCalled: true,
		},
		{
			name:              "struct: should set and return default value when cache miss",
			getMockVal:        nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        nil,
			defaultFn:         structDefaultFn,
			expectedVal:       defaultStruct,
			expectedErr:       nil,
			expectedSetCalled: true,
		},

		// --- Error Scenarios ---
		{
			name:        "should return error when Get returns a non-cache-miss error",
			getMockVal:  nil,
			getMockErr:  errors.New("get error"),
			defaultFn:   stringDefaultFn,
			expectedVal: "",
			expectedErr: errors.New("get error"),
		},
		{
			name:        "should return error when defaultFn fails after cache miss",
			getMockVal:  nil,
			getMockErr:  ErrCacheMiss,
			setMockErr:  nil,
			defaultFn:   errorDefaultFn,
			expectedVal: "",
			expectedErr: errors.New("default fn error"),
		},
		{
			name:              "should return default value but also return error when Set fails after cache miss",
			getMockVal:        nil,
			getMockErr:        ErrCacheMiss,
			setMockErr:        errors.New("set error"),
			defaultFn:         stringDefaultFn,
			expectedVal:       defaultString,
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
			mockStore.Mock.On("Get", ctx, key).Return(tt.getMockVal, tt.getMockErr)

			// Mock Set only when relevant
			if tt.expectedSetCalled {
				mockStore.Mock.On("Set", ctx, key, tt.expectedVal, ttl).Return(tt.setMockErr)
			}

			manager := &Manager{store: mockStore}

			var (
				result any
				err    error
			)

			// --- Act ---
			switch expected := tt.expectedVal.(type) {
			case string:
				g := G[string](manager)
				result, err = g.GetOrSet(ctx, key, ttl, tt.defaultFn.(func() (string, error)))
			case int:
				g := G[int](manager)
				result, err = g.GetOrSet(ctx, key, ttl, tt.defaultFn.(func() (int, error)))
			case TestStruct:
				g := G[TestStruct](manager)
				result, err = g.GetOrSet(ctx, key, ttl, tt.defaultFn.(func() (TestStruct, error)))
			default:
				t.Fatalf("unsupported expected type: %T", expected)
			}

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "Get", ctx, key)

			if tt.expectedSetCalled {
				mockStore.Mock.AssertCalled(t, "Set", ctx, key, tt.expectedVal, ttl)
			} else {
				mockStore.Mock.AssertNotCalled(t, "Set", ctx, key, tt.expectedVal, ttl)
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
