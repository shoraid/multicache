package omnicache

import (
	"context"
	"errors"
	"testing"

	omnicachemock "github.com/shoraid/omnicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestManager_Clear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "should clear successfully",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return error when clear fails",
			mockErr:     errors.New("clear failed"),
			expectedErr: errors.New("clear failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.ClearFunc = func(_ context.Context) error {
				return tt.mockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			err := manager.Clear(ctx)

			// Assert
			mockStore.CalledOnce(t, "Clear")

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error when clear fails")
				assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
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
			name:        "should delete key successfully",
			key:         key,
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return error when delete fails",
			key:         key,
			mockErr:     errors.New("delete failed"),
			expectedErr: errors.New("delete failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.DeleteFunc = func(_ context.Context, k string) error {
				assert.Equal(t, tt.key, k, "expected correct key to be used")
				return tt.mockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			err := manager.Delete(ctx, tt.key)

			// Assert
			mockStore.CalledOnce(t, "Delete")

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error when delete fails")
				assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
				return
			}

			assert.NoError(t, err, "expected no error when delete succeeds")
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
			name:        "should delete by pattern successfully",
			pattern:     pattern,
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return error when delete by pattern fails",
			pattern:     pattern,
			mockErr:     errors.New("delete by pattern failed"),
			expectedErr: errors.New("delete by pattern failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.DeleteByPatternFunc = func(_ context.Context, p string) error {
				assert.Equal(t, tt.pattern, p, "expected correct pattern to be used")
				return tt.mockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			err := manager.DeleteByPattern(ctx, tt.pattern)

			// Assert
			mockStore.CalledOnce(t, "DeleteByPattern")

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error when delete by pattern fails")
				assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
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
			name:        "should delete many keys successfully",
			keys:        keys,
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "should return error when delete many fails",
			keys:        keys,
			mockErr:     errors.New("delete many failed"),
			expectedErr: errors.New("delete many failed"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.DeleteManyFunc = func(_ context.Context, keys ...string) error {
				assert.Equal(t, tt.keys, keys, "expected correct keys to be used")
				return tt.mockErr
			}

			manager := &Manager{store: mockStore}

			// Act
			err := manager.DeleteMany(ctx, tt.keys...)

			// Assert
			mockStore.CalledOnce(t, "DeleteMany")

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error when delete many fails")
				assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
				return
			}

			assert.NoError(t, err, "expected no error when delete many succeeds")
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
		expectedErr error
	}{
		{
			name:        "should delete many by pattern successfully",
			patterns:    patterns,
			mockErrors:  []error{nil, nil},
			expectedErr: nil,
		},
		{
			name:        "should return error if one deletion fails",
			patterns:    patterns,
			mockErrors:  []error{errors.New("delete user pattern failed"), nil},
			expectedErr: errors.New("delete user pattern failed"),
		},
		{
			name:        "should return error if all deletions fail",
			patterns:    patterns,
			mockErrors:  []error{errors.New("delete user pattern failed"), errors.New("delete product pattern failed")},
			expectedErr: errors.New("delete user pattern failed"),
		},
		{
			name:        "should handle empty patterns list",
			patterns:    []string{},
			mockErrors:  []error{},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			mockStore := new(omnicachemock.MockStore)
			mockStore.DeleteByPatternFunc = func(_ context.Context, p string) error {
				return tt.mockErrors[0]
			}

			manager := &Manager{store: mockStore}

			// Act
			err := manager.DeleteManyByPattern(ctx, tt.patterns...)

			// Assert
			callCount := mockStore.CallCount("DeleteByPattern")
			assert.Equal(t, callCount, len(tt.patterns), "expected DeleteByPattern to be called for each pattern")

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected error when delete many by pattern fails")
				assert.EqualError(t, err, tt.expectedErr.Error(), "expected correct error message")
			} else {
				assert.NoError(t, err, "expected no error when delete many by pattern succeeds")
			}
		})
	}
}
