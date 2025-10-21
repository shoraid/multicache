package omnicache

import (
	"testing"

	"github.com/shoraid/omnicache/contract"
	"github.com/shoraid/omnicache/internal/assert"
	omnicachemock "github.com/shoraid/omnicache/mock"
)

func TestManager_NewManager(t *testing.T) {
	t.Parallel()

	// --- Act ---
	manager := NewManager()

	// --- Assert ---
	assert.NotNil(t, manager, "NewManager must return a non-nil Manager instance")
	assert.Empty(t, manager.stores, "the stores map must be initialized but start empty")
	assert.Nil(t, manager.store, "the default store must be nil when no stores have been registered yet")
}

func TestManager_Register(t *testing.T) {
	t.Parallel()

	mockStore1 := new(omnicachemock.MockStore)
	mockStore2 := new(omnicachemock.MockStore)

	tests := []struct {
		name                 string
		alias                string
		storeToRegister      contract.Store
		setup                func(m *Manager)
		expectedErr          error
		expectedStoreLen     int
		expectedDefaultStore contract.Store
	}{
		{
			name:                 "should register the first store as default when no stores exist",
			alias:                "store1",
			storeToRegister:      mockStore1,
			setup:                func(m *Manager) {},
			expectedErr:          nil,
			expectedStoreLen:     1,
			expectedDefaultStore: mockStore1,
		},
		{
			name:            "should register a new store when another store already exists without changing the default",
			alias:           "store2",
			storeToRegister: mockStore2,
			setup: func(m *Manager) {
				m.stores["store1"] = mockStore1
				m.store = mockStore1
			},
			expectedErr:          nil,
			expectedStoreLen:     2,
			expectedDefaultStore: mockStore1,
		},
		{
			name:            "should return an error when alias is already registered",
			alias:           "store1",
			storeToRegister: mockStore2,
			setup: func(m *Manager) {
				m.stores["store1"] = mockStore1
				m.store = mockStore1
			},
			expectedErr:          ErrStoreAlreadyRegistered,
			expectedStoreLen:     1,
			expectedDefaultStore: mockStore1,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			manager := &Manager{
				stores: make(map[string]contract.Store),
			}
			tt.setup(manager)

			// --- Act ---
			err := manager.Register(tt.alias, tt.storeToRegister)

			// --- Assert ---
			if tt.expectedErr != nil {
				assert.Error(t, err, "must return an error when Register fails")
				assert.EqualError(t, tt.expectedErr, err, "the error returned by Register must match the expected error")
				return
			}

			assert.NoError(t, err, "must not return an error when Register succeeds")
			assert.Equal(t, tt.expectedStoreLen, len(manager.stores), "the number of registered stores must match the expected count")
			assert.Equal(t, tt.storeToRegister, manager.stores[tt.alias], "the registered store under alias must match the provided store instance")
			assert.Equal(t, tt.expectedDefaultStore, manager.store, "the default store must be set or preserved correctly after registration")
		})
	}
}

func TestManager_SetDefault(t *testing.T) {
	t.Parallel()

	mockStore1 := new(omnicachemock.MockStore)
	mockStore2 := new(omnicachemock.MockStore)

	tests := []struct {
		name                 string
		aliasToSet           string
		setup                func(m *Manager)
		expectedErr          error
		expectedDefaultStore contract.Store
	}{
		{
			name:       "should set the specified store as the new default when alias exists",
			aliasToSet: "store2",
			setup: func(m *Manager) {
				m.stores["store1"] = mockStore1
				m.stores["store2"] = mockStore2
				m.store = mockStore1 // initial default
			},
			expectedErr:          nil,
			expectedDefaultStore: mockStore2,
		},
		{
			name:       "should return an error when alias does not exist and preserve the current default store",
			aliasToSet: "nonexistent",
			setup: func(m *Manager) {
				m.stores["store1"] = mockStore1
				m.store = mockStore1
			},
			expectedErr:          ErrInvalidDefaultStore,
			expectedDefaultStore: mockStore1, // default should not change
		},
		{
			name:       "should not change the default store when alias is already the current default",
			aliasToSet: "store1",
			setup: func(m *Manager) {
				m.stores["store1"] = mockStore1
				m.store = mockStore1
			},
			expectedErr:          nil,
			expectedDefaultStore: mockStore1,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			manager := &Manager{
				stores: make(map[string]contract.Store),
			}
			tt.setup(manager)

			// --- Act ---
			err := manager.SetDefault(tt.aliasToSet)

			// --- Assert ---
			if tt.expectedErr != nil {
				assert.Error(t, err, "must return an error when SetDefault fails")
				assert.EqualError(t, tt.expectedErr, err, "the error returned by SetDefault must match the expected error")
				assert.Equal(t, tt.expectedDefaultStore, manager.store, "the default store must remain unchanged when SetDefault fails")
				return
			}

			assert.NoError(t, err, "must not return an error when SetDefault succeeds")
			assert.Equal(t, tt.expectedDefaultStore, manager.store, "the default store must be updated to the specified alias when SetDefault succeeds")
		})
	}
}

func TestManager_Store(t *testing.T) {
	t.Parallel()

	mockMemory := new(omnicachemock.MockStore)
	mockRedis := new(omnicachemock.MockStore)

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
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			manager := &Manager{
				stores: make(map[string]contract.Store),
			}
			manager.stores["memory"] = mockMemory
			manager.stores["redis"] = mockRedis
			manager.store = mockMemory // Set initial default

			aliasedManager := manager.Store(tt.alias)

			assert.NotNil(t, aliasedManager, "expected a manager instance")
			assert.Equal(t, tt.expectedStore, aliasedManager.store, "expected the aliased store to be set correctly")
			assert.Equal(t, manager.stores, aliasedManager.stores, "expected the stores map to remain the same")
		})
	}
}
