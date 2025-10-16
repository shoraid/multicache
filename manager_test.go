package multicache

import (
	"testing"

	"github.com/shoraid/multicache/contract"
	multicachemock "github.com/shoraid/multicache/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	manager := NewManager()

	assert.NotNil(t, manager, "NewManager should return a non-nil manager")
	assert.Empty(t, manager.(*managerImpl).stores, "NewManager should initialize with an empty stores map")
	assert.Nil(t, manager.(*managerImpl).store, "NewManager should initialize with a nil default store")
}

func TestManager_Register(t *testing.T) {
	t.Parallel()

	mockStore1 := new(multicachemock.MockStore)
	mockStore2 := new(multicachemock.MockStore)

	tests := []struct {
		name                 string
		alias                string
		storeToRegister      contract.Store
		setup                func(m *managerImpl)
		expectedErr          error
		expectedStoreCount   int
		expectedDefaultStore contract.Store
	}{
		{
			name:                 "should register first store as default",
			alias:                "store1",
			storeToRegister:      mockStore1,
			setup:                func(m *managerImpl) {},
			expectedErr:          nil,
			expectedStoreCount:   1,
			expectedDefaultStore: mockStore1,
		},
		{
			name:            "should register additional store without changing default",
			alias:           "store2",
			storeToRegister: mockStore2,
			setup: func(m *managerImpl) {
				m.stores["store1"] = mockStore1
				m.store = mockStore1
			},
			expectedErr:          nil,
			expectedStoreCount:   2,
			expectedDefaultStore: mockStore1,
		},
		{
			name:            "should return error if alias already registered",
			alias:           "store1",
			storeToRegister: mockStore2,
			setup: func(m *managerImpl) {
				m.stores["store1"] = mockStore1
				m.store = mockStore1
			},
			expectedErr:          ErrStoreAlreadyRegistered,
			expectedStoreCount:   1,
			expectedDefaultStore: mockStore1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			manager := &managerImpl{
				stores: make(map[string]contract.Store),
			}

			tt.setup(manager)

			err := manager.Register(tt.alias, tt.storeToRegister)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedStoreCount, len(manager.stores), "expected correct number of registered stores")
				assert.Equal(t, tt.storeToRegister, manager.stores[tt.alias], "expected store to be registered under alias")
				assert.Equal(t, tt.expectedDefaultStore, manager.store, "expected correct default store")
			}
		})
	}
}

func TestManager_SetDefault(t *testing.T) {
	t.Parallel()

	mockStore1 := new(multicachemock.MockStore)
	mockStore2 := new(multicachemock.MockStore)

	tests := []struct {
		name                 string
		aliasToSet           string
		setup                func(m *managerImpl)
		expectedErr          error
		expectedDefaultStore contract.Store
	}{
		{
			name:       "should set default store successfully",
			aliasToSet: "store2",
			setup: func(m *managerImpl) {
				m.stores["store1"] = mockStore1
				m.stores["store2"] = mockStore2
				m.store = mockStore1 // initial default
			},
			expectedErr:          nil,
			expectedDefaultStore: mockStore2,
		},
		{
			name:       "should return error if alias does not exist",
			aliasToSet: "nonexistent",
			setup: func(m *managerImpl) {
				m.stores["store1"] = mockStore1
				m.store = mockStore1
			},
			expectedErr:          ErrInvalidDefaultStore,
			expectedDefaultStore: mockStore1, // default should not change
		},
		{
			name:       "should not change default if setting to current default",
			aliasToSet: "store1",
			setup: func(m *managerImpl) {
				m.stores["store1"] = mockStore1
				m.store = mockStore1
			},
			expectedErr:          nil,
			expectedDefaultStore: mockStore1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			manager := &managerImpl{
				stores: make(map[string]contract.Store),
			}

			tt.setup(manager)

			err := manager.SetDefault(tt.aliasToSet)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedDefaultStore, manager.store, "expected correct default store")
			}
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

			manager := &managerImpl{
				stores: make(map[string]contract.Store),
			}
			manager.stores["memory"] = mockMemory
			manager.stores["redis"] = mockRedis
			manager.store = mockMemory // Set initial default

			aliasedManager := manager.Store(tt.alias)

			assert.NotNil(t, aliasedManager, "expected a manager instance")
			assert.Equal(t, tt.expectedStore, aliasedManager.(*managerImpl).store, "expected the aliased store to be set correctly")
			assert.Equal(t, manager.stores, aliasedManager.(*managerImpl).stores, "expected the stores map to remain the same")
		})
	}
}
