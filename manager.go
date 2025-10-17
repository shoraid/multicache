package omnicache

import (
	"fmt"
	"os"
	"sync"

	"github.com/shoraid/omnicache/contract"
)

type Manager struct {
	mu     sync.RWMutex
	stores map[string]contract.Store
	store  contract.Store
}

func NewManager() *Manager {
	return &Manager{
		stores: make(map[string]contract.Store),
	}
}

// Register adds a new store with the given alias.
// The first registered store becomes the default.
// Returns an error if the alias is already registered.
func (m *Manager) Register(alias string, store contract.Store) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// First store becomes default
	if len(m.stores) == 0 {
		m.store = store
	}

	if _, exists := m.stores[alias]; exists {
		return ErrStoreAlreadyRegistered
	}

	m.stores[alias] = store

	return nil
}

// SetDefault sets the store with the given alias as the default store.
// Returns an error if the alias has not been registered.
func (m *Manager) SetDefault(alias string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	store, exists := m.stores[alias]
	if !exists {
		return ErrInvalidDefaultStore
	}

	m.store = store

	return nil
}

// Store switches the active cache store to the one registered under
// the given alias. It returns a new Manager instance bound to that
// store. If the alias does not exist, the implementation may return
// a no-op manager or panic, depending on configuration.
func (m *Manager) Store(alias string) *Manager {
	m.mu.RLock()
	defer m.mu.RUnlock()

	store, exists := m.stores[alias]
	if !exists {
		fmt.Fprintf(os.Stderr, "[omnicache] warning: store alias %q not found, using default store\n", alias)
		return m
	}

	return &Manager{
		stores: m.stores,
		store:  store,
	}
}
