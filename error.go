package multicache

import "errors"

var (
	ErrCacheMiss              = errors.New("cache: cache miss")
	ErrInternal               = errors.New("cache: internal error")
	ErrInvalidConfig          = errors.New("cache: invalid config")
	ErrInvalidDefaultStore    = errors.New("cache: invalid default cache store")
	ErrInvalidStore           = errors.New("cache: invalid cache store")
	ErrStoreAlreadyRegistered = errors.New("cache: store already registered")
	ErrInvalidValue           = errors.New("cache: invalid value")
	ErrTypeMismatch           = errors.New("cache: value type mismatch")
)
