package multicachemock

import (
	"context"
	"time"

	"github.com/shoraid/multicache/contract"
	"github.com/stretchr/testify/mock"
)

type MockManager struct {
	mock.Mock
}

func (m *MockManager) Store(alias string) contract.Manager {
	args := m.Called(alias)
	return args.Get(0).(contract.Manager)
}

func (m *MockManager) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockManager) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockManager) DeleteByPattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockManager) DeleteMany(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockManager) DeleteManyByPattern(ctx context.Context, patterns ...string) error {
	args := m.Called(ctx, patterns)
	return args.Error(0)
}

func (m *MockManager) Get(ctx context.Context, key string) (any, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockManager) GetOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (any, error)) (any, error) {
	args := m.Called(ctx, key, ttl, defaultFn)
	return args.Get(0), args.Error(1)
}

func (m *MockManager) GetBool(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockManager) GetBoolOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (bool, error)) (bool, error) {
	args := m.Called(ctx, key, ttl, defaultFn)
	return args.Bool(0), args.Error(1)
}

func (m *MockManager) GetInt(ctx context.Context, key string) (int, error) {
	args := m.Called(ctx, key)
	return args.Int(0), args.Error(1)
}

func (m *MockManager) GetInt64(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockManager) GetInts(ctx context.Context, key string) ([]int, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockManager) GetIntOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (int, error)) (int, error) {
	args := m.Called(ctx, key, ttl, defaultFn)
	return args.Int(0), args.Error(1)
}

func (m *MockManager) GetInt64OrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (int64, error)) (int64, error) {
	args := m.Called(ctx, key, ttl, defaultFn)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockManager) GetIntsOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() ([]int, error)) ([]int, error) {
	args := m.Called(ctx, key, ttl, defaultFn)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockManager) GetString(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockManager) GetStrings(ctx context.Context, key string) ([]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockManager) GetStringOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() (string, error)) (string, error) {
	args := m.Called(ctx, key, ttl, defaultFn)
	return args.String(0), args.Error(1)
}

func (m *MockManager) GetStringsOrSet(ctx context.Context, key string, ttl time.Duration, defaultFn func() ([]string, error)) ([]string, error) {
	args := m.Called(ctx, key, ttl, defaultFn)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockManager) Has(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockManager) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}
