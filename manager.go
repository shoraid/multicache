package multicache

import (
	"encoding/json"
	"time"

	"github.com/spf13/cast"
)

type Manager interface {
	Store(alias string) Manager
	Clear() error
	Delete(key string) error
	DeleteByPattern(pattern string) error
	DeleteMany(keys ...string) error
	Get(key string) (any, error)
	GetBool(key string) (bool, error)
	GetInt(key string) (int, error)
	GetInt64(key string) (int64, error)
	GetInts(key string) ([]int, error)
	GetString(key string) (string, error)
	GetStrings(key string) ([]string, error)
	GetOrSet(key string, ttl time.Duration, value any) (any, error)
	Has(key string) (bool, error)
	Set(key string, value any, ttl time.Duration) error
}

type managerImpl struct {
	stores map[string]Store
	store  Store
}

func NewManager(defaultStore string, stores map[string]Store) (Manager, error) {
	store, exists := stores[defaultStore]
	if !exists {
		return nil, ErrInvalidDefaultStore
	}

	return &managerImpl{stores, store}, nil
}

func (m *managerImpl) Store(alias string) Manager {
	return &managerImpl{
		stores: m.stores,
		store:  m.stores[alias],
	}
}

func (m *managerImpl) Clear() error {
	return m.store.Clear()
}

func (m *managerImpl) Delete(key string) error {
	return m.store.Delete(key)
}

func (m *managerImpl) DeleteByPattern(pattern string) error {
	return m.store.DeleteByPattern(pattern)
}

func (m *managerImpl) DeleteMany(keys ...string) error {
	return m.store.DeleteMany(keys...)
}

func (m *managerImpl) Get(key string) (any, error) {
	return m.store.Get(key)
}

func (m *managerImpl) GetBool(key string) (bool, error) {
	val, err := m.store.Get(key)
	if err != nil {
		return false, err
	}

	boolVal, err := cast.ToBoolE(val)
	if err != nil {
		return false, err
	}

	return boolVal, nil
}

func (m *managerImpl) GetInt(key string) (int, error) {
	val, err := m.store.Get(key)
	if err != nil {
		return 0, err
	}

	intVal, err := cast.ToIntE(val)
	if err != nil {
		return 0, err
	}

	return intVal, nil
}

func (m *managerImpl) GetInt64(key string) (int64, error) {
	val, err := m.store.Get(key)
	if err != nil {
		return 0, err
	}

	intVal, err := cast.ToInt64E(val)
	if err != nil {
		return 0, err
	}

	return intVal, nil
}

func (m *managerImpl) GetInts(key string) ([]int, error) {
	val, err := m.store.Get(key)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case string:
		var data []int
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return nil, err
		}
		return data, nil

	case []byte:
		var data []int
		if err := json.Unmarshal(v, &data); err != nil {
			return nil, err
		}
		return data, nil

	case []int:
		return v, nil

	default:
		return nil, ErrTypeMismatch
	}
}

func (m *managerImpl) GetString(key string) (string, error) {
	val, err := m.store.Get(key)
	if err != nil {
		return "", err
	}

	strVal, err := cast.ToStringE(val)
	if err != nil {
		return "", err
	}

	return strVal, nil
}

func (m *managerImpl) GetStrings(key string) ([]string, error) {
	val, err := m.store.Get(key)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case string:
		var data []string
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return nil, err
		}
		return data, nil

	case []byte:
		var data []string
		if err := json.Unmarshal(v, &data); err != nil {
			return nil, err
		}
		return data, nil

	case []string:
		return v, nil

	default:
		return nil, ErrTypeMismatch
	}
}

func (m *managerImpl) GetOrSet(key string, ttl time.Duration, value any) (any, error) {
	return m.store.GetOrSet(key, ttl, value)
}

func (m *managerImpl) Has(key string) (bool, error) {
	return m.store.Has(key)
}

func (m *managerImpl) Set(key string, value any, ttl time.Duration) error {
	return m.store.Set(key, value, ttl)
}
