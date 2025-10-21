package omnicachemock

import (
	"context"
	"testing"
	"time"

	"github.com/shoraid/omnicache/internal/testutil"
)

// MockStore is a lightweight mock for contract.Store, powered by testutil.MockHelper.
type MockStore struct {
	Mock *testutil.MockHelper
}

// NewMockStore creates a new mock store.
func NewMockStore(t *testing.T) *MockStore {
	return &MockStore{Mock: testutil.NewMockHelper(t)}
}

// --- Store method mocks ---

func (m *MockStore) Clear(ctx context.Context) error {
	args := m.Mock.Called("Clear", ctx)
	if len(args) == 0 {
		return nil
	}
	if err, ok := args[0].(error); ok {
		return err
	}
	return nil
}

func (m *MockStore) Delete(ctx context.Context, key string) error {
	args := m.Mock.Called("Delete", ctx, key)
	if len(args) == 0 {
		return nil
	}
	if err, ok := args[0].(error); ok {
		return err
	}
	return nil
}

func (m *MockStore) DeleteByPattern(ctx context.Context, pattern string) error {
	args := m.Mock.Called("DeleteByPattern", ctx, pattern)
	if len(args) == 0 {
		return nil
	}
	if err, ok := args[0].(error); ok {
		return err
	}
	return nil
}

func (m *MockStore) DeleteMany(ctx context.Context, keys ...string) error {
	args := m.Mock.Called("DeleteMany", ctx, keys)
	if len(args) == 0 {
		return nil
	}
	if err, ok := args[0].(error); ok {
		return err
	}
	return nil
}

func (m *MockStore) Get(ctx context.Context, key string) (any, error) {
	args := m.Mock.Called("Get", ctx, key)
	if len(args) >= 2 {
		return args[0], asError(args[1])
	}
	return nil, nil
}

func (m *MockStore) Has(ctx context.Context, key string) (bool, error) {
	args := m.Mock.Called("Has", ctx, key)
	if len(args) >= 2 {
		val, _ := args[0].(bool)
		return val, asError(args[1])
	}
	return false, nil
}

func (m *MockStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	args := m.Mock.Called("Set", ctx, key, value, ttl)
	if len(args) == 0 {
		return nil
	}
	return asError(args[0])
}

// --- Helpers ---

func asError(v any) error {
	if v == nil {
		return nil
	}
	if err, ok := v.(error); ok {
		return err
	}
	return nil
}
