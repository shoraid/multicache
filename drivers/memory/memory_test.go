package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shoraid/multicache"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryStore(t *testing.T) {
	t.Run("should create store with default cleanup interval when config is empty", func(t *testing.T) {
		storeInterface, err := NewMemoryStore(map[string]any{})
		assert.NoError(t, err, "expected no error when creating store with empty config")

		store, ok := storeInterface.(*MemoryStore)
		assert.True(t, ok, "expected returned store to be of type *MemoryStore")
		assert.NotNil(t, store.cancelCleanup, "expected cancelCleanup function to be initialized")
	})

	t.Run("should use provided valid cleanup_interval from config", func(t *testing.T) {
		interval := 50 * time.Millisecond
		storeInterface, err := NewMemoryStore(map[string]any{
			"cleanup_interval": interval,
		})
		assert.NoError(t, err, "expected no error when creating store with valid cleanup_interval")

		store := storeInterface.(*MemoryStore)
		assert.NotNil(t, store.cancelCleanup, "expected cancelCleanup function to be initialized")
	})

	t.Run("should ignore invalid cleanup_interval values", func(t *testing.T) {
		storeInterface, err := NewMemoryStore(map[string]any{
			"cleanup_interval": "invalid",
		})
		assert.NoError(t, err, "expected no error when creating store with invalid cleanup_interval type")

		store := storeInterface.(*MemoryStore)
		assert.NotNil(t, store.cancelCleanup, "expected cancelCleanup function to be initialized even with invalid interval")
	})

	t.Run("should delete expired keys automatically", func(t *testing.T) {
		storeInterface, err := NewMemoryStore(map[string]any{
			"cleanup_interval": 10 * time.Millisecond,
		})
		assert.NoError(t, err, "expected no error when creating store with short cleanup interval")
		store := storeInterface.(*MemoryStore)

		// Add expired key
		store.data.Store("expired", memoryItem{
			value:      1,
			expiration: time.Now().Add(-1 * time.Second),
		})

		time.Sleep(20 * time.Millisecond) // wait for cleanup
		_, exists := store.data.Load("expired")
		assert.False(t, exists, "expected expired key to be deleted by cleanupExpiredKeys goroutine")

		// Cancel cleanup goroutine
		store.cancelCleanup()
	})

	t.Run("should allow multiple independent stores", func(t *testing.T) {
		store1, _ := NewMemoryStore(map[string]any{})
		store2, _ := NewMemoryStore(map[string]any{})

		s1 := store1.(*MemoryStore)
		s2 := store2.(*MemoryStore)

		s1.data.Store("key", memoryItem{
			value:      1,
			expiration: time.Now().Add(1 * time.Hour),
		})
		s2.data.Store("key", memoryItem{
			value:      2,
			expiration: time.Now().Add(1 * time.Hour),
		})

		item1, _ := s1.data.Load("key")
		actual1, _ := item1.(memoryItem)

		item2, _ := s2.data.Load("key")
		actual2, _ := item2.(memoryItem)

		assert.Equal(t, 1, actual1.value, "expected store1 to return its own value independently")
		assert.Equal(t, 2, actual2.value, "expected store2 to return its own value independently")

		s1.cancelCleanup()
		s2.cancelCleanup()
	})
}

func TestMemoryStore_cleanupExpiredKeys(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	t.Run("should stop immediately when context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			store.cleanupExpiredKeys(ctx, 100*time.Millisecond)
			close(done)
		}()

		cancel() // cancel context immediately

		select {
		case <-done:
			// success
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected cleanupExpiredKeys to stop after context is canceled")
		}
	})

	t.Run("should delete expired keys at intervals", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// setup keys
		store.data = sync.Map{}
		store.data.Store("key1", memoryItem{value: 1, expiration: time.Now().Add(-1 * time.Minute)}) // expired
		store.data.Store("key2", memoryItem{value: 2, expiration: time.Now().Add(1 * time.Hour)})    // not expired

		go store.cleanupExpiredKeys(ctx, 50*time.Millisecond)

		time.Sleep(100 * time.Millisecond)

		// check key1 is deleted
		_, ok1 := store.data.Load("key1")
		assert.False(t, ok1, "expected expired key1 to be deleted during cleanup")

		// check key2 still exists
		_, ok2 := store.data.Load("key2")
		assert.True(t, ok2, "expected non-expired key2 to remain during cleanup")

		cancel()
	})

	t.Run("should handle store with no expired keys", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// setup keys
		store.data = sync.Map{}
		store.data.Store("key1", memoryItem{value: 1, expiration: time.Now().Add(1 * time.Minute)})
		store.data.Store("key2", memoryItem{value: 2, expiration: time.Now().Add(2 * time.Minute)})

		go store.cleanupExpiredKeys(ctx, 50*time.Millisecond)
		time.Sleep(100 * time.Millisecond)

		// count keys
		count := 0
		store.data.Range(func(_, _ any) bool {
			count++
			return true
		})

		assert.Equal(t, 2, count, "expected all keys to remain because none are expired")
		cancel()
	})
}

func collectKeys(m *sync.Map) []string {
	keys := []string{}
	m.Range(func(k, _ any) bool {
		keys = append(keys, k.(string))
		return true
	})
	return keys
}

func TestMemoryStore_deleteExpiredKeys(t *testing.T) {

	store := &MemoryStore{
		data: sync.Map{},
	}

	t.Run("should do nothing when store is empty", func(t *testing.T) {
		store.data = sync.Map{} // reset

		store.deleteExpiredKeys()

		count := 0
		store.data.Range(func(_, _ any) bool {
			count++
			return true
		})
		assert.Equal(t, 0, count, "expected store to remain empty when no keys exist")
	})

	t.Run("should keep all keys when none are expired", func(t *testing.T) {
		store.data = sync.Map{}
		store.data.Store("key1", memoryItem{value: 1, expiration: time.Now().Add(1 * time.Hour)})
		store.data.Store("key2", memoryItem{value: 2, expiration: time.Now().Add(10 * time.Minute)})

		store.deleteExpiredKeys()

		keys := collectKeys(&store.data)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys, "expected all keys to remain because none expired")
	})

	t.Run("should delete all keys when all are expired", func(t *testing.T) {
		store.data = sync.Map{}
		store.data.Store("key1", memoryItem{value: 1, expiration: time.Now().Add(-1 * time.Hour)})
		store.data.Store("key2", memoryItem{value: 2, expiration: time.Now().Add(-10 * time.Minute)})

		store.deleteExpiredKeys()

		keys := collectKeys(&store.data)
		assert.Empty(t, keys, "expected all expired keys to be deleted")
	})

	t.Run("should delete only expired keys when some are expired", func(t *testing.T) {
		store.data = sync.Map{}
		store.data.Store("key1", memoryItem{value: 1, expiration: time.Now().Add(-1 * time.Hour)})
		store.data.Store("key2", memoryItem{value: 2, expiration: time.Now().Add(10 * time.Minute)})
		store.data.Store("key3", memoryItem{value: 3, expiration: time.Now().Add(-5 * time.Minute)})

		store.deleteExpiredKeys()

		keys := collectKeys(&store.data)
		assert.Equal(t, []string{"key2"}, keys, "expected only non-expired keys to remain")
	})

	t.Run("should keep keys with zero expiration", func(t *testing.T) {
		store.data = sync.Map{}
		store.data.Store("key1", memoryItem{value: 1, expiration: time.Time{}}) // zero means never expire
		store.data.Store("key2", memoryItem{value: 2, expiration: time.Now().Add(-1 * time.Minute)})

		store.deleteExpiredKeys()

		keys := collectKeys(&store.data)
		assert.Equal(t, []string{"key1"}, keys, "expected only zero-expiration keys to remain")
	})
}

func TestMemoryStore_Clear(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	t.Run("should delete all caches", func(t *testing.T) {
		key1 := "key_1"
		key2 := "key_2"
		expected1 := 1
		expected2 := 2

		store.data.Store(key1, memoryItem{
			value:      expected1,
			expiration: time.Now().Add(1 * time.Hour),
		})
		store.data.Store(key2, memoryItem{
			value:      expected2,
			expiration: time.Now().Add(1 * time.Hour),
		})

		item1, _ := store.data.Load(key1)
		actual1, _ := item1.(memoryItem)
		assert.Equal(t, expected1, actual1.value, "expected value before deletion")

		item2, _ := store.data.Load(key2)
		actual2, _ := item2.(memoryItem)
		assert.Equal(t, expected2, actual2.value, "expected value before deletion")

		err := store.Clear()
		assert.NoError(t, err, "expected no error when clear all items")

		item1, _ = store.data.Load(key1)
		actual1, _ = item1.(memoryItem)
		assert.Nil(t, actual1.value, "expected nil value after deletion")

		item2, _ = store.data.Load(key2)
		actual2, _ = item2.(memoryItem)
		assert.Nil(t, actual2.value, "expected nil value after deletion")
	})
}

func TestMemoryStore_Delete(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	t.Run("should delete cache", func(t *testing.T) {
		key := "delete_cache"
		expected := 123
		store.data.Store(key, memoryItem{
			value:      expected,
			expiration: time.Now().Add(1 * time.Hour),
		})

		item, _ := store.data.Load(key)
		actual, _ := item.(memoryItem)
		assert.Equal(t, expected, actual.value, "expected value before deletion")

		err := store.Delete(key)
		assert.NoError(t, err, "expected no error when deleting")

		item, _ = store.data.Load(key)
		actual, _ = item.(memoryItem)
		assert.Nil(t, actual.value, "expected nil value after deletion")
	})
}

func TestMemoryStore_DeleteByPattern(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	// Setup initial data
	store.data.Store("auth:tenant:1:user:123:access_token:456", memoryItem{
		value:      "token1",
		expiration: time.Time{},
	})
	store.data.Store("auth:tenant:2:user:123:access_token:789", memoryItem{
		value:      "token2",
		expiration: time.Time{},
	})
	store.data.Store("user_permissions:tenant:123:user:456", memoryItem{
		value:      "perm1",
		expiration: time.Time{},
	})
	store.data.Store("user_products:user:123:products:1:item:123", memoryItem{
		value:      "prod1",
		expiration: time.Time{},
	})
	store.data.Store("other_key", memoryItem{
		value:      "other",
		expiration: time.Time{},
	})

	t.Run("should delete keys matching single wildcard pattern", func(t *testing.T) {
		err := store.DeleteByPattern("auth:tenant:*:user:123:access_token:*")
		assert.NoError(t, err, "expected no error deleting by pattern")

		// Keys with tenant:1 and tenant:2 should be deleted
		item1, _ := store.data.Load("auth:tenant:1:user:123:access_token:456")
		actual1, _ := item1.(memoryItem)

		item2, _ := store.data.Load("auth:tenant:2:user:123:access_token:789")
		actual2, _ := item2.(memoryItem)

		assert.Nil(t, actual1.value, "expected deleted key to be missing")
		assert.Nil(t, actual2.value, "expected deleted key to be missing")
	})

	t.Run("should not delete keys if no match", func(t *testing.T) {
		err := store.DeleteByPattern("not:matching:*")
		assert.NoError(t, err, "expected no error when no keys match")

		// "user_permissions..." should still exist
		item, _ := store.data.Load("user_permissions:tenant:123:user:456")
		actual, _ := item.(memoryItem)

		assert.Equal(t, "perm1", actual.value)
	})

	t.Run("should delete keys matching multiple wildcards", func(t *testing.T) {
		err := store.DeleteByPattern("user_products:user:123:products:*:item:123")
		assert.NoError(t, err, "expected no error deleting with multiple wildcards")

		item, _ := store.data.Load("user_products:user:123:products:1:item:123")
		actual, _ := item.(memoryItem)

		assert.Nil(t, actual.value, "expected key to be deleted")
	})

	t.Run("should delete keys with exact match pattern", func(t *testing.T) {
		err := store.DeleteByPattern("user_permissions:tenant:123:user:456")
		assert.NoError(t, err, "expected no error deleting with exact match")

		item, _ := store.data.Load("user_permissions:tenant:123:user:456")
		actual, _ := item.(memoryItem)

		assert.Nil(t, actual.value, "expected exact matched key deleted")
	})

	t.Run("should not delete unrelated keys", func(t *testing.T) {
		item, _ := store.data.Load("other_key")
		actual, _ := item.(memoryItem)

		assert.Equal(t, "other", actual.value)
	})
}

func TestMemoryStore_DeleteMany(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	// Setup initial data
	store.data.Store("key1", memoryItem{
		value:      "value1",
		expiration: time.Time{},
	})
	store.data.Store("key2", memoryItem{
		value:      "value2",
		expiration: time.Time{},
	})
	store.data.Store("key3", memoryItem{
		value:      "value3",
		expiration: time.Time{},
	})

	t.Run("should delete existing keys", func(t *testing.T) {
		err := store.DeleteMany("key1", "key2")
		assert.NoError(t, err, "expected no error when deleting existing keys")

		item1, _ := store.data.Load("key1")
		actual1, _ := item1.(memoryItem)
		assert.Nil(t, actual1.value, "expected key1 to be deleted")

		item2, _ := store.data.Load("key2")
		actual2, _ := item2.(memoryItem)
		assert.Nil(t, actual2.value, "expected key2 to be deleted")

		item3, _ := store.data.Load("key3")
		actual3, _ := item3.(memoryItem)
		assert.NotNil(t, actual3.value, "expected key3 to remain")
	})

	t.Run("should handle non-existing key gracefully", func(t *testing.T) {
		err := store.DeleteMany("missing")
		assert.NoError(t, err, "expected no error when deleting non-existing key")

		item3, _ := store.data.Load("key3")
		actual3, _ := item3.(memoryItem)
		assert.NotNil(t, actual3.value, "expected key3 to remain")
	})

	t.Run("should delete mix of existing and missing keys", func(t *testing.T) {
		err := store.DeleteMany("key3", "missing")
		assert.NoError(t, err, "expected no error when deleting mix of keys")

		item3, _ := store.data.Load("key3")
		actual3, _ := item3.(memoryItem)
		assert.Nil(t, actual3.value, "expected key3 to be deleted")
	})

	t.Run("should handle empty input without error", func(t *testing.T) {
		err := store.DeleteMany()
		assert.NoError(t, err, "expected no error when deleting with empty input")
	})
}

func TestMemoryStore_Get(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	t.Run("should return value for non-expired key", func(t *testing.T) {
		key := "shortlived"
		expected := 123
		store.data.Store(key, memoryItem{
			value:      expected,
			expiration: time.Now().Add(10 * time.Second),
		})

		actual, err := store.Get(key)

		assert.NoError(t, err, "expected no error for non-expired key")
		assert.Equal(t, expected, actual, "expected value for key that has not expired")
	})

	t.Run("should return error cache miss for non-existing key", func(t *testing.T) {
		key := "not_exists"

		actual, err := store.Get(key)

		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error for non-existing key")
		assert.Nil(t, actual, "expected nil value for non-existing key")
	})

	t.Run("should return error cache miss for expired key", func(t *testing.T) {
		key := "expired"
		expected := 123
		store.data.Store(key, memoryItem{
			value:      expected,
			expiration: time.Now().Add(-10 * time.Minute),
		})

		actual, err := store.Get(key)

		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error for expired key")
		assert.Nil(t, actual, "expected nil value for expired key")
	})
}

func TestMemoryStore_GetOrSet(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	t.Run("should return existing value when key is present and not expired", func(t *testing.T) {
		key := "exists"
		expected := "cached_value"
		store.data.Store(key, memoryItem{
			value:      expected,
			expiration: time.Now().Add(5 * time.Minute),
		})

		actual, err := store.GetOrSet(key, 1*time.Minute, "new_value")

		assert.NoError(t, err, "expected no error when key exists")
		assert.Equal(t, expected, actual, "expected existing value to be returned, not the new one")
	})

	t.Run("should set and return new value when key is missing", func(t *testing.T) {
		key := "missing"
		expected := "new_value"

		actual, err := store.GetOrSet(key, 1*time.Minute, expected)

		assert.NoError(t, err, "expected no error when setting a new value")
		assert.Equal(t, expected, actual, "expected new value to be returned when key is missing")

		// ensure it was stored
		item, _ := store.data.Load(key)
		stored, _ := item.(memoryItem)
		assert.Equal(t, expected, stored.value, "expected new value to be stored")
	})

	t.Run("should set and return new value when key is expired", func(t *testing.T) {
		key := "expired"
		expected := "new_value"
		store.data.Store(key, memoryItem{
			value:      "old_value",
			expiration: time.Now().Add(-1 * time.Minute), // expired
		})

		actual, err := store.GetOrSet(key, 1*time.Minute, expected)

		assert.NoError(t, err, "expected no error when replacing expired value")
		assert.Equal(t, expected, actual, "expected new value to be returned after expiration")

		// ensure old value was replaced
		item, _ := store.data.Load(key)
		stored, _ := item.(memoryItem)

		assert.Equal(t, expected, stored.value, "expected expired value to be replaced with new one")
	})

	t.Run("should return error if Set fails", func(t *testing.T) {
		key := "invalid"
		invalidTTL := -5 * time.Minute
		expected := "value"

		actual, err := store.GetOrSet(key, invalidTTL, expected)

		assert.ErrorIs(t, err, multicache.ErrInvalidValue, "expected invalid TTL error")
		assert.Nil(t, actual, "expected no value returned on Set error")
	})
}

func TestMemoryStore_Has(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	t.Run("should return false for missing key", func(t *testing.T) {
		exists, err := store.Has("missing")

		assert.NoError(t, err, "expected no error for missing key")
		assert.False(t, exists, "expected missing key to return false")
	})

	t.Run("should return true for existing key", func(t *testing.T) {
		store.data.Store("exists", memoryItem{
			value:      123,
			expiration: time.Now().Add(1 * time.Hour),
		})

		exists, err := store.Has("exists")

		assert.NoError(t, err, "expected no error for existing key")
		assert.True(t, exists, "expected existing key to return true")
	})

	t.Run("should return false for expired key", func(t *testing.T) {
		store.data.Store("expired", memoryItem{
			value:      456,
			expiration: time.Now().Add(-1 * time.Minute),
		})

		exists, err := store.Has("expired")

		assert.NoError(t, err, "expected no error for expired key")
		assert.False(t, exists, "expected expired key to return false")
	})
}

func TestMemoryStore_Set(t *testing.T) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	t.Run("should add a new cache item", func(t *testing.T) {
		key := "add_cache"
		expected := 123

		err := store.Set(key, expected, 1*time.Hour)
		assert.NoError(t, err, "expected no error when storing value")

		item, _ := store.data.Load(key)
		actual, _ := item.(memoryItem)
		assert.NoError(t, err, "expected no error when retrieving stored value")
		assert.Equal(t, expected, actual.value, "expected stored value to be returned")
	})

	t.Run("should overwrite an existing cache item", func(t *testing.T) {
		key := "overwrite"
		initial := 123

		err := store.Set(key, initial, 1*time.Hour)
		assert.NoError(t, err, "expected no error when storing value")

		item, _ := store.data.Load(key)
		actual, _ := item.(memoryItem)
		assert.Equal(t, initial, actual.value, "expected initial value to be returned")

		newValue := 100
		store.Set(key, newValue, 1*time.Hour)

		item, _ = store.data.Load(key)
		actual, _ = item.(memoryItem)
		assert.Equal(t, newValue, actual.value, "expected updated value to overwrite initial value")
	})

	t.Run("should keep value forever when TTL is 0", func(t *testing.T) {
		key := "forever_zero"
		expected := "value-forever"

		err := store.Set(key, expected, 0)
		assert.NoError(t, err, "expected no error when storing value")

		// Wait a bit to simulate time passing
		time.Sleep(20 * time.Millisecond)

		// Key should still exist (never expires)
		item, _ := store.data.Load(key)
		actual, _ := item.(memoryItem)
		assert.Equal(t, expected, actual.value, "expected value to never expire with TTL=0")
	})

	t.Run("should return error with negative TTL", func(t *testing.T) {
		err := store.Set("negatif", 123, -1)

		assert.ErrorIs(t, err, multicache.ErrInvalidValue, "expected invalid value error")
	})
}

func BenchmarkMemoryStore_deleteExpiredKeys(b *testing.B) {
	// Prepare store with mixed expired and non-expired keys
	store := &MemoryStore{
		data: sync.Map{},
	}

	// Fill the store with 10000 keys, half expired, half not
	for i := range 10000 {
		key := fmt.Sprintf("key-%d", i)
		var expiration time.Time
		if i%2 == 0 {
			expiration = time.Now().Add(-1 * time.Minute) // expired
		} else {
			expiration = time.Now().Add(1 * time.Hour) // not expired
		}
		store.data.Store(key, memoryItem{
			value:      i,
			expiration: expiration,
		})
	}

	for b.Loop() {
		store.deleteExpiredKeys()
	}
}

func BenchmarkMemoryStore_Clear(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	for i := range 1000 {
		key := fmt.Sprintf("key-%d", i)
		store.data.Store(key, memoryItem{
			value:      i,
			expiration: time.Now().Add(time.Hour),
		})
	}

	for b.Loop() {
		store.Clear()
	}
}

func BenchmarkMemoryStore_Delete(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	for i := range 10000 {
		key := fmt.Sprintf("key-%d", i)
		store.data.Store(key, memoryItem{
			value:      i,
			expiration: time.Now().Add(time.Hour),
		})
	}

	cases := []struct {
		name string
		key  string
	}{
		{"forget existing key", "key-500"},
		{"forget non-existing key", "key-not-exist"},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				store.Delete(tt.key)
			}
		})
	}
}

func BenchmarkMemoryStore_DeleteByPattern(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	numKeys := 100_000
	for i := range numKeys {
		if i%2 == 0 {
			store.Set(fmt.Sprintf("auth:tenant:%d:user:%d:access_token:%d", i%100, i%1000, i), "value", 0)
		} else {
			store.Set(fmt.Sprintf("user_permissions:tenant:%d:user:%d", i%100, i%1000), "value", 0)
		}
	}

	for b.Loop() {
		store.DeleteByPattern("auth:tenant:*:user:123:access_token:*")
	}
}

func BenchmarkMemoryStore_DeleteMany(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	makeKeys := func(prefix string, n int) []string {
		keys := make([]string, n)
		for i := 0; i < n; i++ {
			key := fmt.Sprintf("%s-%d", prefix, i+1)
			keys[i] = key
			store.data.Store(key, memoryItem{
				value:      "value",
				expiration: time.Time{},
			})
		}
		return keys
	}

	key1 := makeKeys("delete_1", 1)
	key10 := makeKeys("delete_10", 10)
	key100 := makeKeys("delete_100", 100)

	cases := []struct {
		name string
		keys []string
	}{
		{"delete 1", key1},
		{"delete 10", key10},
		{"delete 100", key100},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				store.DeleteMany(tt.keys...)
			}
		})
	}
}

func BenchmarkMemoryStore_Get(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	store.data.Store("shortlived", memoryItem{
		value:      123,
		expiration: time.Now().Add(10 * time.Second),
	})
	store.data.Store("expired", memoryItem{
		value:      123,
		expiration: time.Now().Add(-10 * time.Minute),
	})
	store.data.Store("forever", memoryItem{
		value:      123,
		expiration: time.Time{},
	})

	cases := []struct {
		name string
		key  string
	}{
		{"non-existing key", "not_exists"},
		{"non-expired key", "shortlived"},
		{"cache forever key", "forever"},
		{"expired key", "expired"},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				store.Get(tt.key)
			}
		})
	}
}

func BenchmarkMemoryStore_GetOrSet(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	cases := []struct {
		name  string
		key   string
		value any
		ttl   time.Duration
	}{
		{"cache miss then set", "miss_then_set", 123, time.Hour},
		{"cache hit existing key", "hit_existing", 123, time.Hour},
		{"cache forever with TTL 0", "forever", 123, 0},
	}

	// Pre-populate "hit_existing"
	_, _ = store.GetOrSet("hit_existing", time.Hour, 123)

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				store.GetOrSet(tt.key, tt.ttl, tt.value)
			}
		})
	}
}

func BenchmarkMemoryStore_Has(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	store.data.Store("valid", memoryItem{
		value:      123,
		expiration: time.Now().Add(1 * time.Hour), // valid
	})
	store.data.Store("expired", memoryItem{
		value:      456,
		expiration: time.Now().Add(-1 * time.Hour), // expired
	})

	cases := []struct {
		name string
		key  string
		want bool
	}{
		{"key exists", "valid", true},
		{"key missing", "missing", false},
		{"key expired", "expired", false},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				store.Has(tt.key)
			}
		})
	}
}

func BenchmarkMemoryStore_Set(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	cases := []struct {
		name  string
		key   string
		value any
		ttl   time.Duration
	}{
		{"add new cache", "add_cache", 123, 1 * time.Hour},
		{"overwrite existing cache", "overwrite", 123, 1 * time.Hour},
		{"cache forever with TTL 0", "forever", 123, 0},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				store.Set(tt.key, tt.value, tt.ttl)
			}
		})
	}
}
