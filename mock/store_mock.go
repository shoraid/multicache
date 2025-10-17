package multicachemock

import (
	"context"
	"sync"
	"testing"
	"time"
)

// MockStore is a thread-safe manual mock for contract.Store.
type MockStore struct {
	mu sync.RWMutex

	ClearFunc           func(ctx context.Context) error
	DeleteFunc          func(ctx context.Context, key string) error
	DeleteByPatternFunc func(ctx context.Context, pattern string) error
	DeleteManyFunc      func(ctx context.Context, keys ...string) error
	GetFunc             func(ctx context.Context, key string) (any, error)
	HasFunc             func(ctx context.Context, key string) (bool, error)
	SetFunc             func(ctx context.Context, key string, value any, ttl time.Duration) error

	Calls []string
}

func (m *MockStore) recordCall(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, method)
}

func (m *MockStore) Clear(ctx context.Context) error {
	m.recordCall("Clear")
	if m.ClearFunc != nil {
		return m.ClearFunc(ctx)
	}
	return nil
}

func (m *MockStore) Delete(ctx context.Context, key string) error {
	m.recordCall("Delete")
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

func (m *MockStore) DeleteByPattern(ctx context.Context, pattern string) error {
	m.recordCall("DeleteByPattern")
	if m.DeleteByPatternFunc != nil {
		return m.DeleteByPatternFunc(ctx, pattern)
	}
	return nil
}

func (m *MockStore) DeleteMany(ctx context.Context, keys ...string) error {
	m.recordCall("DeleteMany")
	if m.DeleteManyFunc != nil {
		return m.DeleteManyFunc(ctx, keys...)
	}
	return nil
}

func (m *MockStore) Get(ctx context.Context, key string) (any, error) {
	m.recordCall("Get")
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockStore) Has(ctx context.Context, key string) (bool, error) {
	m.recordCall("Has")
	if m.HasFunc != nil {
		return m.HasFunc(ctx, key)
	}
	return false, nil
}

func (m *MockStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	m.recordCall("Set")
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, ttl)
	}
	return nil
}

// Reset clears all mock behaviors and recorded calls.
func (m *MockStore) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ClearFunc = nil
	m.DeleteFunc = nil
	m.DeleteByPatternFunc = nil
	m.DeleteManyFunc = nil
	m.GetFunc = nil
	m.HasFunc = nil
	m.SetFunc = nil
	m.Calls = nil
}

// CallCount returns the number of times a method was called.
func (m *MockStore) CallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, c := range m.Calls {
		if c == method {
			count++
		}
	}
	return count
}

func (m *MockStore) CalledOnce(t *testing.T, method string) {
	t.Helper()
	count := m.CallCount(method)
	if count != 1 {
		t.Fatalf("expected %s to be called once, but was called %d times", method, count)
	}
}

func (m *MockStore) NotCalled(t *testing.T, method string) {
	t.Helper()
	count := m.CallCount(method)
	if count > 0 {
		t.Fatalf("expected %s not to be called, but it was called %d times", method, count)
	}
}
