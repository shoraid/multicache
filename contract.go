package multicache

import "time"

type Factory func(config map[string]any) (Store, error)

type Store interface {
	Clear() error
	Delete(key string) error
	DeleteByPattern(pattern string) error
	DeleteMany(keys ...string) error
	Get(key string) (any, error)
	GetOrSet(key string, ttl time.Duration, value any) (any, error)
	Has(key string) (bool, error)
	Set(key string, value any, ttl time.Duration) error
}
