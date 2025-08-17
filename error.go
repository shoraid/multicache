package multicache

import "errors"

var ErrCacheMiss = errors.New("cache: cache miss")
var ErrItemAlreadyExists = errors.New("cache: item already exists")
var ErrTypeMismatch = errors.New("cache: value type mismatch")
