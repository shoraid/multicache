package multicache

import (
	"context"
	"errors"
	"testing"

	"github.com/shoraid/multicache/contract"
	multicachemock "github.com/shoraid/multicache/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestManager_Clear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockReturn  error
		expectedErr bool
	}{
		{
			name:        "should clear successfully",
			mockReturn:  nil,
			expectedErr: false,
		},
		{
			name:        "should return error when clear fails",
			mockReturn:  errors.New("clear failed"),
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
				On("Clear", ctx).
				Return(tt.mockReturn).
				Once()

			err := manager.Clear(ctx)

			if tt.expectedErr {
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
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name        string
		key         string
		mockReturn  error
		expectedErr bool
	}{
		{
			name:        "should delete key successfully",
			key:         key,
			mockReturn:  nil,
			expectedErr: false,
		},
		{
			name:        "should return error when delete fails",
			key:         key,
			mockReturn:  errors.New("delete failed"),
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
				On("Delete", ctx, tt.key).
				Return(tt.mockReturn).
				Once()

			err := manager.Delete(ctx, tt.key)

			if tt.expectedErr {
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
	t.Parallel()

	pattern := "user:*"

	tests := []struct {
		name        string
		pattern     string
		mockReturn  error
		expectedErr bool
	}{
		{
			name:        "should delete by pattern successfully",
			pattern:     pattern,
			mockReturn:  nil,
			expectedErr: false,
		},
		{
			name:        "should return error when delete by pattern fails",
			pattern:     pattern,
			mockReturn:  errors.New("delete by pattern failed"),
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

			mockStore.ExpectedCalls = nil // reset calls between cases
			mockStore.On("DeleteByPattern", ctx, tt.pattern).Return(tt.mockReturn).Once()

			err := manager.DeleteByPattern(ctx, tt.pattern)

			if tt.expectedErr {
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
	t.Parallel()

	keys := []string{"key1", "key2", "key3"}

	tests := []struct {
		name        string
		keys        []string
		mockReturn  error
		expectedErr bool
	}{
		{
			name:        "should delete many keys successfully",
			keys:        keys,
			mockReturn:  nil,
			expectedErr: false,
		},
		{
			name:        "should return error when delete many fails",
			keys:        keys,
			mockReturn:  errors.New("delete many failed"),
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
				On("DeleteMany", ctx, tt.keys).
				Return(tt.mockReturn).
				Once()

			err := manager.DeleteMany(ctx, tt.keys...)

			if tt.expectedErr {
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
	t.Parallel()

	patterns := []string{"user:*", "product:*"}

	tests := []struct {
		name        string
		patterns    []string
		mockReturns []error // one error per pattern deletion
		expectedErr bool
	}{
		{
			name:        "should delete many by pattern successfully",
			patterns:    patterns,
			mockReturns: []error{nil, nil},
			expectedErr: false,
		},
		{
			name:        "should return error if one deletion fails",
			patterns:    patterns,
			mockReturns: []error{errors.New("delete user pattern failed"), nil},
			expectedErr: true,
		},
		{
			name:        "should return error if all deletions fail",
			patterns:    patterns,
			mockReturns: []error{errors.New("delete user pattern failed"), errors.New("delete product pattern failed")},
			expectedErr: true,
		},
		{
			name:        "should handle empty patterns list",
			patterns:    []string{},
			mockReturns: []error{},
			expectedErr: false,
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

			if tt.expectedErr {
				assert.Error(t, err, "expected error when delete many by pattern fails")
			} else {
				assert.NoError(t, err, "expected no error when delete many by pattern succeeds")
			}

			mockStore.AssertExpectations(t)
		})
	}
}
