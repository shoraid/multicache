package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type mockRedisClient struct {
	flushDBFunc func(ctx context.Context) *redis.StatusCmd
	delFunc     func(ctx context.Context, keys ...string) *redis.IntCmd
	scanFunc    func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	getFunc     func(ctx context.Context, key string) *redis.StringCmd
	existsFunc  func(ctx context.Context, keys ...string) *redis.IntCmd
	setFunc     func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

func (m *mockRedisClient) FlushDB(ctx context.Context) *redis.StatusCmd {
	return m.flushDBFunc(ctx)
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return m.delFunc(ctx, keys...)
}

func (m *mockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	return m.scanFunc(ctx, cursor, match, count)
}
func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return m.getFunc(ctx, key)
}
func (m *mockRedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	return m.existsFunc(ctx, keys...)
}
func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	return m.setFunc(ctx, key, value, ttl)
}

type mockIter struct {
	keys []string
	i    int
	err  error
}

func (m *mockIter) Next(ctx context.Context) bool {
	if m.i >= len(m.keys) {
		return false
	}
	m.i++
	return true
}
func (m *mockIter) Val() string { return m.keys[m.i-1] }
func (m *mockIter) Err() error  { return m.err }
