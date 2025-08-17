package memory

import (
	"context"
	"fmt"
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
		assert.NotNil(t, store.data, "expected store data map to be initialized")
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
		store.data["expired"] = memoryItem{
			value:      1,
			expiration: time.Now().Add(-1 * time.Second),
		}

		time.Sleep(20 * time.Millisecond) // wait for cleanup
		_, exists := store.data["expired"]
		assert.False(t, exists, "expected expired key to be deleted by cleanupExpiredKeys goroutine")

		// Cancel cleanup goroutine
		store.cancelCleanup()
	})

	t.Run("should allow multiple independent stores", func(t *testing.T) {
		store1, _ := NewMemoryStore(map[string]any{})
		store2, _ := NewMemoryStore(map[string]any{})

		s1 := store1.(*MemoryStore)
		s2 := store2.(*MemoryStore)

		s1.Put("key", 1)
		s2.Put("key", 2)

		v1, _ := s1.Get("key")
		v2, _ := s2.Get("key")

		assert.Equal(t, 1, v1, "expected store1 to return its own value independently")
		assert.Equal(t, 2, v2, "expected store2 to return its own value independently")

		s1.cancelCleanup()
		s2.cancelCleanup()
	})
}

func TestMemoryStore_cleanupExpiredKeys(t *testing.T) {
	baseTime := time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC)
	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	moveForward := func(d time.Duration) {
		baseTime = baseTime.Add(d)
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
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(-1 * time.Minute)}, // expired
			"key2": {value: 2, expiration: baseTime.Add(1 * time.Hour)},    // not expired
		}

		go store.cleanupExpiredKeys(ctx, 50*time.Millisecond)

		time.Sleep(100 * time.Millisecond)
		assert.NotContains(t, store.data, "key1", "expected expired key1 to be deleted during cleanup")
		assert.Contains(t, store.data, "key2", "expected non-expired key2 to remain during cleanup")

		cancel()
	})

	t.Run("should handle store with no expired keys", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(1 * time.Minute)},
			"key2": {value: 2, expiration: baseTime.Add(2 * time.Minute)},
		}

		go store.cleanupExpiredKeys(ctx, 50*time.Millisecond)
		time.Sleep(100 * time.Millisecond)

		assert.Len(t, store.data, 2, "expected all keys to remain because none are expired")
		cancel()
	})

	t.Run("should delete keys that expire after time moves forward", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(1 * time.Minute)},
			"key2": {value: 2, expiration: baseTime.Add(2 * time.Minute)},
		}

		go store.cleanupExpiredKeys(ctx, 50*time.Millisecond)

		moveForward(90 * time.Second) // 1.5 minutes
		time.Sleep(100 * time.Millisecond)

		assert.NotContains(t, store.data, "key1", "expected key1 to be deleted after expiration")
		assert.Contains(t, store.data, "key2", "expected key2 to remain until it expires")

		cancel()
	})
}

func TestMemoryStore_deleteExpiredKeys(t *testing.T) {
	baseTime := time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC)
	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	moveForward := func(d time.Duration) {
		baseTime = baseTime.Add(d)
	}

	t.Run("should do nothing when store is empty", func(t *testing.T) {
		store.data = map[string]memoryItem{}

		store.deleteExpiredKeys()
		assert.Empty(t, store.data, "expected store to remain empty when no keys exist")
	})

	t.Run("should keep all keys when none are expired", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(1 * time.Hour)},
			"key2": {value: 2, expiration: baseTime.Add(10 * time.Minute)},
		}

		store.deleteExpiredKeys()
		assert.Len(t, store.data, 2, "expected all keys to remain because none expired")
	})

	t.Run("should delete all keys when all are expired", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(-1 * time.Hour)},
			"key2": {value: 2, expiration: baseTime.Add(-10 * time.Minute)},
		}

		store.deleteExpiredKeys()
		assert.Empty(t, store.data, "expected all expired keys to be deleted")
	})

	t.Run("should delete only expired keys when some are expired", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(-1 * time.Hour)},
			"key2": {value: 2, expiration: baseTime.Add(10 * time.Minute)},
			"key3": {value: 3, expiration: baseTime.Add(-5 * time.Minute)},
		}

		store.deleteExpiredKeys()
		assert.Len(t, store.data, 1, "expected only non-expired keys to remain")
		assert.Contains(t, store.data, "key2", "expected key2 to remain because it is not expired")
	})

	t.Run("should keep keys with zero expiration", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: time.Time{}}, // zero means never expire
			"key2": {value: 2, expiration: baseTime.Add(-1 * time.Minute)},
		}

		store.deleteExpiredKeys()
		assert.Len(t, store.data, 1, "expected only zero-expiration keys to remain")
		assert.Contains(t, store.data, "key1", "expected key1 with zero expiration to remain")
	})

	t.Run("should delete keys that expire after time moves forward", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(1 * time.Minute)},
			"key2": {value: 2, expiration: baseTime.Add(2 * time.Minute)},
		}

		moveForward(90 * time.Second) // 1.5 minutes

		store.deleteExpiredKeys()
		assert.Len(t, store.data, 1, "expected only non-expired keys to remain after time move")
		assert.Contains(t, store.data, "key2", "expected key2 to remain because it is not expired")
		assert.NotContains(t, store.data, "key1", "expected key1 to be deleted because it expired")
	})
}

func TestMemoryStore_findExpiredKeys(t *testing.T) {
	baseTime := time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC)
	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	moveForward := func(d time.Duration) {
		baseTime = baseTime.Add(d)
	}

	t.Run("should return no keys when store is empty", func(t *testing.T) {
		keys := store.findExpiredKeys()
		assert.Empty(t, keys, "expected no expired keys when store is empty")
	})

	t.Run("should return no keys when all keys are non-expired", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(1 * time.Hour)},
			"key2": {value: 2, expiration: baseTime.Add(10 * time.Minute)},
		}

		keys := store.findExpiredKeys()
		assert.Empty(t, keys, "expected no expired keys when all keys are valid")
	})

	t.Run("should return all keys when all keys are expired", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(-1 * time.Hour)},
			"key2": {value: 2, expiration: baseTime.Add(-10 * time.Minute)},
		}

		keys := store.findExpiredKeys()
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys, "expected all keys to be expired")
	})

	t.Run("should return only expired keys when some keys are expired", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(-1 * time.Hour)},
			"key2": {value: 2, expiration: baseTime.Add(10 * time.Minute)},
			"key3": {value: 3, expiration: baseTime.Add(-5 * time.Minute)},
		}

		keys := store.findExpiredKeys()
		assert.ElementsMatch(t, []string{"key1", "key3"}, keys, "expected only expired keys to be returned")
	})

	t.Run("should ignore keys with zero expiration", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: time.Time{}}, // zero means never expire
			"key2": {value: 2, expiration: baseTime.Add(-1 * time.Minute)},
		}

		keys := store.findExpiredKeys()
		assert.ElementsMatch(t, []string{"key2"}, keys, "expected zero-expiration keys to never be considered expired")
	})

	t.Run("should return keys that expire after time moves forward", func(t *testing.T) {
		store.data = map[string]memoryItem{
			"key1": {value: 1, expiration: baseTime.Add(1 * time.Minute)},
			"key2": {value: 2, expiration: baseTime.Add(2 * time.Minute)},
		}

		moveForward(90 * time.Second) // 1.5 minutes

		keys := store.findExpiredKeys()
		assert.ElementsMatch(t, []string{"key1"}, keys, "expected only key1 to be expired after time moved forward")
	})
}

func TestMemoryStore_now(t *testing.T) {
	t.Run("should return custom time when nowFunc is set", func(t *testing.T) {
		expected := time.Date(2025, 8, 16, 12, 0, 0, 0, time.UTC)
		store := &MemoryStore{
			nowFunc: func() time.Time {
				return expected
			},
		}

		actual := store.now()
		assert.Equal(t, expected, actual, "expected now() to return the value from nowFunc")
	})

	t.Run("should return current time when nowFunc is nil", func(t *testing.T) {
		store := &MemoryStore{
			nowFunc: nil,
		}

		before := time.Now()
		actual := store.now()
		after := time.Now()

		// Check actual is between before and after
		assert.True(t, actual.Equal(before) || actual.After(before), "expected now() to be after or equal to before time")
		assert.True(t, actual.Equal(after) || actual.Before(after), "expected now() to be before or equal to after time")
	})
}

func TestMemoryStore_Add(t *testing.T) {
	baseTime := time.Date(2025, 8, 17, 0, 0, 0, 0, time.UTC)
	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	t.Run("should add new key successfully", func(t *testing.T) {
		err := store.Add("new", 999)
		assert.NoError(t, err, "expected no error when adding new key")

		item, exists := store.data["new"]
		assert.True(t, exists, "expected key to exist after Add")
		assert.Equal(t, 999, item.value, "expected stored value for key 'new' to be 999")
	})

	t.Run("should return error when adding duplicate key", func(t *testing.T) {
		store.data["exists"] = memoryItem{
			value:      123,
			expiration: baseTime.Add(1 * time.Hour),
		}

		err := store.Add("exists", 456)
		assert.ErrorIs(t, err, multicache.ErrItemAlreadyExists)
	})

	t.Run("should add expired key as new", func(t *testing.T) {
		store.data["expired"] = memoryItem{
			value:      888,
			expiration: baseTime.Add(-1 * time.Minute),
		}
		err := store.Add("expired", 777)
		assert.NoError(t, err, "expected no error when adding expired key")

		item, ok := store.data["expired"]
		assert.True(t, ok, "expected key to exist after Add")
		assert.Equal(t, 777, item.value, "expected stored value for key 'expired' to be 777")
	})
}

func TestMemoryStore_Flush(t *testing.T) {
	store := &MemoryStore{
		data: make(map[string]memoryItem),
	}

	t.Run("should delete all caches", func(t *testing.T) {
		key1 := "key_1"
		key2 := "key_2"
		expected1 := 1
		expected2 := 2

		store.Put(key1, expected1)
		store.Put(key2, expected2)

		actual1, err := store.Get(key1)
		assert.NoError(t, err, "expected no error before deletion")
		assert.Equal(t, expected1, actual1, "expected value before deletion")

		actual2, err := store.Get(key2)
		assert.NoError(t, err, "expected no error before deletion")
		assert.Equal(t, expected2, actual2, "expected value before deletion")

		store.Flush()

		actual1, err = store.Get(key1)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error after deletion")
		assert.Nil(t, actual1, "expected nil value after deletion")

		actual2, err = store.Get(key2)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error after deletion")
		assert.Nil(t, actual2, "expected nil value after deletion")
	})
}

func TestMemoryStore_Forget(t *testing.T) {
	baseTime := time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC)

	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	t.Run("should delete cache", func(t *testing.T) {
		key := "delete_cache"
		expected := 123

		store.Put(key, expected)

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error before deletion")
		assert.Equal(t, expected, actual, "expected value before deletion")

		store.Forget(key)

		actual, err = store.Get(key)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error after deletion")
		assert.Nil(t, actual, "expected nil value after deletion")
	})
}

func TestMemoryStore_Get(t *testing.T) {
	baseTime := time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC)

	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	moveForward := func(d time.Duration) {
		baseTime = baseTime.Add(d)
	}

	t.Run("should return error cache miss for non-existing key", func(t *testing.T) {
		key := "not_exists"

		actual, err := store.Get(key)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error for non-existing key")
		assert.Nil(t, actual, "expected nil value for non-existing key")
	})

	t.Run("should return fallback value for non-existing key", func(t *testing.T) {
		key := "non_existing_key"
		fallback := "fallback_value"

		actual, err := store.Get(key, fallback)
		assert.NoError(t, err, "expected no error when fallback value is provided")
		assert.Equal(t, fallback, actual, "expected fallback value when provided")
	})

	t.Run("should return value from fallback function for non-existing key", func(t *testing.T) {
		key := "non_existing_key_func"

		fallbackFunc := func() any {
			return "value_from_func"
		}

		actual, err := store.Get(key, fallbackFunc())
		assert.NoError(t, err, "expected no error when fallback function is provided")
		assert.Equal(t, "value_from_func", actual, "expected value returned from fallback function")
	})

	t.Run("should return fallback value for expired key", func(t *testing.T) {
		key := "expired_key"
		storedValue := 123
		fallback := 100

		store.Put(key, storedValue, 10*time.Minute)
		moveForward(11 * time.Minute) // key is now expired

		actual, err := store.Get(key, fallback)
		assert.NoError(t, err, "expected no error when fallback value is provided")
		assert.NotEqual(t, storedValue, actual, "expected expired value not to be returned")
		assert.Equal(t, fallback, actual, "expected fallback value when provided")
	})

	t.Run("should return value from fallback function for expired key", func(t *testing.T) {
		key := "expired_key_func"
		storedValue := 123
		fallbackFunc := func() any {
			return "value_from_func"
		}

		store.Put(key, storedValue, 10*time.Minute)
		moveForward(11 * time.Minute) // key is now expired

		actual, err := store.Get(key, fallbackFunc())
		assert.NoError(t, err, "expected no error when fallback function is provided for expired key")
		assert.NotEqual(t, storedValue, actual, "expected expired value not to be returned")
		assert.Equal(t, "value_from_func", actual, "expected value returned from fallback function for expired key")
	})

	t.Run("should return value for cache forever key even after long time", func(t *testing.T) {
		key := "forever"
		expected := "value_forever"

		store.Put(key, expected, 0)
		moveForward(10 * 365 * 24 * time.Hour) // 10 years

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error even after long time")
		assert.Equal(t, expected, actual, "expected value for key with TTL 0 even after long time")
	})

	t.Run("should return value for non-expired key", func(t *testing.T) {
		key := "shortlived"
		expected := 123

		store.Put(key, expected, 10*time.Second)

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error for non-expired key")
		assert.Equal(t, expected, actual, "expected value for key that has not expired")
	})

	t.Run("should return error cache miss for expired key", func(t *testing.T) {
		key := "expired"
		expected := 123

		store.Put(key, expected, 10*time.Minute)
		moveForward(11 * time.Minute)

		actual, err := store.Get(key)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error for expired key")
		assert.Nil(t, actual, "expected nil value for expired key")
	})
}

func TestMemoryStore_Has(t *testing.T) {
	baseTime := time.Date(2025, 8, 17, 0, 0, 0, 0, time.UTC)
	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	t.Run("should return false for missing key", func(t *testing.T) {
		exists, err := store.Has("missing")
		assert.NoError(t, err, "expected no error for missing key")
		assert.False(t, exists, "expected missing key to return false")
	})

	t.Run("should return true for existing key", func(t *testing.T) {
		store.data["exists"] = memoryItem{
			value:      123,
			expiration: baseTime.Add(1 * time.Hour),
		}

		exists, err := store.Has("exists")
		assert.NoError(t, err, "expected no error for existing key")
		assert.True(t, exists, "expected existing key to return true")
	})

	t.Run("should return false for expired key", func(t *testing.T) {
		store.data["expired"] = memoryItem{
			value:      456,
			expiration: baseTime.Add(-1 * time.Minute),
		}

		exists, err := store.Has("expired")
		assert.NoError(t, err, "expected no error for expired key")
		assert.False(t, exists, "expected expired key to return false")
	})
}

func TestMemoryStore_Put(t *testing.T) {
	baseTime := time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC)

	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	moveForward := func(d time.Duration) {
		baseTime = baseTime.Add(d)
	}

	t.Run("should add a new cache item", func(t *testing.T) {
		key := "add_cache"
		expected := 123

		store.Put(key, expected, 1*time.Hour)

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error when retrieving stored value")
		assert.Equal(t, expected, actual, "expected stored value to be returned")
	})

	t.Run("should overwrite an existing cache item", func(t *testing.T) {
		key := "overwrite"
		initial := 123

		store.Put(key, initial, 1*time.Hour)

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error for initial value")
		assert.Equal(t, initial, actual, "expected initial value to be returned")

		newValue := 100
		store.Put(key, newValue, 1*time.Hour)

		actual, err = store.Get(key)
		assert.NoError(t, err, "expected no error for updated value")
		assert.Equal(t, newValue, actual, "expected updated value to overwrite initial value")
	})

	t.Run("should expire cache after duration", func(t *testing.T) {
		key := "with_duration"
		expected := 123

		store.Put(key, expected, 10*time.Second)

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error before expiration")
		assert.Equal(t, expected, actual, "expected value before expiration")

		moveForward(11 * time.Second)

		actual, err = store.Get(key)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error after expiration")
		assert.Nil(t, actual, "expected nil value after expiration")
	})

	t.Run("should keep value forever when TTL is 0", func(t *testing.T) {
		key := "forever_zero"
		expected := "value-forever"

		store.Put(key, expected, 0)
		moveForward(10 * 365 * 24 * time.Hour) // 10 years

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error even after long duration")
		assert.Equal(t, expected, actual, "expected value to never expire with TTL=0")
	})

	t.Run("should keep value forever when TTL is not provided", func(t *testing.T) {
		key := "forever_none"
		expected := "value-forever"

		store.Put(key, expected)
		moveForward(10 * 365 * 24 * time.Hour) // 10 years

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error even after long duration")
		assert.Equal(t, expected, actual, "expected value to never expire without TTL")
	})

	t.Run("should delete cache item with negative TTL", func(t *testing.T) {
		key := "negatif"
		expected := 123

		store.Put(key, expected)

		actual, err := store.Get(key)
		assert.NoError(t, err, "expected no error before deletion")
		assert.Equal(t, expected, actual, "expected value before deletion")

		store.Put(key, expected, -1)

		actual, err = store.Get(key)
		assert.ErrorIs(t, err, multicache.ErrCacheMiss, "expected cache miss error after negative TTL")
		assert.Nil(t, actual, "expected nil value after negative TTL")
	})
}

func BenchmarkMemoryStore_deleteExpiredKeys(b *testing.B) {
	baseTime := time.Now()

	// Prepare store with mixed expired and non-expired keys
	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	// Fill the store with 10000 keys, half expired, half not
	for i := range 10000 {
		key := fmt.Sprintf("key-%d", i)
		var expiration time.Time
		if i%2 == 0 {
			expiration = baseTime.Add(-1 * time.Minute) // expired
		} else {
			expiration = baseTime.Add(1 * time.Hour) // not expired
		}
		store.data[key] = memoryItem{
			value:      i,
			expiration: expiration,
		}
	}

	for b.Loop() {
		store.deleteExpiredKeys()
	}
}

func BenchmarkMemoryStore_findExpiredKeys(b *testing.B) {
	baseTime := time.Now()
	store := &MemoryStore{
		data: make(map[string]memoryItem),
		nowFunc: func() time.Time {
			return baseTime
		},
	}

	// Fill the store with 10000 keys, half expired, half not
	for i := range 10000 {
		key := fmt.Sprintf("key-%d", i)
		var expiration time.Time
		if i%2 == 0 {
			expiration = baseTime.Add(-1 * time.Minute) // expired
		} else {
			expiration = baseTime.Add(1 * time.Hour) // not expired
		}
		store.data[key] = memoryItem{
			value:      i,
			expiration: expiration,
		}
	}

	for b.Loop() {
		_ = store.findExpiredKeys()
	}
}

func BenchmarkMemoryStore_Add(b *testing.B) {
	baseTime := time.Date(2025, 8, 17, 0, 0, 0, 0, time.UTC)

	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	// Prepopulate keys for testing duplicates and expired
	store.data["exists"] = memoryItem{
		value:      123,
		expiration: baseTime.Add(1 * time.Hour),
	}
	store.data["expired"] = memoryItem{
		value:      456,
		expiration: baseTime.Add(-1 * time.Minute),
	}

	cases := []struct {
		name  string
		key   string
		value any
	}{
		{"add new key", "new", 999},
		{"add existing key", "exists", 111},
		{"add expired key", "expired", 777},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				_ = store.Add(tt.key, tt.value)
			}
		})
	}
}

func BenchmarkMemoryStore_Flush(b *testing.B) {
	store := &MemoryStore{
		data: make(map[string]memoryItem),
	}

	for i := range 1000 {
		key := fmt.Sprintf("key-%d", i)
		store.Put(key, i, time.Hour)
	}

	for b.Loop() {
		store.Flush()
	}
}

func BenchmarkMemoryStore_Forget(b *testing.B) {
	store := &MemoryStore{
		data: make(map[string]memoryItem),
	}

	for i := range 10000 {
		key := fmt.Sprintf("key-%d", i)
		store.Put(key, i, time.Hour)
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
				store.Forget(tt.key)
			}
		})
	}
}

func BenchmarkMemoryStore_Get(b *testing.B) {
	baseTime := time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC)
	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	moveForward := func(d time.Duration) {
		baseTime = baseTime.Add(d)
	}

	store.Put("expired_key", 123, 10*time.Minute)
	store.Put("forever", "value_forever", 0)
	store.Put("shortlived", 123, 10*time.Second)
	store.Put("expired", 123, 10*time.Minute)

	cases := []struct {
		name     string
		key      string
		fb       any
		duration time.Duration
	}{
		{"non-existing key", "not_exists", nil, 0},
		{"non-existing key with fallback", "non_existing_key", "fallback_value", 0},
		{"expired key with fallback", "expired_key", 100, 11 * time.Minute},
		{"cache forever key", "forever", nil, 10 * 365 * 24 * time.Hour},
		{"non-expired key", "shortlived", nil, 0},
		{"expired key", "expired", nil, 11 * time.Minute},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				moveForward(tt.duration)
				store.Get(tt.key, tt.fb)
			}
		})
	}
}

func BenchmarkMemoryStore_Has(b *testing.B) {
	baseTime := time.Date(2025, 8, 17, 0, 0, 0, 0, time.UTC)

	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	store.data["valid"] = memoryItem{
		value:      123,
		expiration: baseTime.Add(1 * time.Hour), // valid
	}
	store.data["expired"] = memoryItem{
		value:      456,
		expiration: baseTime.Add(-1 * time.Hour), // expired
	}

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

func BenchmarkMemoryStore_Put(b *testing.B) {
	baseTime := time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC)
	store := &MemoryStore{
		data:    make(map[string]memoryItem),
		nowFunc: func() time.Time { return baseTime },
	}

	cases := []struct {
		name  string
		key   string
		value any
		ttl   []time.Duration
	}{
		{"add new cache", "add_cache", 123, []time.Duration{1 * time.Hour}},
		{"overwrite existing cache", "overwrite", 100, []time.Duration{1 * time.Hour}},
		{"expire after duration", "with_duration", 123, []time.Duration{10 * time.Second}},
		{"cache forever with TTL 0", "forever_zero", "value-forever", []time.Duration{0}},
		{"cache forever without TTL", "forever_none", "value-forever", nil},
		{"delete with negative TTL", "negatif", 123, []time.Duration{-1}},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				store.Put(tt.key, tt.value, tt.ttl...)
			}
		})
	}
}
