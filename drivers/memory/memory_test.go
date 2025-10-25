package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shoraid/omnicache"
	"github.com/shoraid/omnicache/internal/assert"
)

func TestMemoryStore_NewMemoryStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         MemoryConfig
		expectInterval time.Duration
	}{
		{
			name:           "should use default cleanup interval when not provided",
			config:         MemoryConfig{},
			expectInterval: DefaultCleanupInterval,
		},
		{
			name: "should use custom cleanup interval when provided",
			config: MemoryConfig{
				CleanupInterval: 100 * time.Millisecond,
			},
			expectInterval: 100 * time.Millisecond,
		},
		{
			name: "should fallback to default when cleanup interval is zero",
			config: MemoryConfig{
				CleanupInterval: 0,
			},
			expectInterval: DefaultCleanupInterval,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Act ---
			store, err := NewMemoryStore(tt.config)

			// --- Assert ---
			assert.NoError(t, err, "expected no error when creating store")
			assert.NotNil(t, store, "expected error when creating store")

			memStore, ok := store.(*MemoryStore)
			assert.True(t, ok, "expected store to be of type *MemoryStore")
			assert.NotNil(t, memStore.cancelCleanup, "expected cancelCleanup function to be initialized")

			// cancel goroutine safely to avoid leaks
			memStore.cancelCleanup()
		})
	}
}

func TestMemoryStore_Clear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(*MemoryStore)
	}{
		{
			name: "should clear all keys when store has multiple entries",
			setup: func(m *MemoryStore) {
				m.data.Store("user:1", memoryItem{value: "John"})
				m.data.Store("user:2", memoryItem{value: "Jane"})
			},
		},
		{
			name: "should clear store with mixed expired and non-expired keys",
			setup: func(m *MemoryStore) {
				m.data.Store("expired_key", memoryItem{value: "val1", expiration: time.Now().Add(-1 * time.Hour)})
				m.data.Store("valid_key", memoryItem{value: "val2", expiration: time.Now().Add(1 * time.Hour)})
			},
		},
		{
			name: "should handle clear when store is already empty",
			setup: func(m *MemoryStore) {
				// nothing to setup
			},
		},
		{
			name: "should not panic when called multiple times consecutively",
			setup: func(m *MemoryStore) {
				m.data.Store("key", memoryItem{value: "data"})
				m.data.Clear() // first clear
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			store := &MemoryStore{}
			tt.setup(store)

			// --- Act ---
			err := store.Clear(ctx)

			// --- Assert ---
			assert.NoError(t, err, "expected no error when clearing store")

			// Verify that store is empty
			count := 0
			store.data.Range(func(_, _ any) bool {
				count++
				return true
			})

			assert.Equal(t, 0, count, "expected store to be empty after Clear()")
		})
	}
}

func TestMemoryStore_Close(t *testing.T) {
	t.Parallel()

	t.Run("should stop cleanup goroutine when Close is called", func(t *testing.T) {
		t.Parallel()

		// --- Arrange ---
		cleanupInterval := 50 * time.Millisecond
		store, err := NewMemoryStore(MemoryConfig{CleanupInterval: cleanupInterval})
		assert.NoError(t, err)

		memoryStore, ok := store.(*MemoryStore)
		assert.True(t, ok, "expected store to be *MemoryStore")

		// --- Act ---
		err = memoryStore.Close(context.Background())

		// --- Assert ---
		assert.NoError(t, err, "expected no error when calling Close")

		// Wait for doneCh to close (goroutine exit signal)
		select {
		case <-memoryStore.doneCh:
			// success — cleanup goroutine stopped
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected cleanup goroutine to stop quickly")
		}

		// --- Act Again (ensure idempotency) ---
		err = memoryStore.Close(context.Background())
		assert.NoError(t, err, "expected no error when calling Close multiple times")
	})

	t.Run("should safely handle nil cancelCleanup", func(t *testing.T) {
		t.Parallel()

		// --- Arrange ---
		store := &MemoryStore{}

		// --- Act ---
		err := store.Close(context.Background())

		// --- Assert ---
		assert.NoError(t, err, "expected no error when Close called with nil cancelCleanup")
	})
}

func TestMemoryStore_cleanupExpiredKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		interval      time.Duration
		setup         func(*MemoryStore)
		expectedCount int // Expected number of keys after cleanup
	}{
		{
			name:     "should remove expired keys when cleanup interval elapses",
			interval: 10 * time.Millisecond,
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "val1", expiration: time.Now().Add(-5 * time.Millisecond)}) // Expired
				m.data.Store("key2", memoryItem{value: "val2", expiration: time.Now().Add(1 * time.Hour)})         // Not expired
			},
			expectedCount: 1,
		},
		{
			name:     "should keep non-expired and no-expiration keys when cleanup runs",
			interval: 10 * time.Millisecond,
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "val1", expiration: time.Now().Add(1 * time.Hour)}) // Not expired
				m.data.Store("key2", memoryItem{value: "val2", expiration: time.Time{}})                   // No expiration
			},
			expectedCount: 2,
		},
		{
			name:     "should not fail when store is empty",
			interval: 10 * time.Millisecond,
			setup: func(m *MemoryStore) {
				// empty
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			store := &MemoryStore{}
			tt.setup(store)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel() // Ensure the goroutine is stopped

			// --- Act ---
			go store.cleanupExpiredKeys(ctx, tt.interval)
			time.Sleep(tt.interval + (5 * time.Millisecond)) // Let cleanup run
			cancel()                                         // Stop goroutine
			time.Sleep(10 * time.Millisecond)                // Ensure exit

			// --- Assert ---
			count := 0
			store.data.Range(func(_, _ any) bool {
				count++
				return true
			})
			assert.Equal(t, tt.expectedCount, count, "expected number of keys after cleanup mismatch")
		})
	}
}

func TestMemoryStore_deleteExpiredKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		setup                 func(*MemoryStore)
		expectedRemainingKeys []string
	}{
		{
			name: "should delete expired keys when some have passed expiration",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "val1", expiration: time.Now().Add(-1 * time.Second)})
				m.data.Store("key2", memoryItem{value: "val2", expiration: time.Now().Add(-5 * time.Minute)})
				m.data.Store("key3", memoryItem{value: "val3", expiration: time.Now().Add(1 * time.Hour)}) // Not expired
				m.data.Store("key4", memoryItem{value: "val4", expiration: time.Time{}})                   // No expiration
			},
			expectedRemainingKeys: []string{"key3", "key4"},
		},
		{
			name: "should keep all keys when none are expired",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "val1", expiration: time.Now().Add(1 * time.Hour)})
				m.data.Store("key2", memoryItem{value: "val2", expiration: time.Time{}})
			},
			expectedRemainingKeys: []string{"key1", "key2"},
		},
		{
			name: "should do nothing when store is empty",
			setup: func(m *MemoryStore) {
				// empty store
			},
			expectedRemainingKeys: []string{},
		},
		{
			name: "should delete only expired keys when both valid and expired exist",
			setup: func(m *MemoryStore) {
				m.data.Store("expired_a", memoryItem{value: "a", expiration: time.Now().Add(-10 * time.Second)})
				m.data.Store("valid_b", memoryItem{value: "b", expiration: time.Now().Add(1 * time.Minute)})
				m.data.Store("expired_c", memoryItem{value: "c", expiration: time.Now().Add(-30 * time.Second)})
				m.data.Store("valid_d", memoryItem{value: "d", expiration: time.Time{}})
			},
			expectedRemainingKeys: []string{"valid_b", "valid_d"},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			store := &MemoryStore{}
			tt.setup(store)

			// --- Act ---
			store.deleteExpiredKeys()

			// --- Assert ---
			remainingKeys := make(map[string]struct{})
			store.data.Range(func(k, _ any) bool {
				keyStr, ok := k.(string)
				assert.True(t, ok, "expected key to be a string")
				remainingKeys[keyStr] = struct{}{}
				return true
			})

			assert.Equal(t, len(tt.expectedRemainingKeys), len(remainingKeys), "number of remaining keys mismatch")

			for _, expectedKey := range tt.expectedRemainingKeys {
				_, exists := remainingKeys[expectedKey]
				assert.True(t, exists, "expected key to remain in cache")
			}
		})
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		setup                 func(*MemoryStore)
		keyToDelete           string
		expectedRemainingKeys []string // keys expected to remain after deletion
	}{
		{
			name: "should delete key when it exists",
			setup: func(m *MemoryStore) {
				m.data.Store("user:1", memoryItem{value: "John"})
				m.data.Store("user:2", memoryItem{value: "Jane"})
			},
			keyToDelete:           "user:1",
			expectedRemainingKeys: []string{"user:2"},
		},
		{
			name: "should do nothing when key does not exist",
			setup: func(m *MemoryStore) {
				m.data.Store("user:1", memoryItem{value: "John"})
			},
			keyToDelete:           "nonexistent_key",
			expectedRemainingKeys: []string{"user:1"},
		},
		{
			name: "should do nothing when store is empty",
			setup: func(m *MemoryStore) {
				// empty store
			},
			keyToDelete:           "any_key",
			expectedRemainingKeys: []string{},
		},
		{
			name: "should delete key when it is the only entry",
			setup: func(m *MemoryStore) {
				m.data.Store("single_key", memoryItem{value: "value"})
			},
			keyToDelete:           "single_key",
			expectedRemainingKeys: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			store := &MemoryStore{}
			tt.setup(store)

			// --- Act ---
			err := store.Delete(ctx, tt.keyToDelete)

			// --- Assert ---
			assert.NoError(t, err, "expected no error when deleting key")

			remainingKeys := make(map[string]struct{})
			store.data.Range(func(k, _ any) bool {
				keyStr, ok := k.(string)
				assert.True(t, ok, "expected key to be a string")
				remainingKeys[keyStr] = struct{}{}
				return true
			})

			assert.Equal(t, len(tt.expectedRemainingKeys), len(remainingKeys), "number of remaining keys mismatch")

			for _, expectedKey := range tt.expectedRemainingKeys {
				_, exists := remainingKeys[expectedKey]
				assert.True(t, exists, "expected key to remain in cache")
			}
		})
	}
}

func TestMemoryStore_DeleteByPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		setup                 func(*MemoryStore)
		pattern               string
		expectedRemainingKeys []string // keys expected to remain after deletion
		expectedError         bool
		cancelCtx             bool
	}{
		{
			name: "should delete matching keys when using literal pattern without wildcard",
			setup: func(m *MemoryStore) {
				m.data.Store("exact_match", memoryItem{value: "val1"})
				m.data.Store("exact_match_suffix", memoryItem{value: "val2"})
				m.data.Store("prefix_exact_match", memoryItem{value: "val3"})
			},
			pattern:               "exact_match",
			expectedRemainingKeys: []string{"exact_match_suffix", "prefix_exact_match"},
			expectedError:         false,
		},
		{
			name: "should delete all keys when using global '*' pattern",
			setup: func(m *MemoryStore) {
				m.data.Store("user:1", memoryItem{value: "John"})
				m.data.Store("product:1", memoryItem{value: "Laptop"})
			},
			pattern:               "*",
			expectedRemainingKeys: []string{},
			expectedError:         false,
		},
		{
			name: "should delete matching keys when using simple prefix pattern",
			setup: func(m *MemoryStore) {
				m.data.Store("user:1", memoryItem{value: "John"})
				m.data.Store("user:2", memoryItem{value: "Jane"})
				m.data.Store("product:1", memoryItem{value: "Laptop"})
			},
			pattern:               "user:*",
			expectedRemainingKeys: []string{"product:1"},
			expectedError:         false,
		},
		{
			name: "should delete matching keys when using suffix pattern",
			setup: func(m *MemoryStore) {
				m.data.Store("item:apple", memoryItem{value: "Apple"})
				m.data.Store("fruit:apple", memoryItem{value: "Red Apple"})
				m.data.Store("item:orange", memoryItem{value: "Orange"})
			},
			pattern:               "*:apple",
			expectedRemainingKeys: []string{"item:orange"},
			expectedError:         false,
		},
		{
			name: "should delete matching keys when using infix pattern",
			setup: func(m *MemoryStore) {
				m.data.Store("prefix:middle:suffix", memoryItem{value: "1"})
				m.data.Store("another:middle:one", memoryItem{value: "2"})
				m.data.Store("no:match", memoryItem{value: "3"})
			},
			pattern:               "*:middle:*",
			expectedRemainingKeys: []string{"no:match"},
			expectedError:         false,
		},
		{
			name: "should not delete any keys when no key matches the pattern",
			setup: func(m *MemoryStore) {
				m.data.Store("user:1", memoryItem{value: "John"})
				m.data.Store("product:1", memoryItem{value: "Laptop"})
			},
			pattern:               "nonexistent:*",
			expectedRemainingKeys: []string{"user:1", "product:1"},
			expectedError:         false,
		},
		{
			name: "should handle gracefully when store is empty",
			setup: func(m *MemoryStore) {
				// empty
			},
			pattern:               "user:*",
			expectedRemainingKeys: []string{},
			expectedError:         false,
		},
		{
			name: "should delete matching keys when pattern includes special characters",
			setup: func(m *MemoryStore) {
				m.data.Store("key.1", memoryItem{value: "val1"})
				m.data.Store("key-2", memoryItem{value: "val2"})
				m.data.Store("key:3", memoryItem{value: "val3"})
				m.data.Store("other_key", memoryItem{value: "val4"})
			},
			pattern:               "key*",
			expectedRemainingKeys: []string{"other_key"},
			expectedError:         false,
		},
		{
			name: "should delete matching keys when using multiple wildcards",
			setup: func(m *MemoryStore) {
				m.data.Store("a:b:c", memoryItem{value: "1"})
				m.data.Store("a:x:c", memoryItem{value: "2"})
				m.data.Store("a:b:d", memoryItem{value: "3"})
				m.data.Store("x:y:z", memoryItem{value: "4"})
			},
			pattern:               "a:*:c",
			expectedRemainingKeys: []string{"a:b:d", "x:y:z"},
			expectedError:         false,
		},
		{
			name: "should perform no operation when pattern is empty",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "value1"})
			},
			pattern:               "",
			expectedRemainingKeys: []string{"key1"},
			expectedError:         false,
		},
		{
			name: "should stop iteration early when context is canceled before execution",
			setup: func(m *MemoryStore) {
				// Populate many keys to ensure Range() iteration starts
				for i := 0; i < 100; i++ {
					m.data.Store(fmt.Sprintf("key:%d", i), memoryItem{value: i})
				}
			},
			pattern:               "key:*",
			expectedRemainingKeys: nil, // we don't verify keys here
			expectedError:         false,
			cancelCtx:             true,
		},
		{
			name: "should return error when invalid regex pattern is provided without wildcard",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "value1"})
			},
			pattern:               "[", // Invalid regex
			expectedRemainingKeys: []string{"key1"},
			expectedError:         true,
		},
		{
			name: "should skip non-string keys when encountered during deletion",
			setup: func(m *MemoryStore) {
				m.data.Store(12345, memoryItem{value: "int-key"})
				m.data.Store("valid:string:key", memoryItem{value: "string-key"})
			},
			pattern:               "valid:*",
			expectedRemainingKeys: []string{}, // non-string key is ignored
			expectedError:         false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			store := &MemoryStore{}
			tt.setup(store)

			var ctx context.Context
			if tt.cancelCtx {
				// Create a context that is canceled before DeleteByPattern starts
				c, cancel := context.WithCancel(context.Background())
				cancel() // immediately cancel
				ctx = c
			} else {
				ctx = context.Background()
			}

			// --- Act ---
			err := store.DeleteByPattern(ctx, tt.pattern)

			// --- Assert ---
			if tt.expectedError {
				assert.Error(t, err, "expected an error for invalid pattern")
				return
			}

			assert.NoError(t, err, "expected no error from DeleteByPattern")

			// Skip remaining key validation if context was canceled
			if tt.cancelCtx {
				return
			}

			// Verify remaining keys
			remainingKeys := make(map[string]struct{})
			store.data.Range(func(k, _ any) bool {
				keyStr, ok := k.(string)
				if !ok {
					// Skip non-string keys
					return true
				}
				remainingKeys[keyStr] = struct{}{}
				return true
			})
			assert.Equal(t, len(tt.expectedRemainingKeys), len(remainingKeys), "number of remaining keys mismatch")

			for _, expectedKey := range tt.expectedRemainingKeys {
				_, exists := remainingKeys[expectedKey]
				assert.True(t, exists, "expected key to remain in cache")
			}
		})
	}
}

func TestMemoryStore_DeleteMany(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		setup                 func(*MemoryStore)
		keysToDelete          []string
		expectedRemainingKeys []string // keys expected to remain after deletion
	}{
		{
			name: "should delete keys when multiple exist",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "value1"})
				m.data.Store("key2", memoryItem{value: "value2"})
				m.data.Store("key3", memoryItem{value: "value3"})
			},
			keysToDelete:          []string{"key1", "key3"},
			expectedRemainingKeys: []string{"key2"},
		},
		{
			name: "should ignore non-existent keys when deleting",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "value1"})
			},
			keysToDelete:          []string{"key1", "nonexistent_key", "another_nonexistent"},
			expectedRemainingKeys: []string{},
		},
		{
			name: "should do nothing when delete list is empty",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "value1"})
			},
			keysToDelete:          []string{},
			expectedRemainingKeys: []string{"key1"},
		},
		{
			name: "should do nothing when store is empty",
			setup: func(m *MemoryStore) {
				// empty
			},
			keysToDelete:          []string{"key1", "key2"},
			expectedRemainingKeys: []string{},
		},
		{
			name: "should delete all keys when all provided",
			setup: func(m *MemoryStore) {
				m.data.Store("keyA", memoryItem{value: "valA"})
				m.data.Store("keyB", memoryItem{value: "valB"})
				m.data.Store("keyC", memoryItem{value: "valC"})
			},
			keysToDelete:          []string{"keyA", "keyB", "keyC"},
			expectedRemainingKeys: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			store := &MemoryStore{}
			tt.setup(store)

			// --- Act ---
			err := store.DeleteMany(ctx, tt.keysToDelete...)

			// --- Assert ---
			assert.NoError(t, err, "expected no error from DeleteMany")

			// Verify remaining keys
			remainingKeys := make(map[string]struct{})
			store.data.Range(func(k, _ any) bool {
				keyStr, ok := k.(string)
				assert.True(t, ok, "expected key to be a string")
				remainingKeys[keyStr] = struct{}{}
				return true
			})

			assert.Equal(t, len(tt.expectedRemainingKeys), len(remainingKeys), "number of remaining keys mismatch")

			for _, expectedKey := range tt.expectedRemainingKeys {
				_, exists := remainingKeys[expectedKey]
				assert.True(t, exists, "expected key to remain in cache")
			}
		})
	}
}

func TestMemoryStore_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(*MemoryStore)
		key           string
		expectedValue any
		expectedError error
	}{
		{
			name: "should return value when key exists and not expired",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "value1", expiration: time.Time{}}) // No expiration
			},
			key:           "key1",
			expectedValue: "value1",
			expectedError: nil,
		},
		{
			name: "should return value when key has future expiration",
			setup: func(m *MemoryStore) {
				m.data.Store("future_key", memoryItem{value: "future_value", expiration: time.Now().Add(1 * time.Hour)})
			},
			key:           "future_key",
			expectedValue: "future_value",
			expectedError: nil,
		},
		{
			name: "should return correct value when key stores different data type",
			setup: func(m *MemoryStore) {
				m.data.Store("int_key", memoryItem{value: 123, expiration: time.Time{}})
				m.data.Store("bool_key", memoryItem{value: true, expiration: time.Time{}})
			},
			key:           "int_key",
			expectedValue: 123,
			expectedError: nil,
		},
		{
			name: "should return boolean value when key stores a boolean",
			setup: func(m *MemoryStore) {
				m.data.Store("bool_key", memoryItem{value: true, expiration: time.Time{}})
			},
			key:           "bool_key",
			expectedValue: true,
			expectedError: nil,
		},
		{
			name: "should return ErrCacheMiss when key does not exist",
			setup: func(m *MemoryStore) {
				// empty store
			},
			key:           "nonexistent_key",
			expectedValue: nil,
			expectedError: omnicache.ErrCacheMiss,
		},
		{
			name: "should return ErrCacheMiss when key is expired and remove it",
			setup: func(m *MemoryStore) {
				m.data.Store("expired_key", memoryItem{value: "expired_value", expiration: time.Now().Add(-1 * time.Second)})
			},
			key:           "expired_key",
			expectedValue: nil,
			expectedError: omnicache.ErrCacheMiss,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			store := &MemoryStore{}
			tt.setup(store)

			// --- Act ---
			value, err := store.Get(ctx, tt.key)

			// --- Assert ---
			if tt.expectedError != nil {
				assert.EqualError(t, tt.expectedError, err, "expected error mismatch")
				assert.Nil(t, value, "expected nil value on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.expectedValue, value, "expected value mismatch")
			}

			// If key was expected to be expired and deleted, verify its absence
			if errors.Is(tt.expectedError, omnicache.ErrCacheMiss) {
				_, exists := store.data.Load(tt.key)
				assert.False(t, exists, "expected expired key to be deleted from store")
			}
		})
	}
}

func TestMemoryStore_Has(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(*MemoryStore)
		key           string
		expectedHas   bool
		expectedError error
	}{
		{
			name: "should return true when key exists and not expired",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "value1", expiration: time.Time{}}) // No expiration
			},
			key:         "key1",
			expectedHas: true,
		},
		{
			name: "should return true when key has future expiration",
			setup: func(m *MemoryStore) {
				m.data.Store("future_key", memoryItem{value: "future_value", expiration: time.Now().Add(1 * time.Hour)})
			},
			key:         "future_key",
			expectedHas: true,
		},
		{
			name: "should return false when key does not exist",
			setup: func(m *MemoryStore) {
				// empty store
			},
			key:         "nonexistent_key",
			expectedHas: false,
		},
		{
			name: "should return false when key is expired and remove it",
			setup: func(m *MemoryStore) {
				m.data.Store("expired_key", memoryItem{value: "expired_value", expiration: time.Now().Add(-1 * time.Second)})
			},
			key:         "expired_key",
			expectedHas: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			store := &MemoryStore{}
			tt.setup(store)

			// --- Act ---
			has, err := store.Has(ctx, tt.key)

			// --- Assert ---
			assert.NoError(t, err, "expected no error from Has")
			assert.Equal(t, tt.expectedHas, has, "expected Has result mismatch")

			// If key was expected to be expired and deleted, verify its absence
			if !tt.expectedHas {
				_, exists := store.data.Load(tt.key)
				assert.False(t, exists, "expected expired key to be deleted from store")
			}
		})
	}
}

func TestMemoryStore_Set(t *testing.T) {
	t.Parallel()

	type args struct {
		key   string
		value any
		ttl   time.Duration
	}

	type User struct {
		ID   int
		Name string
	}

	tests := []struct {
		name          string
		setup         func(*MemoryStore)
		args          args
		expectedError error
		expectedValue any
		expectedTTL   time.Duration
	}{
		{
			name: "should set key when TTL is positive",
			setup: func(m *MemoryStore) {
				// empty
			},
			args: args{
				key:   "key1",
				value: "value1",
				ttl:   1 * time.Minute,
			},
			expectedError: nil,
			expectedValue: "value1",
			expectedTTL:   1 * time.Minute,
		},
		{
			name: "should overwrite existing key when new value and TTL are provided",
			setup: func(m *MemoryStore) {
				m.data.Store("key1", memoryItem{value: "old_value", expiration: time.Now().Add(1 * time.Hour)})
			},
			args: args{
				key:   "key1",
				value: "new_value",
				ttl:   5 * time.Minute,
			},
			expectedError: nil,
			expectedValue: "new_value",
			expectedTTL:   5 * time.Minute,
		},
		{
			name: "should set key when TTL is zero (no expiration)",
			setup: func(m *MemoryStore) {
				// empty
			},
			args: args{
				key:   "key_no_ttl",
				value: 123,
				ttl:   0,
			},
			expectedError: nil,
			expectedValue: 123,
			expectedTTL:   0,
		},
		{
			name:  "should set key when value is boolean",
			setup: func(m *MemoryStore) {},
			args: args{
				key:   "bool_key",
				value: true,
				ttl:   10 * time.Second,
			},
			expectedError: nil,
			expectedValue: true,
			expectedTTL:   10 * time.Second,
		},
		{
			name:  "should set key when value is a slice",
			setup: func(m *MemoryStore) {},
			args: args{
				key:   "slice_key",
				value: []int{1, 2, 3},
				ttl:   15 * time.Second,
			},
			expectedError: nil,
			expectedValue: []int{1, 2, 3},
			expectedTTL:   15 * time.Second,
		},
		{
			name:  "should set key when value is a map",
			setup: func(m *MemoryStore) {},
			args: args{
				key:   "map_key",
				value: map[string]string{"foo": "bar"},
				ttl:   20 * time.Second,
			},
			expectedError: nil,
			expectedValue: map[string]string{"foo": "bar"},
			expectedTTL:   20 * time.Second,
		},
		{
			name:  "should set key when value is a struct",
			setup: func(m *MemoryStore) {},
			args: args{
				key:   "struct_key",
				value: User{ID: 1, Name: "Alice"},
				ttl:   25 * time.Second,
			},
			expectedError: nil,
			expectedValue: User{ID: 1, Name: "Alice"},
			expectedTTL:   25 * time.Second,
		},
		{
			name: "should return error when TTL is negative",
			setup: func(m *MemoryStore) {
				// empty
			},
			args: args{
				key:   "key_invalid_ttl",
				value: "value",
				ttl:   -1 * time.Minute,
			},
			expectedError: omnicache.ErrInvalidValue,
			expectedValue: nil,
			expectedTTL:   0, // Not applicable, as it should error
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// --- Arrange ---
			ctx := context.Background()
			store := &MemoryStore{}
			tt.setup(store)

			// --- Act ---
			err := store.Set(ctx, tt.args.key, tt.args.value, tt.args.ttl)

			// --- Assert ---
			if tt.expectedError != nil {
				assert.EqualError(t, tt.expectedError, err, "expected error mismatch")
				// For negative TTL, the key should not be set or should retain its previous state
				if tt.args.ttl < 0 {
					_, loaded := store.data.Load(tt.args.key)
					assert.False(t, loaded, "expected key not to be loaded for invalid TTL")
				}
				return
			}

			assert.NoError(t, err, "expected no error")

			// Verify the stored item
			item, loaded := store.data.Load(tt.args.key)
			assert.True(t, loaded, "expected key to be in cache")

			memItem, ok := item.(memoryItem)
			assert.True(t, ok, "expected stored item to be of type memoryItem")
			assert.Equal(t, tt.expectedValue, memItem.value, "expected stored value to match")

			if tt.expectedTTL > 0 {
				assert.False(t, memItem.expiration.IsZero(), "expected expiration to be set for positive TTL")
				assert.WithinDuration(t, time.Now().Add(tt.expectedTTL), memItem.expiration, 5*time.Second, "expiration time mismatch")
			} else {
				assert.True(t, memItem.expiration.IsZero(), "expected expiration to be zero for zero TTL")
			}
		})
	}
}

func BenchmarkMemoryStore_deleteExpiredKeys(b *testing.B) {
	// Prepare store with mixed expired and non-expired keys
	store := &MemoryStore{
		data: sync.Map{},
	}

	// Fill the store with 10000 keys, half expired, half not
	const n = 10000
	for i := 0; i < n; i++ {
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

	const n = 10000
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key-%d", i)
		store.data.Store(key, memoryItem{
			value:      i,
			expiration: time.Now().Add(time.Hour),
		})
	}

	for b.Loop() {
		store.Clear(context.Background())
	}
}

func BenchmarkMemoryStore_Delete(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	const n = 10000
	for i := 0; i < n; i++ {
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
				store.Delete(context.Background(), tt.key)
			}
		})
	}
}

func BenchmarkMemoryStore_DeleteByPattern(b *testing.B) {
	store := &MemoryStore{
		data: sync.Map{},
	}

	numKeys := 100_000
	for i := 0; i < numKeys; i++ {
		if i%2 == 0 {
			store.Set(context.Background(), fmt.Sprintf("auth:tenant:%d:user:%d:access_token:%d", i%100, i%1000, i), "value", 0)
		} else {
			store.Set(context.Background(), fmt.Sprintf("user_permissions:tenant:%d:user:%d", i%100, i%1000), "value", 0)
		}
	}

	for b.Loop() {
		store.DeleteByPattern(context.Background(), "auth:tenant:*:user:123:access_token:*")
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
				store.DeleteMany(context.Background(), tt.keys...)
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
				store.Get(context.Background(), tt.key)
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
				store.Has(context.Background(), tt.key)
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
				store.Set(context.Background(), tt.key, tt.value, tt.ttl)
			}
		})
	}
}
