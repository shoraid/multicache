package multicache

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockStore implements multicache.Store using testify/mock
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Clear() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStore) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockStore) DeleteByPattern(pattern string) error {
	args := m.Called(pattern)
	return args.Error(0)
}

func (m *MockStore) DeleteMany(keys ...string) error {
	args := m.Called(keys)
	return args.Error(0)
}

func (m *MockStore) Get(key string) (any, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockStore) GetOrSet(key string, ttl time.Duration, value any) (any, error) {
	args := m.Called(key, ttl, value)
	return args.Get(0), args.Error(1)
}

func (m *MockStore) Has(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *MockStore) Set(key string, value any, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}
