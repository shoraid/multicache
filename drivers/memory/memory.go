package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/shoraid/multicache"
)

type MemoryStore struct {
	data          map[string]memoryItem
	mutex         sync.RWMutex
	cancelCleanup context.CancelFunc
	nowFunc       func() time.Time
}

type memoryItem struct {
	value      any
	expiration time.Time
}

// NewMemoryStore creates a new in-memory cache store.
// It starts a background goroutine to periodically clean up expired keys.
// The cleanup interval can be provided via the config map under "cleanup_interval".
// If not provided, a default interval of 10 minutes is used.
func NewMemoryStore(config map[string]any) (multicache.Store, error) {
	store := &MemoryStore{
		data: make(map[string]memoryItem), // initialize the internal map
	}

	// Set the cleanup interval from config, or use default
	cleanupInterval := multicache.DefaultCleanupInterval

	if v, exists := config["cleanup_interval"]; exists {
		if d, isDuration := v.(time.Duration); isDuration && d > 0 {
			cleanupInterval = d
		}
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
	keys := m.findExpiredKeys()
	if len(keys) == 0 {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, k := range keys {
		delete(m.data, k)
	}
}

// findExpiredKeys returns a slice of expired keys without deleting them.
func (m *MemoryStore) findExpiredKeys() []string {
	keys := []string{}
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for k, v := range m.data {
		if !v.expiration.IsZero() && m.now().After(v.expiration) {
			keys = append(keys, k)
		}
	}
	return keys
}

// now returns the current time. It can be overridden by a custom function (useful for testing)
func (m *MemoryStore) now() time.Time {
	if m.nowFunc != nil {
		return m.nowFunc()
	}

	return time.Now()
}

// Add inserts a key-value pair into the MemoryStore with an optional TTL.
// Returns ErrItemAlreadyExists if the key exists and is not expired.
// Overwrites the key if it is expired. Propagates errors from Has or Put.
func (m *MemoryStore) Add(key string, value any, ttl ...time.Duration) error {
	// Check if the key already exists in the store
	exists, err := m.Has(key)
	if err != nil {
		return err
	}

	// If the key exists and is not expired, return ErrItemAlreadyExists
	if exists {
		return multicache.ErrItemAlreadyExists
	}

	// Key does not exist or is expired, so store the new value with optional TTL
	err = m.Put(key, value, ttl...)
	if err != nil {
		return err
	}

	// Successfully added the key-value pair
	return nil
}

// Flush clears all entries from the cache.
//
// This resets the internal storage by replacing it
// with a new empty map. After calling Flush, the
// cache will be completely empty.
func (m *MemoryStore) Flush() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Clear all cache entries by reinitializing the map
	m.data = make(map[string]memoryItem)

	return nil
}

// Forget removes the given key from the cache.
//
// If the key does not exist, the call has no effect.
func (m *MemoryStore) Forget(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.data, key)

	return nil
}

// Get returns the value stored in the cache for the given key.
//
// If the key does not exist or is expired:
//   - It returns the first fallback value if provided.
//   - Otherwise, it returns (nil, ErrCacheMiss).
//
// If the key exists and is valid, the stored value is returned.
func (m *MemoryStore) Get(key string, fallback ...any) (any, error) {
	m.mutex.RLock()
	item, found := m.data[key]
	m.mutex.RUnlock()

	// Key does not exist in the cache
	if !found {
		// Return the first fallback value if provided
		if len(fallback) > 0 {
			return fallback[0], nil
		}

		// Return error cache miss if no fallback is provided
		return nil, multicache.ErrCacheMiss
	}

	// Key exists but has expired
	if !item.expiration.IsZero() && m.now().After(item.expiration) {
		m.mutex.Lock()
		// Remove the expired key
		delete(m.data, key)
		m.mutex.Unlock()

		// Return the first fallback value if provided
		if len(fallback) > 0 {
			return fallback[0], nil
		}

		// Return error cache miss if no fallback is provided
		return nil, multicache.ErrCacheMiss
	}

	// Key exists and is not expired
	return item.value, nil
}

// Has returns true if the key exists in the cache, false if missing or expired.
// Returns an error only if a non-cache-related issue occurs.
func (m *MemoryStore) Has(key string) (bool, error) {
	// Try to get the value for the given key
	_, err := m.Get(key)
	if err != nil {
		// If the key is missing or expired, treat it as not existing
		if errors.Is(err, multicache.ErrCacheMiss) {
			return false, nil
		}
		// If any other error occurred, propagate it
		return false, err
	}

	// Key exists and has a valid value
	return true, nil
}

// Put stores a value in the cache with the given key.
//
// Behavior:
//   - If a positive TTL is provided, the entry will expire after that duration.
//   - If TTL is zero or not provided, the entry will never expire.
//   - If a negative TTL is provided, the key will be immediately deleted.
func (m *MemoryStore) Put(key string, value any, ttl ...time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(ttl) > 0 {
		if ttl[0] < 0 {
			// Negative TTL: immediately delete the key
			delete(m.data, key)
			return nil
		}
		// TTL = 0: cache forever, handled below
	}

	var expiration time.Time
	if len(ttl) > 0 && ttl[0] > 0 {
		// Positive TTL: set expiration time
		expiration = m.now().Add(ttl[0])
	} else {
		// TTL 0 or not provided: cache forever
		expiration = time.Time{}
	}

	// Store the value with expiration
	m.data[key] = memoryItem{
		value,
		expiration,
	}

	return nil
}
