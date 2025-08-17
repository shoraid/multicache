package multicache

import "time"

type Factory func(config map[string]any) (Store, error)

type Store interface {
	Flush() error
	Forget(key string) error
	Get(key string, fallback ...any) (any, error)
	Put(key string, value any, ttl ...time.Duration) error
}
