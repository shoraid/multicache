package multicache

import "errors"

var ErrCacheMiss = errors.New("cache: cache miss")
var ErrInvalidValue = errors.New("cache: invalid value")
var ErrTypeMismatch = errors.New("cache: value type mismatch")
