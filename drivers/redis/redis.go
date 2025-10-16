package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shoraid/multicache"
	"github.com/shoraid/multicache/contract"
)

// define this in redis_store.go
type redisClient interface {
	FlushDB(ctx context.Context) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
}

type scanIterator interface {
	Next(ctx context.Context) bool
	Val() string
	Err() error
}

// newScanIterator creates an iterator. In prod it wraps go-redis.
// In tests you can override this var to return a mock iterator.
var newScanIterator = func(ctx context.Context, c redisClient, pattern string) scanIterator {
	return c.Scan(ctx, 0, pattern, 0).Iterator()
}

type RedisStore struct {
	client redisClient
}

func NewRedisStore(cfg RedisConfig) (contract.Store, error) {
	var tlsConfig *tls.Config

	// Only build TLS config if UseTLS is explicitly true
	if cfg.UseTLS {
		var err error
		tlsConfig, err = buildTLSConfig(cfg.TLSConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
	}

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
		TLSConfig:       tlsConfig, // nil when TLS not used
	}

	client := redis.NewClient(opts)

	return &RedisStore{client}, nil
}

func (r *RedisStore) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStore) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := newScanIterator(ctx, r.client, pattern)
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

func (r *RedisStore) DeleteMany(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisStore) Get(ctx context.Context, key string) (any, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, multicache.ErrCacheMiss
	} else if err != nil {
		return nil, err
	}

	return value, nil
}

func (r *RedisStore) Has(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *RedisStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl < 0 {
		return multicache.ErrInvalidValue
	}

	// Marshal value to JSON to handle all Go types safely
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}
