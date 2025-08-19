package multicache

import "errors"

var ErrCacheMiss = errors.New("cache: cache miss")
var ErrInvalidDefaultStore = errors.New("cache: invalid default cache store")
var ErrInvalidValue = errors.New("cache: invalid value")
var ErrTypeMismatch = errors.New("cache: value type mismatch")
