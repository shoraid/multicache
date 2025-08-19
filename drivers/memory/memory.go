package memory

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/shoraid/multicache"
)

type MemoryStore struct {
	data          sync.Map
	cancelCleanup context.CancelFunc
}

type memoryItem struct {
	value      any
	expiration time.Time
}

// NewMemoryStore creates a new in-memory cache store.
// It starts a background goroutine to periodically clean up expired keys.
// The cleanup interval can be provided via the config map under "cleanup_interval".
// If not provided, a default interval of 10 minutes is used.
func NewMemoryStore(config MemoryConfig) (multicache.Store, error) {
	store := &MemoryStore{}

	// Set the cleanup interval from config, or use default
	cleanupInterval := multicache.DefaultCleanupInterval

	if config.CleanupInterval > 0 {
		cleanupInterval = config.CleanupInterval
	}

	// Start the background goroutine for cleaning up expired keys
	ctx, cancel := context.WithCancel(context.Background())
	store.cancelCleanup = cancel

	go store.cleanupExpiredKeys(ctx, cleanupInterval)

	return store, nil
}

// cleanupExpiredKeys runs in a background goroutine to remove expired keys at regular intervals.
func (m *MemoryStore) cleanupExpiredKeys(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.deleteExpiredKeys()
		case <-ctx.Done():
			return
		}
	}
}

// deleteExpiredKeys finds and removes all expired keys from the store.
func (m *MemoryStore) deleteExpiredKeys() {
	now := time.Now()
	m.data.Range(func(k, v any) bool {
		item := v.(memoryItem)
		if !item.expiration.IsZero() && now.After(item.expiration) {
			m.data.Delete(k)
		}
		return true
	})
}

// Clear removes all entries from the cache.
//
// Behavior:
//   - All keys and values currently stored will be deleted.
//   - This operation affects the entire cache, regardless of TTL or expiration.
//
// Use cases:
//   - Useful for testing, resetting state, or administrative cleanup.
//   - Should be used carefully in production as it clears all data.
func (m *MemoryStore) Clear() error {
	m.data.Clear()

	return nil
}

// Delete removes the entry associated with the given key from the cache.
//
// Behavior:
//   - If the key exists, it is removed immediately.
//   - If the key does not exist, the operation is a no-op (no error is returned).
//
// Use cases:
//   - Explicit cache invalidation for a single key.
//   - Useful when data becomes stale or needs to be refreshed.
func (m *MemoryStore) Delete(key string) error {
	m.data.Delete(key)

	return nil
}

// DeleteByPattern removes all cache entries whose keys match the given pattern.
//
// Pattern rules:
//   - The pattern supports '*' as a wildcard, matching zero or more characters.
//   - Other characters are treated literally.
//
// Example:
//
//	store.DeleteByPattern("user:*") // deletes all keys starting with "user:"
//
// Returns an error if the pattern cannot be compiled into a valid regular expression.
//
// Performance note:
//   - This operation iterates over all keys in the cache and performs regex
//     matching on each key.
//   - On large caches, this can be slow and should be used sparingly
//     (e.g., for administrative cleanup rather than frequent operations).
func (m *MemoryStore) DeleteByPattern(pattern string) error {
	regexPattern := "^" + regexp.QuoteMeta(pattern) + "$"
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", ".*")

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return err
	}

	m.data.Range(func(k, _ any) bool {
		keyStr, ok := k.(string)
		if ok && re.MatchString(keyStr) {
			m.data.Delete(k)
		}
		return true
	})

	return nil
}

// DeleteMany removes multiple keys from the cache in a single call.
//
// Behavior:
//   - Iterates over the provided keys and deletes each one from the cache.
//   - If a key does not exist, it is silently ignored (no error).
//   - Returns nil always, since deletion is best-effort and non-critical.
func (m *MemoryStore) DeleteMany(keys ...string) error {
	for _, key := range keys {
		m.data.Delete(key)
	}

	return nil
}

// Get retrieves a value from the cache by key.
//
// Behavior:
//   - If the key does not exist, returns (nil, ErrCacheMiss).
//   - If the key exists but is expired, deletes it and returns (nil, ErrCacheMiss).
//   - If the key exists and is valid, returns (value, nil).
//
// Notes:
//   - Expiration is checked at read time, expired entries are lazily removed.
//   - Thread-safe since sync.Map is used internally.
func (m *MemoryStore) Get(key string) (any, error) {
	value, exists := m.data.Load(key)
	if !exists {
		return nil, multicache.ErrCacheMiss
	}

	item := value.(memoryItem)

	// Key exists but expired
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		m.data.Delete(key)
		return nil, multicache.ErrCacheMiss
	}

	return item.value, nil
}

// GetOrSet retrieves the value for the given key from the cache.
//
// Behavior:
//   - If the key exists and is valid, the cached value is returned.
//   - If the key is missing or expired, the provided value is stored with the given TTL
//     and then returned.
//   - If any error other than ErrCacheMiss occurs, it is returned directly.
func (m *MemoryStore) GetOrSet(key string, ttl time.Duration, value any) (any, error) {
	item, err := m.Get(key)
	if err == nil {
		return item, nil
	}

	if !errors.Is(err, multicache.ErrCacheMiss) {
		return nil, err
	}

	if err := m.Set(key, value, ttl); err != nil {
		return nil, err
	}

	return value, nil
}

// Has checks whether a key exists and is still valid in the cache.
//
// Behavior:
//   - Returns true if the key is present and not expired.
//   - Returns false if the key is missing or expired.
func (m *MemoryStore) Has(key string) (bool, error) {
	_, err := m.Get(key)
	if err != nil {
		return false, nil
	}

	// Key exists and has a valid value
	return true, nil
}

// Set stores a value in the cache with the given key.
//
// Behavior:
//   - TTL > 0: entry expires after duration
//   - TTL = 0: entry never expires
//   - TTL < 0: returns ErrInvalidValue
//
// Existing keys are overwritten. Thread-safe via sync.Map.
func (m *MemoryStore) Set(key string, value any, ttl time.Duration) error {
	if ttl < 0 {
		return multicache.ErrInvalidValue
	}

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	} else {
		// TTL 0 means no expiration
		expiration = time.Time{}
	}

	// Store the value with expiration
	m.data.Store(key, memoryItem{
		value,
		expiration,
	})

	return nil
}
