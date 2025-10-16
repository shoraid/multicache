package multicache

import (
	"fmt"
	"os"
	"sync"

	"github.com/shoraid/multicache/contract"
)

type managerImpl struct {
	mu     sync.RWMutex
	stores map[string]contract.Store
	store  contract.Store
}

func NewManager() contract.Manager {
	return &managerImpl{
		stores: make(map[string]contract.Store),
	}
}

func (m *managerImpl) Register(alias string, store contract.Store) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.stores) == 0 {
		m.store = store
	}

	if _, exists := m.stores[alias]; exists {
		return ErrStoreAlreadyRegistered
	}

	m.stores[alias] = store

	return nil
}

func (m *managerImpl) SetDefault(alias string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	store, exists := m.stores[alias]
	if !exists {
		return ErrInvalidDefaultStore
	}

	m.store = store

	return nil
}

func (m *managerImpl) Store(alias string) contract.Manager {
	m.mu.RLock()
	defer m.mu.RUnlock()

	store, exists := m.stores[alias]
	if !exists {
		fmt.Fprintf(os.Stderr, "[multicache] warning: store alias %q not found, using default store\n", alias)
		return m
	}

	return &managerImpl{
		stores: m.stores,
		store:  store,
	}
}
