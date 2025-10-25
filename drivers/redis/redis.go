package redisstore

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"github.com/shoraid/omnicache"
	"github.com/shoraid/omnicache/contract"
)

type redisClient interface {
	FlushDB(ctx context.Context) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
}

type RedisStore struct {
	client redisClient
}

// NewRedisWithClient creates a new RedisStore instance with a pre-existing redisClient.
// This is useful for testing or when you want to manage the Redis client lifecycle externally.
func NewRedisWithClient(client redisClient) (contract.Store, error) {
	return &RedisStore{client}, nil
}

// NewRedisStore creates a new RedisStore instance with the given RedisConfig.
// It initializes a new Redis client based on the provided configuration.
func NewRedisStore(cfg RedisConfig) (contract.Store, error) {
	opts := &redis.Options{
		Addr:            cfg.Addr,
		ClientName:      cfg.ClientName,
		Username:        cfg.Username,
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolSize:        cfg.PoolSize,
		PoolTimeout:     cfg.PoolTimeout,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	}

	client := redis.NewClient(opts)

	return &RedisStore{client}, nil
}

// Clear removes all entries from the cache.
func (r *RedisStore) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Close closes the Redis client.
func (r *RedisStore) Close(ctx context.Context) error {
	if client, ok := r.client.(*redis.Client); ok {
		return client.Close()
	}

	return nil
}

// Delete removes the entry associated with the given key from the cache.
func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// DeleteByPattern removes all cache entries whose keys match the given pattern.
func (r *RedisStore) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// DeleteMany removes multiple keys from the cache in a single call.
func (r *RedisStore) DeleteMany(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(ctx, keys...).Err()
}

// Get retrieves a value from the cache by key.
func (r *RedisStore) Get(ctx context.Context, key string) (any, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, omnicache.ErrCacheMiss
	} else if err != nil {
		return nil, err
	}

	return value, nil
}

// Has checks whether a key exists in the cache.
func (r *RedisStore) Has(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Set stores a value in the cache with the given key and an optional time-to-live (TTL).
func (r *RedisStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl < 0 {
		return omnicache.ErrInvalidValue
	}

	// Marshal value to JSON to handle all Go types safely
	data, err := sonic.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}
