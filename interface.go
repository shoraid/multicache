package multicache

import (
	"context"
	"time"
)

type Store interface {
	Clear(ctx context.Context) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	DeleteMany(ctx context.Context, keys ...string) error
	Get(ctx context.Context, key string) (any, error)
	Has(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}
