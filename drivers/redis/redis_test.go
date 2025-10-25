package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shoraid/omnicache"
	redismock "github.com/shoraid/omnicache/drivers/redis/mock"
	"github.com/shoraid/omnicache/internal/assert"
)

func TestRedisStore_NewRedisWithClient2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupClient  func() redisClient
		expectedType string
	}{
		{
			name: "should return RedisStore instance when valid redis client is provided",
			setupClient: func() redisClient {
				return &redismock.MockRedisClient{}
			},
			expectedType: "*redis.RedisStore",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// -- Arrange ---
			client := tt.setupClient()

			// --- Act ---
			store, err := NewRedisWithClient(client)

			// --- Assert ---
			assert.NoError(t, err, "expected no error when creating store")
			assert.NotNil(t, store, "expected store should not be nil")

			// Check if the store really wraps the same client
			_, ok := store.(*RedisStore)
			assert.True(t, ok, "expected store to be *redis.RedisStore type")
		})
	}
}
func TestRedisStore_NewRedisWithClient(t *testing.T) {
	t.Parallel()

	t.Run("should return RedisStore instance when valid redis client is provided", func(t *testing.T) {
		t.Parallel()

		// -- Arrange ---
		client := &redismock.MockRedisClient{}

		// --- Act ---
		store, err := NewRedisWithClient(client)

		// --- Assert ---
		assert.NoError(t, err, "expected no error when creating store")
		assert.NotNil(t, store, "expected store should not be nil")

		// Check if the store really wraps the same client
		_, ok := store.(*RedisStore)
		assert.True(t, ok, "expected store to be *redis.RedisStore type")
	})
}

func TestRedisStore_NewRedisStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  RedisConfig
	}{
		{
			name: "should return RedisStore instance when valid config is provided",
			cfg: RedisConfig{
				Addr: "localhost:6379",
				DB:   0,
			},
		},
		{
			name: "should return RedisStore instance with all config options set",
			cfg: RedisConfig{
				Addr:            "localhost:6379",
				ClientName:      "test-client",
				Username:        "default",
				Password:        "password",
				DB:              1,
				MaxRetries:      5,
				MinRetryBackoff: 10 * time.Millisecond,
				MaxRetryBackoff: 1 * time.Second,
				DialTimeout:     2 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
				PoolSize:        20,
				PoolTimeout:     4 * time.Second,
				MinIdleConns:    5,
				ConnMaxIdleTime: 30 * time.Minute,
				ConnMaxLifetime: 1 * time.Hour,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Act ---
			store, err := NewRedisStore(tt.cfg)

			// --- Assert ---
			assert.NoError(t, err, "expected no error when creating store with valid config")
			assert.NotNil(t, store, "expected store should not be nil")
			_, ok := store.(*RedisStore)
			assert.True(t, ok, "expected store to be *redis.RedisStore type")
		})
	}
}

func TestRedisStore_Clear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mock        func(mock *redismock.MockRedisClient)
		expectedErr error
	}{
		{
			name: "should clear the database successfully when FlushDB succeeds",
			mock: func(mock *redismock.MockRedisClient) {
				mock.FlushDBFunc = func(ctx context.Context) *redis.StatusCmd {
					cmd := redis.NewStatusCmd(ctx)
					cmd.SetVal("OK")
					return cmd
				}
			},
			expectedErr: nil,
		},
		{
			name: "should return an error when FlushDB fails",
			mock: func(mock *redismock.MockRedisClient) {
				mock.FlushDBFunc = func(ctx context.Context) *redis.StatusCmd {
					cmd := redis.NewStatusCmd(ctx)
					cmd.SetErr(errors.New("flushdb error"))
					return cmd
				}
			},
			expectedErr: errors.New("flushdb error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			mock := &redismock.MockRedisClient{}
			store := &RedisStore{client: mock}
			ctx := context.Background()

			tt.mock(mock)

			// --- Act ---
			err := store.Clear(ctx)

			// --- Assert ---
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error when FlushDB fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				return
			}

			assert.NoError(t, err, "expected no error when FlushDB succeeds")
		})
	}
}

func TestRedisStore_Close(t *testing.T) {
	t.Parallel()

	t.Run("should close redis client successfully when client is *redis.Client", func(t *testing.T) {
		t.Parallel()

		// --- Arrange ---
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})

		store := &RedisStore{client: client}
		ctx := context.Background()

		// --- Act ---
		err := store.Close(ctx)

		// --- Assert ---
		assert.NoError(t, err, "expected no error when closing real redis client")
	})

	t.Run("should not panic and return nil when client is not *redis.Client", func(t *testing.T) {
		t.Parallel()

		// --- Arrange ---
		mock := &redismock.MockRedisClient{}
		store := &RedisStore{client: mock}
		ctx := context.Background()

		// --- Act ---
		err := store.Close(ctx)

		// --- Assert ---
		assert.NoError(t, err, "expected no error when client is mock (non-redis.Client)")
	})
}

func TestRedisStore_Delete(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name        string
		mock        func(mock *redismock.MockRedisClient)
		expectedErr error
	}{
		{
			name: "should delete the key successfully when Del succeeds",
			mock: func(mock *redismock.MockRedisClient) {
				mock.DelFunc = func(ctx context.Context, keys ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetVal(1)
					return cmd
				}
			},
			expectedErr: nil,
		},
		{
			name: "should return an error when Del fails",
			mock: func(mock *redismock.MockRedisClient) {
				mock.DelFunc = func(ctx context.Context, keys ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetErr(errors.New("del error"))
					return cmd
				}
			},
			expectedErr: errors.New("del error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			mock := &redismock.MockRedisClient{}
			store := &RedisStore{client: mock}
			ctx := context.Background()

			tt.mock(mock)

			// --- Act ---
			err := store.Delete(ctx, key)

			// --- Assert ---
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error when Del fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				return
			}

			assert.NoError(t, err, "expected no error when Del succeeds")
		})
	}
}

func TestRedisStore_DeleteByPattern(t *testing.T) {
	t.Parallel()

	pattern := "user:*"

	tests := []struct {
		name        string
		mock        func(mock *redismock.MockRedisClient)
		expectedErr error
	}{
		{
			name: "should delete keys matching pattern successfully when Scan and Del succeed",
			mock: func(mock *redismock.MockRedisClient) {
				mock.ScanFunc = func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
					return redis.NewScanCmdResult([]string{"user:1", "user:2"}, 0, nil)
				}
				mock.DelFunc = func(ctx context.Context, keys ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetVal(int64(len(keys)))
					return cmd
				}
			},
			expectedErr: nil,
		},
		{
			name: "should handle no keys matching pattern gracefully",
			mock: func(mock *redismock.MockRedisClient) {
				mock.ScanFunc = func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
					return redis.NewScanCmdResult([]string{}, 0, nil)
				}
			},
			expectedErr: nil,
		},
		{
			name: "should return error when Scan fails",
			mock: func(mock *redismock.MockRedisClient) {
				mock.ScanFunc = func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
					return redis.NewScanCmdResult(nil, 0, errors.New("scan error"))
				}
			},
			expectedErr: errors.New("scan error"),
		},
		{
			name: "should return error when Del fails for one of the keys",
			mock: func(mock *redismock.MockRedisClient) {
				mock.ScanFunc = func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
					return redis.NewScanCmdResult([]string{"user:1", "user:2"}, 0, nil)
				}
				mock.DelFunc = func(ctx context.Context, keys ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					if keys[0] == "user:2" {
						cmd.SetErr(errors.New("del error"))
					} else {
						cmd.SetVal(1)
					}
					return cmd
				}
			},
			expectedErr: errors.New("del error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			mock := &redismock.MockRedisClient{}
			store := &RedisStore{client: mock}
			ctx := context.Background()

			tt.mock(mock)

			// --- Act ---
			err := store.DeleteByPattern(ctx, pattern)

			// --- Assert ---
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error when DeleteByPattern fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				return
			}

			assert.NoError(t, err, "expected no error when DeleteByPattern succeeds")
		})
	}
}

func TestRedisStore_DeleteMany(t *testing.T) {
	t.Parallel()

	keys := []string{"key1", "key2", "key3"}

	tests := []struct {
		name        string
		keys        []string
		mock        func(mock *redismock.MockRedisClient)
		expectedErr error
	}{
		{
			name: "should delete multiple keys successfully when Del succeeds",
			keys: keys,
			mock: func(mock *redismock.MockRedisClient) {
				mock.DelFunc = func(ctx context.Context, k ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetVal(int64(len(k)))
					return cmd
				}
			},
			expectedErr: nil,
		},
		{
			name: "should do nothing when no keys are provided",
			keys: []string{},
			mock: func(mock *redismock.MockRedisClient) {
				// no redis call expected
				mock.DelFunc = func(ctx context.Context, k ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetVal(0)
					return cmd
				}
			},
			expectedErr: nil,
		},
		{
			name: "should return an error when Del fails",
			keys: keys,
			mock: func(mock *redismock.MockRedisClient) {
				mock.DelFunc = func(ctx context.Context, k ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetErr(errors.New("del many error"))
					return cmd
				}
			},
			expectedErr: errors.New("del many error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			mock := &redismock.MockRedisClient{}
			store := &RedisStore{client: mock}
			ctx := context.Background()

			tt.mock(mock)

			// --- Act ---
			err := store.DeleteMany(ctx, tt.keys...)

			// --- Assert ---
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error when DeleteMany fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				return
			}

			assert.NoError(t, err, "expected no error when DeleteMany succeeds")
		})
	}
}

func TestRedisStore_Get(t *testing.T) {
	t.Parallel()

	key := "test-key"
	value := `"test-value"` // JSON marshaled string

	tests := []struct {
		name        string
		mock        func(mock *redismock.MockRedisClient)
		expectedVal any
		expectedErr error
	}{
		{
			name: "should return the value successfully when Get succeeds",
			mock: func(mock *redismock.MockRedisClient) {
				mock.GetFunc = func(ctx context.Context, k string) *redis.StringCmd {
					cmd := redis.NewStringCmd(ctx)
					cmd.SetVal(value)
					return cmd
				}
			},
			expectedVal: value,
			expectedErr: nil,
		},
		{
			name: "should return ErrCacheMiss when key does not exist (redis.Nil)",
			mock: func(mock *redismock.MockRedisClient) {
				mock.GetFunc = func(ctx context.Context, k string) *redis.StringCmd {
					cmd := redis.NewStringCmd(ctx)
					cmd.SetErr(redis.Nil)
					return cmd
				}
			},
			expectedVal: nil,
			expectedErr: omnicache.ErrCacheMiss,
		},
		{
			name: "should return an error when Get fails with other error",
			mock: func(mock *redismock.MockRedisClient) {
				mock.GetFunc = func(ctx context.Context, k string) *redis.StringCmd {
					cmd := redis.NewStringCmd(ctx)
					cmd.SetErr(errors.New("get error"))
					return cmd
				}
			},
			expectedVal: nil,
			expectedErr: errors.New("get error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			mock := &redismock.MockRedisClient{}
			store := &RedisStore{client: mock}
			ctx := context.Background()

			tt.mock(mock)

			// --- Act ---
			result, err := store.Get(ctx, key)

			// --- Assert ---
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error when Get fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				assert.Nil(t, result, "result must be nil on error")
				return
			}

			assert.NoError(t, err, "expected no error when Get succeeds")
			assert.Equal(t, tt.expectedVal, result, "result must match the expected value")
		})
	}
}

func TestRedisStore_Has(t *testing.T) {
	t.Parallel()

	key := "test-key"

	tests := []struct {
		name        string
		mock        func(mock *redismock.MockRedisClient)
		expectedVal bool
		expectedErr error
	}{
		{
			name: "should return true when key exists",
			mock: func(mock *redismock.MockRedisClient) {
				mock.ExistsFunc = func(ctx context.Context, keys ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetVal(1)
					return cmd
				}
			},
			expectedVal: true,
			expectedErr: nil,
		},
		{
			name: "should return false when key does not exist",
			mock: func(mock *redismock.MockRedisClient) {
				mock.ExistsFunc = func(ctx context.Context, keys ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetVal(0)
					return cmd
				}
			},
			expectedVal: false,
			expectedErr: nil,
		},
		{
			name: "should return an error when Exists fails",
			mock: func(mock *redismock.MockRedisClient) {
				mock.ExistsFunc = func(ctx context.Context, keys ...string) *redis.IntCmd {
					cmd := redis.NewIntCmd(ctx)
					cmd.SetErr(errors.New("exists error"))
					return cmd
				}
			},
			expectedVal: false,
			expectedErr: errors.New("exists error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			mock := &redismock.MockRedisClient{}
			store := &RedisStore{client: mock}
			ctx := context.Background()

			tt.mock(mock)

			// --- Act ---
			result, err := store.Has(ctx, key)

			// --- Assert ---
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error when Has fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				assert.Equal(t, tt.expectedVal, result, "result must be false on error")
				return
			}

			assert.NoError(t, err, "expected no error when Has succeeds")
			assert.Equal(t, tt.expectedVal, result, "result must match the expected value")
		})
	}
}

func TestRedisStore_Set(t *testing.T) {
	t.Parallel()

	key := "test-key"
	value := "test-value"
	marshaledValue, _ := json.Marshal(value)
	ttl := 5 * time.Minute

	tests := []struct {
		name        string
		key         string
		value       any
		ttl         time.Duration
		mock        func(mock *redismock.MockRedisClient)
		expectedErr error
	}{
		{
			name:  "should set the value successfully when Set succeeds",
			key:   key,
			value: value,
			ttl:   ttl,
			mock: func(mock *redismock.MockRedisClient) {
				mock.SetFunc = func(ctx context.Context, k string, v any, exp time.Duration) *redis.StatusCmd {
					cmd := redis.NewStatusCmd(ctx)
					// verify the value matches expected JSON
					assert.Equal(t, marshaledValue, v, "expected marshaled value to match")
					cmd.SetVal("OK")
					return cmd
				}
			},
			expectedErr: nil,
		},
		{
			name:  "should handle zero TTL correctly (no expiration)",
			key:   key,
			value: value,
			ttl:   0,
			mock: func(mock *redismock.MockRedisClient) {
				mock.SetFunc = func(ctx context.Context, k string, v any, exp time.Duration) *redis.StatusCmd {
					cmd := redis.NewStatusCmd(ctx)
					assert.Equal(t, time.Duration(0), exp, "expected zero TTL")
					cmd.SetVal("OK")
					return cmd
				}
			},
			expectedErr: nil,
		},
		{
			name:  "should handle complex struct values",
			key:   key,
			value: struct{ Name string }{Name: "test"},
			ttl:   ttl,
			mock: func(mock *redismock.MockRedisClient) {
				data, _ := json.Marshal(struct{ Name string }{Name: "test"})
				mock.SetFunc = func(ctx context.Context, k string, v any, exp time.Duration) *redis.StatusCmd {
					cmd := redis.NewStatusCmd(ctx)
					assert.Equal(t, data, v, "expected marshaled struct value to match")
					cmd.SetVal("OK")
					return cmd
				}
			},
			expectedErr: nil,
		},
		{
			name:  "should return an error when Set fails",
			key:   key,
			value: value,
			ttl:   ttl,
			mock: func(mock *redismock.MockRedisClient) {
				mock.SetFunc = func(ctx context.Context, k string, v any, exp time.Duration) *redis.StatusCmd {
					cmd := redis.NewStatusCmd(ctx)
					cmd.SetErr(errors.New("set error"))
					return cmd
				}
			},
			expectedErr: errors.New("set error"),
		},
		{
			name:  "should return ErrInvalidValue when TTL is negative",
			key:   key,
			value: value,
			ttl:   -1 * time.Minute,
			mock: func(mock *redismock.MockRedisClient) {
				// No Redis call expected
			},
			expectedErr: omnicache.ErrInvalidValue,
		},
		{
			name:  "should return an error when value cannot be marshaled",
			key:   key,
			value: make(chan int), // cannot be marshaled
			ttl:   ttl,
			mock: func(mock *redismock.MockRedisClient) {
				// no redis call expected since marshalling fails
			},
			expectedErr: errors.New("json: unsupported type: chan int"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &redismock.MockRedisClient{}
			store := &RedisStore{client: mock}
			ctx := context.Background()

			tt.mock(mock)

			err := store.Set(ctx, tt.key, tt.value, tt.ttl)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error when Set fails")
				assert.EqualError(t, tt.expectedErr, err, "error must match the expected error")
				return
			}

			assert.NoError(t, err, "expected no error when Set succeeds")
		})
	}
}
