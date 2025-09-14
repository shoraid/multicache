package redis

import (
	"context"
	"testing"
	"time"

	"github.com/shoraid/multicache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRedisTestClient creates a new Redis client for testing.
// It requires a running Redis instance at the default address (localhost:6379).
func setupRedisTestClient(t *testing.T) *RedisStore {
	config := RedisConfig{
		Addr:     "localhost:6379",
		Password: "", // No password by default
		DB:       1,  // Use DB 1 for testing to avoid conflicts
	}
	store, err := NewRedisStore(config)
	require.NoError(t, err, "failed to create Redis store")

	// Ping the Redis server to ensure connection
	redisStore, ok := store.(*RedisStore)
	require.True(t, ok, "store is not a RedisStore")
	err = redisStore.client.Ping(context.Background()).Err()
	require.NoError(t, err, "failed to connect to Redis, ensure Redis is running on localhost:6379")

	// Clear the test DB before each test
	err = redisStore.Clear(context.Background())
	require.NoError(t, err, "failed to clear Redis DB")

	return redisStore
}

func TestNewRedisStore(t *testing.T) {
	config := RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	store, err := NewRedisStore(config)
	assert.NoError(t, err)
	assert.NotNil(t, store)

	redisStore, ok := store.(*RedisStore)
	assert.True(t, ok)
	assert.NotNil(t, redisStore.client)

	// Test connection by pinging
	err = redisStore.client.Ping(context.Background()).Err()
	assert.NoError(t, err, "failed to ping Redis, ensure it's running")
}

func TestRedisStore_Clear(t *testing.T) {
	store := setupRedisTestClient(t)
	ctx := context.Background()

	// Add some data
	err := store.Set(ctx, "key1", "value1", 0)
	require.NoError(t, err)
	err = store.Set(ctx, "key2", "value2", 0)
	require.NoError(t, err)

	// Verify data exists
	val1, err := store.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val1)

	// Clear the store
	err = store.Clear(ctx)
	assert.NoError(t, err)

	// Verify data is cleared
	_, err = store.Get(ctx, "key1")
	assert.ErrorIs(t, err, multicache.ErrCacheMiss)
	_, err = store.Get(ctx, "key2")
	assert.ErrorIs(t, err, multicache.ErrCacheMiss)
}

func TestRedisStore_Delete(t *testing.T) {
	store := setupRedisTestClient(t)
	ctx := context.Background()

	key := "test_key"
	value := "test_value"

	// Set a key
	err := store.Set(ctx, key, value, 0)
	require.NoError(t, err)

	// Verify it exists
	val, err := store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, val)

	// Delete the key
	err = store.Delete(ctx, key)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = store.Get(ctx, key)
	assert.ErrorIs(t, err, multicache.ErrCacheMiss)

	// Deleting a non-existent key should not return an error
	err = store.Delete(ctx, "non_existent_key")
	assert.NoError(t, err)
}

func TestRedisStore_DeleteByPattern(t *testing.T) {
	store := setupRedisTestClient(t)
	ctx := context.Background()

	// Setup test data
	store.Set(ctx, "user:1:profile", "data1", 0)
	store.Set(ctx, "user:1:settings", "data2", 0)
	store.Set(ctx, "user:2:profile", "data3", 0)
	store.Set(ctx, "product:100:details", "data4", 0)

	t.Run("should delete keys matching a pattern", func(t *testing.T) {
		err := store.DeleteByPattern(ctx, "user:1:*")
		assert.NoError(t, err)

		_, err = store.Get(ctx, "user:1:profile")
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)
		_, err = store.Get(ctx, "user:1:settings")
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)

		// Other keys should remain
		_, err = store.Get(ctx, "user:2:profile")
		assert.NoError(t, err)
		_, err = store.Get(ctx, "product:100:details")
		assert.NoError(t, err)
	})

	t.Run("should handle pattern with no matches", func(t *testing.T) {
		// Clear previous state
		store.Clear(ctx)
		store.Set(ctx, "key1", "val1", 0)

		err := store.DeleteByPattern(ctx, "nonexistent:*")
		assert.NoError(t, err)

		// Key should still exist
		_, err = store.Get(ctx, "key1")
		assert.NoError(t, err)
	})

	t.Run("should delete all keys with '*' pattern", func(t *testing.T) {
		store.Clear(ctx)
		store.Set(ctx, "keyA", "valA", 0)
		store.Set(ctx, "keyB", "valB", 0)

		err := store.DeleteByPattern(ctx, "*")
		assert.NoError(t, err)

		_, err = store.Get(ctx, "keyA")
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)
		_, err = store.Get(ctx, "keyB")
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)
	})
}

func TestRedisStore_DeleteMany(t *testing.T) {
	store := setupRedisTestClient(t)
	ctx := context.Background()

	// Setup test data
	store.Set(ctx, "key1", "value1", 0)
	store.Set(ctx, "key2", "value2", 0)
	store.Set(ctx, "key3", "value3", 0)
	store.Set(ctx, "key4", "value4", 0)

	t.Run("should delete multiple existing keys", func(t *testing.T) {
		err := store.DeleteMany(ctx, "key1", "key2")
		assert.NoError(t, err)

		_, err = store.Get(ctx, "key1")
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)
		_, err = store.Get(ctx, "key2")
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)

		// Other keys should remain
		_, err = store.Get(ctx, "key3")
		assert.NoError(t, err)
		_, err = store.Get(ctx, "key4")
		assert.NoError(t, err)
	})

	t.Run("should handle mix of existing and non-existing keys", func(t *testing.T) {
		err := store.DeleteMany(ctx, "key3", "nonexistent_key")
		assert.NoError(t, err)

		_, err = store.Get(ctx, "key3")
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)

		// Other keys should remain
		_, err = store.Get(ctx, "key4")
		assert.NoError(t, err)
	})

	t.Run("should handle empty list of keys", func(t *testing.T) {
		err := store.DeleteMany(ctx)
		assert.NoError(t, err)
		// No keys should be deleted, so key4 should still exist
		_, err = store.Get(ctx, "key4")
		assert.NoError(t, err)
	})
}

func TestRedisStore_Get(t *testing.T) {
	store := setupRedisTestClient(t)
	ctx := context.Background()

	t.Run("should return value for existing key", func(t *testing.T) {
		key := "get_key"
		value := "get_value"
		err := store.Set(ctx, key, value, 0)
		require.NoError(t, err)

		retrievedValue, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)
	})

	t.Run("should return ErrCacheMiss for non-existing key", func(t *testing.T) {
		_, err := store.Get(ctx, "non_existent_key")
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)
	})

	t.Run("should return ErrCacheMiss for expired key", func(t *testing.T) {
		key := "expired_key"
		value := "expired_value"
		err := store.Set(ctx, key, value, 1*time.Millisecond) // Set with a very short TTL
		require.NoError(t, err)

		time.Sleep(5 * time.Millisecond) // Wait for the key to expire

		_, err = store.Get(ctx, key)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)
	})
}

func TestRedisStore_GetOrSet(t *testing.T) {
	store := setupRedisTestClient(t)
	ctx := context.Background()

	t.Run("should return existing value if key is present", func(t *testing.T) {
		key := "getorset_existing"
		existingValue := "old_value"
		newValue := "new_value"
		ttl := 1 * time.Hour

		err := store.Set(ctx, key, existingValue, ttl)
		require.NoError(t, err)

		retrievedValue, err := store.GetOrSet(ctx, key, ttl, newValue)
		assert.NoError(t, err)
		assert.Equal(t, existingValue, retrievedValue) // Should return the existing value
	})

	t.Run("should set and return new value if key is missing", func(t *testing.T) {
		key := "getorset_missing"
		newValue := "new_value"
		ttl := 1 * time.Hour

		retrievedValue, err := store.GetOrSet(ctx, key, ttl, newValue)
		assert.NoError(t, err)
		assert.Equal(t, newValue, retrievedValue)

		// Verify it was actually set
		storedValue, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, newValue, storedValue)
	})

	t.Run("should set and return new value if key is expired", func(t *testing.T) {
		key := "getorset_expired"
		oldValue := "old_value"
		newValue := "new_value"
		shortTTL := 1 * time.Millisecond
		longTTL := 1 * time.Hour

		err := store.Set(ctx, key, oldValue, shortTTL)
		require.NoError(t, err)
		time.Sleep(5 * time.Millisecond) // Wait for expiration

		retrievedValue, err := store.GetOrSet(ctx, key, longTTL, newValue)
		assert.NoError(t, err)
		assert.Equal(t, newValue, retrievedValue)

		// Verify it was actually set with the new value and TTL
		storedValue, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, newValue, storedValue)
	})

	t.Run("should return error if Set fails (e.g., invalid TTL)", func(t *testing.T) {
		key := "getorset_invalid_ttl"
		newValue := "new_value"
		invalidTTL := -1 * time.Second //
		_, err := store.GetOrSet(ctx, key, invalidTTL, newValue)
		assert.ErrorIs(t, err, multicache.ErrInvalidValue)
	})
}

func TestRedisStore_Has(t *testing.T) {
	store := setupRedisTestClient(t)
	ctx := context.Background()

	t.Run("should return true for existing key", func(t *testing.T) {
		key := "has_key"
		err := store.Set(ctx, key, "value", 0)
		require.NoError(t, err)

		exists, err := store.Has(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return false for non-existing key", func(t *testing.T) {
		exists, err := store.Has(ctx, "non_existent_key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should return false for expired key", func(t *testing.T) {
		key := "has_expired_key"
		err := store.Set(ctx, key, "value", 1*time.Millisecond)
		require.NoError(t, err)

		time.Sleep(5 * time.Millisecond) // Wait for expiration

		exists, err := store.Has(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestRedisStore_Set(t *testing.T) {
	store := setupRedisTestClient(t)
	ctx := context.Background()

	t.Run("should set a new key with no expiration", func(t *testing.T) {
		key := "set_key_no_ttl"
		value := "set_value_no_ttl"
		err := store.Set(ctx, key, value, 0)
		assert.NoError(t, err)

		retrievedValue, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)

		// Verify TTL is -1 (no expiration)
		ttl, err := store.client.TTL(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-1), ttl)
	})

	t.Run("should set a new key with expiration", func(t *testing.T) {
		key := "set_key_with_ttl"
		value := "set_value_with_ttl"
		ttl := 5 * time.Second
		err := store.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		retrievedValue, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)

		// Verify TTL is set correctly (with some tolerance)
		actualTTL, err := store.client.TTL(ctx, key).Result()
		assert.NoError(t, err)
		assert.InDelta(t, ttl.Seconds(), actualTTL.Seconds(), 1.0) // Allow 1 second difference
	})

	t.Run("should overwrite existing key", func(t *testing.T) {
		key := "overwrite_key"
		initialValue := "initial"
		newValue := "new_value"

		err := store.Set(ctx, key, initialValue, 0)
		require.NoError(t, err)

		err = store.Set(ctx, key, newValue, 0)
		assert.NoError(t, err)

		retrievedValue, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, newValue, retrievedValue)
	})

	t.Run("should return error for negative TTL", func(t *testing.T) {
		key := "invalid_ttl_key"
		value := "invalid_ttl_value"
		invalidTTL := -1 * time.Second

		err := store.Set(ctx, key, value, invalidTTL)
		assert.ErrorIs(t, err, multicache.ErrInvalidValue)

		// Ensure key was not set
		_, err = store.Get(ctx, key)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss)
	})
}
