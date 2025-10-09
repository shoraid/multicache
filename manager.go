package multicache

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type Manager interface {
	Store(alias string) Manager

	Clear(ctx context.Context) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	DeleteMany(ctx context.Context, keys ...string) error
	DeleteManyByPattern(ctx context.Context, patterns ...string) error

	Get(ctx context.Context, key string) (any, error)
	GetOrSet(ctx context.Context, key string, ttl time.Duration, value any) (any, error)
	Has(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	GetBool(ctx context.Context, key string) (bool, error)
	GetBoolOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue bool) (bool, error)

	GetInt(ctx context.Context, key string) (int, error)
	GetInt64(ctx context.Context, key string) (int64, error)
	GetInts(ctx context.Context, key string) ([]int, error)
	GetIntOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue int) (int, error)
	GetInt64OrSet(ctx context.Context, key string, ttl time.Duration, defaultValue int64) (int64, error)
	GetIntsOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue []int) ([]int, error)

	GetString(ctx context.Context, key string) (string, error)
	GetStrings(ctx context.Context, key string) ([]string, error)
	GetStringOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue string) (string, error)
	GetStringsOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue []string) ([]string, error)
}

type managerImpl struct {
	mu     sync.RWMutex
	stores map[string]Store
	store  Store
}

func NewManager(defaultStore string, stores map[string]Store) (Manager, error) {
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

func (m *managerImpl) Store(alias string) Manager {
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
