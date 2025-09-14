package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shoraid/multicache"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(config RedisConfig) (multicache.Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})

	store := &RedisStore{
		client: client,
	}

	return store, nil
}

func (r *RedisStore) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStore) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
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

func (r *RedisStore) GetOrSet(ctx context.Context, key string, ttl time.Duration, value any) (any, error) {
	// Try to get the value first
	val, err := r.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// If it's not a cache miss, return the error
	if !errors.Is(err, multicache.ErrCacheMiss) {
		return nil, err
	}

	// If it's a cache miss, set the value
	if err := r.Set(ctx, key, value, ttl); err != nil {
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
	return r.client.Set(ctx, key, value, ttl).Err()
}
