package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shoraid/multicache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisStore_NewRedisStore(t *testing.T) {
	// Backup original buildTLSConfig function so we can restore after test
	origBuildTLSConfig := buildTLSConfig
	defer func() { buildTLSConfig = origBuildTLSConfig }()

	tests := []struct {
		name       string
		cfg        RedisConfig
		mockTLS    func()
		wantErr    bool
		errMessage string
	}{
		{
			name: "should create store without TLS when UseTLS is false",
			cfg: RedisConfig{
				Addr:        "localhost:6379",
				DB:          0,
				UseTLS:      false,
				PoolSize:    5,
				PoolTimeout: 1 * time.Second,
			},
			mockTLS: func() {}, // no override
			wantErr: false,
		},
		{
			name: "should create store with TLS when UseTLS is true and TLSConfig is valid",
			cfg: RedisConfig{
				Addr:   "localhost:6379",
				DB:     0,
				UseTLS: true,
				TLSConfig: &TLSConfig{
					ServerName: "localhost",
					MinVersion: tls.VersionTLS12,
				},
			},
			mockTLS: func() {
				buildTLSConfig = func(cfg *TLSConfig) (*tls.Config, error) {
					return &tls.Config{ServerName: "localhost"}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "should return error when TLSConfig build fails",
			cfg: RedisConfig{
				Addr:   "localhost:6379",
				DB:     0,
				UseTLS: true,
				TLSConfig: &TLSConfig{
					ServerName: "bad",
				},
			},
			mockTLS: func() {
				buildTLSConfig = func(cfg *TLSConfig) (*tls.Config, error) {
					return nil, errors.New("mock: failed to build TLS")
				}
			},
			wantErr:    true,
			errMessage: "expected error when TLS build fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockTLS()
			store, err := NewRedisStore(tt.cfg)

			if tt.wantErr {
				require.Error(t, err, "expected an error but got nil")
				assert.Nil(t, store, "expected store to be nil")
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), "failed to build TLS config", tt.errMessage)
				}
				return
			}

			require.NoError(t, err, "expected no error but got one")
			assert.NotNil(t, store, "expected store to be created")
		})
	}
}

func TestRedisStore_Clear(t *testing.T) {
	tests := []struct {
		name       string
		mockFunc   func(ctx context.Context) *redis.StatusCmd
		wantErr    bool
		errMessage string
	}{
		{
			name: "should return no error when FlushDB succeeds",
			mockFunc: func(ctx context.Context) *redis.StatusCmd {
				cmd := redis.NewStatusCmd(ctx)
				cmd.SetVal("OK")
				return cmd
			},
			wantErr: false,
		},
		{
			name: "should return error when FlushDB fails",
			mockFunc: func(ctx context.Context) *redis.StatusCmd {
				cmd := redis.NewStatusCmd(ctx)
				cmd.SetErr(errors.New("mock: flush failed"))
				return cmd
			},
			wantErr:    true,
			errMessage: "expected error when FlushDB fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &RedisStore{
				client: &mockRedisClient{
					flushDBFunc: tt.mockFunc,
				},
			}

			err := store.Clear(context.Background())

			if tt.wantErr {
				assert.Error(t, err, tt.errMessage)
				assert.Contains(t, err.Error(), "flush failed", "expected error message to contain flush failed")
			} else {
				assert.NoError(t, err, "expected no error but got one")
			}
		})
	}
}

func TestRedisStore_Delete(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		mockFunc   func(ctx context.Context, keys ...string) *redis.IntCmd
		wantErr    bool
		errMessage string
	}{
		{
			name: "should delete key successfully when Del returns no error",
			key:  "foo",
			mockFunc: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(1) // one key deleted
				return cmd
			},
			wantErr: false,
		},
		{
			name: "should return error when Del fails",
			key:  "bar",
			mockFunc: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetErr(errors.New("mock: delete failed"))
				return cmd
			},
			wantErr:    true,
			errMessage: "expected error when Del fails",
		},
		{
			name: "should not fail even when key does not exist",
			key:  "missing",
			mockFunc: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(0) // no keys deleted
				return cmd
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &RedisStore{
				client: &mockRedisClient{delFunc: tt.mockFunc},
			}

			err := store.Delete(context.Background(), tt.key)

			if tt.wantErr {
				assert.Error(t, err, tt.errMessage)
				assert.Contains(t, err.Error(), "delete failed", "expected error message to contain delete failed")
			} else {
				assert.NoError(t, err, "expected no error but got one")
			}
		})
	}
}

func TestRedisStore_DeleteByPattern(t *testing.T) {
	// backup and restore the seam
	origFactory := newScanIterator
	defer func() { newScanIterator = origFactory }()

	tests := []struct {
		name      string
		pattern   string
		iter      *mockIter
		mockDel   func(ctx context.Context, keys ...string) *redis.IntCmd
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "should return no error when no keys found",
			pattern: "no:*",
			iter:    &mockIter{keys: []string{}},
			mockDel: func(ctx context.Context, keys ...string) *redis.IntCmd {
				// should not be called, but return OK if it is
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(0)
				return cmd
			},
			wantErr: false,
		},
		{
			name:    "should delete all matching keys without error",
			pattern: "ok:*",
			iter:    &mockIter{keys: []string{"k1", "k2", "k3"}},
			mockDel: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(int64(len(keys))) // simulate success
				return cmd
			},
			wantErr: false,
		},
		{
			name:    "should return error when Del fails on a key",
			pattern: "bad:*",
			iter:    &mockIter{keys: []string{"k1", "k2"}},
			mockDel: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetErr(errors.New("delete failed"))
				return cmd
			},
			wantErr:   true,
			errSubstr: "delete failed",
		},
		{
			name:    "should return iterator error when iteration ends with error",
			pattern: "iter-err:*",
			iter:    &mockIter{keys: []string{"k1"}, err: errors.New("scan iterator error")},
			mockDel: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(1)
				return cmd
			},
			wantErr:   true,
			errSubstr: "scan iterator error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// override seam to return our mock iterator
			newScanIterator = func(ctx context.Context, c redisClient, pattern string) scanIterator {
				return tt.iter
			}

			store := &RedisStore{
				client: &mockRedisClient{
					delFunc: tt.mockDel,
				},
			}

			err := store.DeleteByPattern(context.Background(), tt.pattern)

			if tt.wantErr {
				assert.Error(t, err, "expected an error")
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr, "expected error message to contain "+tt.errSubstr)
				}
			} else {
				assert.NoError(t, err, "expected no error but got one")
			}
		})
	}
}

func TestRedisStore_DeleteMany(t *testing.T) {
	tests := []struct {
		name       string
		keys       []string
		mockDel    func(ctx context.Context, keys ...string) *redis.IntCmd
		wantErr    bool
		errMessage string
	}{
		{
			name: "should return nil when no keys provided",
			keys: []string{},
			mockDel: func(ctx context.Context, keys ...string) *redis.IntCmd {
				t.Fatal("expected Del not to be called when keys are empty")
				return nil
			},
			wantErr: false,
		},
		{
			name: "should delete all keys successfully when Del succeeds",
			keys: []string{"key1", "key2"},
			mockDel: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(int64(len(keys))) // simulate success
				return cmd
			},
			wantErr: false,
		},
		{
			name: "should return error when Del fails",
			keys: []string{"badkey"},
			mockDel: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetErr(errors.New("mock: delete failed"))
				return cmd
			},
			wantErr:    true,
			errMessage: "expected error when Del fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &RedisStore{
				client: &mockRedisClient{
					delFunc: tt.mockDel,
				},
			}

			err := store.DeleteMany(context.Background(), tt.keys...)

			if tt.wantErr {
				assert.Error(t, err, tt.errMessage)
				assert.Contains(t, err.Error(), "delete failed", "expected error message to contain delete failed")
			} else {
				assert.NoError(t, err, "expected no error but got one")
			}
		})
	}
}

func TestRedisStore_Get(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		mockGet    func(ctx context.Context, key string) *redis.StringCmd
		wantValue  any
		wantErr    error
		errMessage string
	}{
		{
			name: "should return value when key exists",
			key:  "foo",
			mockGet: func(ctx context.Context, key string) *redis.StringCmd {
				cmd := redis.NewStringCmd(ctx)
				cmd.SetVal("bar")
				return cmd
			},
			wantValue: "bar",
			wantErr:   nil,
		},
		{
			name: "should return ErrCacheMiss when key does not exist",
			key:  "missing",
			mockGet: func(ctx context.Context, key string) *redis.StringCmd {
				cmd := redis.NewStringCmd(ctx)
				cmd.SetErr(redis.Nil)
				return cmd
			},
			wantValue: nil,
			wantErr:   multicache.ErrCacheMiss,
		},
		{
			name: "should return error when Get fails with other error",
			key:  "foo",
			mockGet: func(ctx context.Context, key string) *redis.StringCmd {
				cmd := redis.NewStringCmd(ctx)
				cmd.SetErr(errors.New("mock: network error"))
				return cmd
			},
			wantValue:  nil,
			wantErr:    errors.New("mock: network error"),
			errMessage: "expected error when Get fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &RedisStore{
				client: &mockRedisClient{
					getFunc: tt.mockGet,
				},
			}

			val, err := store.Get(context.Background(), tt.key)

			if tt.wantErr != nil {
				assert.Error(t, err, tt.errMessage)
				assert.Equal(t, tt.wantErr.Error(), err.Error(), "expected error to match")
				assert.Nil(t, val, "expected value to be nil on error")
			} else {
				assert.NoError(t, err, "expected no error but got one")
				assert.Equal(t, tt.wantValue, val, "expected returned value to match")
			}
		})
	}
}

func TestRedisStore_Has(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		mockExists func(ctx context.Context, keys ...string) *redis.IntCmd
		want       bool
		wantErr    bool
		errMessage string
	}{
		{
			name: "should return true when key exists",
			key:  "foo",
			mockExists: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(1) // simulate key found
				return cmd
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "should return false when key does not exist",
			key:  "missing",
			mockExists: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(0) // simulate key missing
				return cmd
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "should return error when Exists fails",
			key:  "foo",
			mockExists: func(ctx context.Context, keys ...string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetErr(errors.New("mock: exists failed"))
				return cmd
			},
			want:       false,
			wantErr:    true,
			errMessage: "expected error when Exists fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &RedisStore{
				client: &mockRedisClient{
					existsFunc: tt.mockExists,
				},
			}

			has, err := store.Has(context.Background(), tt.key)

			if tt.wantErr {
				assert.Error(t, err, tt.errMessage)
				assert.Contains(t, err.Error(), "exists failed", "expected error message to contain exists failed")
				assert.False(t, has, "expected has to be false on error")
			} else {
				assert.NoError(t, err, "expected no error but got one")
				assert.Equal(t, tt.want, has, "expected has to match")
			}
		})
	}
}

func TestRedisStore_Set(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       any
		ttl         time.Duration
		mockSet     func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
		wantErr     bool
		errContains string
	}{
		{
			name:  "should set value successfully with positive TTL",
			key:   "foo",
			value: "bar",
			ttl:   10 * time.Second,
			mockSet: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
				cmd := redis.NewStatusCmd(ctx)
				cmd.SetVal("OK")
				return cmd
			},
			wantErr: false,
		},
		{
			name:  "should set value successfully with 0 TTL (no expiration)",
			key:   "foo",
			value: "bar",
			ttl:   0,
			mockSet: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
				cmd := redis.NewStatusCmd(ctx)
				cmd.SetVal("OK")
				return cmd
			},
			wantErr: false,
		},
		{
			name:  "should return error when TTL is negative",
			key:   "foo",
			value: "bar",
			ttl:   -1 * time.Second,
			mockSet: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
				t.Fatal("expected Set not to be called with negative TTL")
				return nil
			},
			wantErr:     true,
			errContains: multicache.ErrInvalidValue.Error(),
		},
		{
			name:  "should return error when Set fails",
			key:   "foo",
			value: "bar",
			ttl:   10 * time.Second,
			mockSet: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
				cmd := redis.NewStatusCmd(ctx)
				cmd.SetErr(errors.New("mock: set command failed"))
				return cmd
			},
			wantErr:     true,
			errContains: "set command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &RedisStore{
				client: &mockRedisClient{
					setFunc: tt.mockSet,
				},
			}

			err := store.Set(context.Background(), tt.key, tt.value, tt.ttl)

			if tt.wantErr {
				assert.Error(t, err, "expected error but got none")
				assert.Contains(t, err.Error(), tt.errContains, "expected error message to contain "+tt.errContains)
			} else {
				assert.NoError(t, err, "expected no error but got one")
			}
		})
	}
}
