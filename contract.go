package multicache

import "time"

type Factory func(config map[string]any) (Store, error)

type Store interface {
	Add(key string, value any, ttl ...time.Duration) error
	Flush() error
	Forget(key string) error
	Get(key string, fallback ...any) (any, error)
	Has(key string) (bool, error)
	Put(key string, value any, ttl ...time.Duration) error
}
