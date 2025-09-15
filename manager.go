package multicache

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

type Manager interface {
	Store(alias string) Manager
	Clear(ctx context.Context) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	DeleteMany(ctx context.Context, keys ...string) error
	DeleteManyByPattern(ctx context.Context, patterns ...string) error
	Get(ctx context.Context, key string) (any, error)
	GetBool(ctx context.Context, key string) (bool, error)
	GetBoolOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue bool) (bool, error)
	GetInt(ctx context.Context, key string) (int, error)
	GetIntOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue int) (int, error)
	GetInt64(ctx context.Context, key string) (int64, error)
	GetInt64OrSet(ctx context.Context, key string, ttl time.Duration, defaultValue int64) (int64, error)
	GetInts(ctx context.Context, key string) ([]int, error)
	GetIntsOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue []int) ([]int, error)
	GetOrSet(ctx context.Context, key string, ttl time.Duration, value any) (any, error)
	GetString(ctx context.Context, key string) (string, error)
	GetStringOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue string) (string, error)
	GetStrings(ctx context.Context, key string) ([]string, error)
	GetStringsOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue []string) ([]string, error)
	Has(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

type managerImpl struct {
	mu     sync.RWMutex
	stores map[string]Store
	store  Store
}

func NewManager(defaultStore string, stores map[string]Store) (Manager, error) {
	if len(stores) == 0 {
		log.
			Error().
			Str("defaultStore", defaultStore).
			Msg("cache manager: stores map is empty")

		return nil, ErrInvalidDefaultStore
	}

	store, exists := stores[defaultStore]
	if !exists {
		log.
			Error().
			Str("defaultStore", defaultStore).
			Msgf("cache manager: default store %q not found", defaultStore)

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
		log.
			Error().
			Str("alias", alias).
			Msgf("cache manager: store with alias %q not found", alias)
	}

	return &managerImpl{
		stores: m.stores,
		store:  store,
	}
}

func (m *managerImpl) Clear(ctx context.Context) error {
	return m.store.Clear(ctx)
}

func (m *managerImpl) Delete(ctx context.Context, key string) error {
	return m.store.Delete(ctx, key)
}

func (m *managerImpl) DeleteByPattern(ctx context.Context, pattern string) error {
	return m.store.DeleteByPattern(ctx, pattern)
}

func (m *managerImpl) DeleteMany(ctx context.Context, keys ...string) error {
	return m.store.DeleteMany(ctx, keys...)
}

func (m *managerImpl) DeleteManyByPattern(ctx context.Context, patterns ...string) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, pattern := range patterns {
		p := pattern
		g.Go(func() error {
			return m.store.DeleteByPattern(ctx, p)
		})
	}

	return g.Wait() // returns first non-nil error (or nil if all succeeded)
}

func (m *managerImpl) Get(ctx context.Context, key string) (any, error) {
	return m.store.Get(ctx, key)
}

func (m *managerImpl) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return false, err
	}

	boolVal, err := cast.ToBoolE(val)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("key", key).
			Msg("cache manager: failed to cast to boolean")

		return false, ErrTypeMismatch
	}

	return boolVal, nil
}

func (m *managerImpl) GetBoolOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue bool) (bool, error) {
	val, err := m.GetBool(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// If value not found or cast error, store default
		if err := m.Set(ctx, key, defaultValue, ttl); err != nil {
			return defaultValue, err
		}
		return defaultValue, nil
	}

	return false, err
}

func (m *managerImpl) GetInt(ctx context.Context, key string) (int, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	intVal, err := cast.ToIntE(val)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("key", key).
			Msg("cache manager: failed to cast to integer")

		return 0, ErrTypeMismatch
	}

	return intVal, nil
}

func (m *managerImpl) GetIntOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue int) (int, error) {
	val, err := m.GetInt(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// If value not found or cast error, store default
		if err := m.Set(ctx, key, defaultValue, ttl); err != nil {
			return defaultValue, err
		}
		return defaultValue, nil
	}

	return 0, err
}

func (m *managerImpl) GetInt64(ctx context.Context, key string) (int64, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	intVal, err := cast.ToInt64E(val)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("key", key).
			Msg("cache manager: failed to cast to integer64")

		return 0, ErrTypeMismatch
	}

	return intVal, nil
}

func (m *managerImpl) GetInt64OrSet(ctx context.Context, key string, ttl time.Duration, defaultValue int64) (int64, error) {
	val, err := m.GetInt64(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// If value not found or cast error, store default
		if err := m.Set(ctx, key, defaultValue, ttl); err != nil {
			return defaultValue, err
		}
		return defaultValue, nil
	}

	return 0, err
}

func (m *managerImpl) GetInts(ctx context.Context, key string) ([]int, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case string:
		var data []int
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			log.
				Error().
				Err(err).
				Str("key", key).
				Msg("cache manager: failed to unmarshal string to []int")

			return nil, ErrTypeMismatch
		}
		return data, nil

	case []byte:
		var data []int
		if err := json.Unmarshal(v, &data); err != nil {
			log.
				Error().
				Err(err).
				Str("key", key).
				Msg("cache manager: failed to unmarshal []byte to []int")

			return nil, ErrTypeMismatch
		}
		return data, nil

	case []int:
		return v, nil

	default:
		return nil, ErrTypeMismatch
	}
}

func (m *managerImpl) GetIntsOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue []int) ([]int, error) {
	val, err := m.GetInts(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// If value not found or cast error, store default
		if err := m.Set(ctx, key, defaultValue, ttl); err != nil {
			return defaultValue, err
		}
		return defaultValue, nil
	}

	return nil, err
}

func (m *managerImpl) GetString(ctx context.Context, key string) (string, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return "", err
	}

	switch v := val.(type) {
	case string:
		return v, nil

	case []byte:
		return string(v), nil

	default:
		return "", ErrTypeMismatch
	}
}

func (m *managerImpl) GetStringOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue string) (string, error) {
	val, err := m.GetString(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// If value not found or cast error, store default
		if err := m.Set(ctx, key, defaultValue, ttl); err != nil {
			return defaultValue, err
		}
		return defaultValue, nil
	}

	return "", err
}

func (m *managerImpl) GetStrings(ctx context.Context, key string) ([]string, error) {
	val, err := m.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case string:
		var data []string
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			log.
				Error().
				Err(err).
				Str("key", key).
				Msg("cache manager: failed to unmarshal string to []string")

			return nil, ErrTypeMismatch
		}
		return data, nil

	case []byte:
		var data []string
		if err := json.Unmarshal(v, &data); err != nil {
			log.
				Error().
				Err(err).
				Str("key", key).
				Msg("cache manager: failed to unmarshal []byte to []string")

			return nil, ErrTypeMismatch
		}
		return data, nil

	case []string:
		return v, nil

	default:
		return nil, ErrTypeMismatch
	}
}

func (m *managerImpl) GetStringsOrSet(ctx context.Context, key string, ttl time.Duration, defaultValue []string) ([]string, error) {
	val, err := m.GetStrings(ctx, key)
	if err == nil {
		return val, nil
	}

	if errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrTypeMismatch) {
		// If value not found or cast error, store default
		if err := m.Set(ctx, key, defaultValue, ttl); err != nil {
			return defaultValue, err
		}
		return defaultValue, nil
	}

	return nil, err
}

func (m *managerImpl) GetOrSet(ctx context.Context, key string, ttl time.Duration, value any) (any, error) {
	return m.store.GetOrSet(ctx, key, ttl, value)
}

func (m *managerImpl) Has(ctx context.Context, key string) (bool, error) {
	return m.store.Has(ctx, key)
}

func (m *managerImpl) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return m.store.Set(ctx, key, value, ttl)
}
