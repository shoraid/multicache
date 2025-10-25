package redismock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// MockRedisClient implements a minimal subset of redis.Cmdable for testing.
type MockRedisClient struct {
	FlushDBFunc func(ctx context.Context) *redis.StatusCmd
	DelFunc     func(ctx context.Context, keys ...string) *redis.IntCmd
	ScanFunc    func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	GetFunc     func(ctx context.Context, key string) *redis.StringCmd
	ExistsFunc  func(ctx context.Context, keys ...string) *redis.IntCmd
	SetFunc     func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	CloseFunc   func() error
}

func (m *MockRedisClient) FlushDB(ctx context.Context) *redis.StatusCmd {
	if m.FlushDBFunc != nil {
		return m.FlushDBFunc(ctx)
	}

	return redis.NewStatusCmd(ctx)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if m.DelFunc != nil {
		return m.DelFunc(ctx, keys...)
	}

	return redis.NewIntCmd(ctx)
}

func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	if m.ScanFunc != nil {
		return m.ScanFunc(ctx, cursor, match, count)
	}

	return redis.NewScanCmdResult([]string{}, 0, nil)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}

	return redis.NewStringCmd(ctx)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, keys...)
	}

	return redis.NewIntCmd(ctx)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) *redis.StatusCmd {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, ttl)
	}

	return redis.NewStatusCmd(ctx)
}

func (m *MockRedisClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
