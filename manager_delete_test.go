package omnicache

import (
	"context"
	"errors"
	"testing"

	"github.com/shoraid/omnicache/internal/assert"
	omnicachemock "github.com/shoraid/omnicache/mock"
)

func TestManager_Clear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "should clear cache successfully when store clears without error",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return error when store clear fails",
			mockErr:     errors.New("clear failed"),
			expectedErr: errors.New("clear failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)
			mockStore.Mock.On("Clear", ctx).Return(tt.mockErr)

			manager := &Manager{store: mockStore}

			// --- Act ---
			err := manager.Clear(ctx)

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "Clear", ctx)
			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return error when store clear fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				return
			}

			assert.NoError(t, err, "expected no error when clear succeeds")
		})
	}
}

func TestManager_Delete(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name        string
		key         string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "should delete key successfully when store delete succeeds",
			key:         key,
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return error when store delete fails",
			key:         key,
			mockErr:     errors.New("delete failed"),
			expectedErr: errors.New("delete failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)
			mockStore.Mock.On("Delete", ctx, tt.key).Return(tt.mockErr)

			manager := &Manager{store: mockStore}

			// --- Act ---
			err := manager.Delete(ctx, tt.key)

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "Delete", ctx, tt.key)
			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return error when store delete fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				return
			}

			assert.NoError(t, err, "must not return error when store delete succeeds")
		})
	}
}

func TestManager_DeleteByPattern(t *testing.T) {
	t.Parallel()

	pattern := "user:*"

	tests := []struct {
		name        string
		pattern     string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "should delete entries by pattern successfully when store operation succeeds",
			pattern:     pattern,
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return error when store delete by pattern fails",
			pattern:     pattern,
			mockErr:     errors.New("delete by pattern failed"),
			expectedErr: errors.New("delete by pattern failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)
			mockStore.Mock.On("DeleteByPattern", ctx, tt.pattern).Return(tt.mockErr)

			manager := &Manager{store: mockStore}

			// --- Act ---
			err := manager.DeleteByPattern(ctx, tt.pattern)

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "DeleteByPattern", ctx, tt.pattern)
			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return error when store delete by pattern fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				return
			}

			assert.NoError(t, err, "expected no error when delete by pattern succeeds")
		})
	}
}

func TestManager_DeleteMany(t *testing.T) {
	t.Parallel()

	keys := []string{"key1", "key2", "key3"}

	tests := []struct {
		name        string
		keys        []string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "should delete multiple keys successfully when store operation succeeds",
			keys:        keys,
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return error when store delete many fails",
			keys:        keys,
			mockErr:     errors.New("delete many failed"),
			expectedErr: errors.New("delete many failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)
			mockStore.Mock.On("DeleteMany", ctx, tt.keys).Return(tt.mockErr)

			manager := &Manager{store: mockStore}

			// --- Act ---
			err := manager.DeleteMany(ctx, tt.keys...)

			// --- Assert ---
			mockStore.Mock.AssertCalled(t, "DeleteMany", ctx, tt.keys)
			mockStore.Mock.AssertExpectations(t)

			if tt.expectedErr != nil {
				assert.Error(t, err, "must return error when store delete many fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the store error")
				return
			}

			assert.NoError(t, err, "must not return error when delete many succeeds")
		})
	}
}

func TestManager_DeleteManyByPattern(t *testing.T) {
	t.Parallel()

	patterns := []string{"user:*", "product:*"}

	tests := []struct {
		name        string
		patterns    []string
		mockErrors  []error // one error per pattern deletion
		expectError bool
	}{
		{
			name:        "should succeed when all patterns are deleted successfully",
			patterns:    patterns,
			mockErrors:  []error{nil, nil},
			expectError: false,
		},
		{
			name:        "should return an error when one pattern deletion fails",
			patterns:    patterns,
			mockErrors:  []error{errors.New("delete user pattern failed"), nil},
			expectError: true,
		},
		{
			name:        "should return an error when all pattern deletions fail",
			patterns:    patterns,
			mockErrors:  []error{errors.New("delete user pattern failed"), errors.New("delete product pattern failed")},
			expectError: true,
		},
		{
			name:        "should do nothing when no patterns are provided",
			patterns:    []string{},
			mockErrors:  []error{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			mockStore := omnicachemock.NewMockStore(t)

			for i, pattern := range tt.patterns {
				if len(tt.mockErrors) > i {
					mockStore.Mock.On("DeleteByPattern", ctx, pattern).Return(tt.mockErrors[i])
				}
			}

			manager := &Manager{store: mockStore}

			// --- Act ---
			err := manager.DeleteManyByPattern(ctx, tt.patterns...)

			// --- Assert ---
			for _, pattern := range tt.patterns {
				mockStore.Mock.AssertCalled(t, "DeleteByPattern", ctx, pattern)
			}
			mockStore.Mock.AssertCalledCount(t, "DeleteByPattern", len(tt.patterns))
			mockStore.Mock.AssertExpectations(t)

			if tt.expectError {
				assert.Error(t, err, "must return an error when one or more deletions fail")

				matched := false
				for _, mockErr := range tt.mockErrors {
					if mockErr != nil && err.Error() == mockErr.Error() {
						matched = true
						break
					}
				}

				if !matched {
					assert.Contains(t, tt.mockErrors, err, "error must match one of the expected errors")
				}
				return
			}

			assert.NoError(t, err, "must not return an error when all deletions succeed")
		})
	}
}
