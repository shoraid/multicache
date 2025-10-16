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

func NewManager(defaultStore string, stores map[string]contract.Store) (contract.Manager, error) {
	if len(stores) == 0 {
		return nil, ErrInvalidDefaultStore
	}

	store, exists := stores[defaultStore]
	if !exists {
		return nil, ErrInvalidDefaultStore
	}

	return &managerImpl{
		stores: stores,
		store:  store,
	}, nil
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
