package multicache

import (
	"testing"

	"github.com/shoraid/multicache/contract"
	multicachemock "github.com/shoraid/multicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestManager_NewManager(t *testing.T) {
	t.Parallel()

	mockStore := new(multicachemock.MockStore)

	tests := []struct {
		name         string
		defaultStore string
		stores       map[string]contract.Store
		expectedErr  error
		expectedNil  bool
	}{
		{
			name:         "should return manager when default store exists",
			defaultStore: "default",
			stores:       map[string]contract.Store{"default": mockStore},
			expectedErr:  nil,
			expectedNil:  false,
		},
		{
			name:         "should return error when default store does not exist",
			defaultStore: "missing",
			stores:       map[string]contract.Store{"default": mockStore},
			expectedErr:  ErrInvalidDefaultStore,
			expectedNil:  true,
		},
		{
			name:         "should return error when stores map is empty",
			defaultStore: "default",
			stores:       map[string]contract.Store{},
			expectedErr:  ErrInvalidDefaultStore,
			expectedNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mgr, err := NewManager(tt.defaultStore, tt.stores)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "expected error to match")
				assert.Nil(t, mgr, "expected manager to be nil")
				return
			}

			assert.NoError(t, err, "expected no error")
			assert.NotNil(t, mgr, "expected manager to be non-nil")
			assert.Equal(t, tt.stores, mgr.(*managerImpl).stores, "expected stores to match input map")
			assert.Equal(t, tt.stores[tt.defaultStore], mgr.(*managerImpl).store, "expected default store to match")
		})
	}
}

func TestManager_Store(t *testing.T) {
	t.Parallel()

	mockMemory := new(multicachemock.MockStore)
	mockRedis := new(multicachemock.MockStore)

	tests := []struct {
		name          string
		alias         string
		expectedStore contract.Store
	}{
		{
			name:          "should return manager with specified store when alias exists",
			alias:         "redis",
			expectedStore: mockRedis,
		},
		{
			name:          "should return manager with default store when alias does not exist",
			alias:         "nonexistent",
			expectedStore: mockMemory, // Should fall back to the manager's current store (memory in this case)
		},
		{
			name:          "should return manager with default store when alias is empty",
			alias:         "",
			expectedStore: mockMemory,
		},
		{
			name:          "should return manager with default store when alias is same as default",
			alias:         "memory",
			expectedStore: mockMemory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stores := map[string]contract.Store{
				"memory": mockMemory,
				"redis":  mockRedis,
			}

			manager, _ := NewManager("memory", stores) // Default is memory

			aliasedManager := manager.Store(tt.alias)
			assert.NotNil(t, aliasedManager, "expected a manager instance")
			assert.Equal(t, tt.expectedStore, aliasedManager.(*managerImpl).store, "expected the aliased store to be set correctly")
			assert.Equal(t, stores, aliasedManager.(*managerImpl).stores, "expected the stores map to remain the same")
		})
	}
}
