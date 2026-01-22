// Package extraction provides tests for memory cache
package extraction

import (
	"fmt"
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	t.Run("creates cache with defaults", func(t *testing.T) {
		cache := NewMemoryCache(0, 0)
		if cache == nil {
			t.Error("cache should not be nil")
		}
	})

	t.Run("creates cache with custom size and TTL", func(t *testing.T) {
		cache := NewMemoryCache(100, 1*time.Hour)
		if cache == nil {
			t.Error("cache should not be nil")
		}
	})
}

func TestMemoryCache_GetSet(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Hour).(*MemoryCache)

	t.Run("sets and gets value", func(t *testing.T) {
		cache.Set("key1", "value1", 100)
		value, ok := cache.Get("key1")
		if !ok {
			t.Error("should find key1")
		}
		if value != "value1" {
			t.Errorf("expected value1, got %s", value)
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		_, ok := cache.Get("nonexistent")
		if ok {
			t.Error("should not find nonexistent key")
		}
	})

	t.Run("evicts oldest when at capacity", func(t *testing.T) {
		// Fill cache to capacity
		for i := 0; i < 10; i++ {
			cache.Set(string(rune('a'+i)), "value", 100)
		}

		// Add one more to trigger eviction
		cache.Set("newkey", "newvalue", 100)

		// First key should be evicted
		_, ok := cache.Get("a")
		if ok {
			t.Error("oldest key should be evicted")
		}

		// New key should exist
		value, ok := cache.Get("newkey")
		if !ok {
			t.Error("new key should exist")
		}
		if value != "newvalue" {
			t.Errorf("expected newvalue, got %s", value)
		}
	})

	t.Run("expires entries after TTL", func(t *testing.T) {
		shortCache := NewMemoryCache(10, 100*time.Millisecond).(*MemoryCache)
		shortCache.Set("expire_key", "value", 100)

		// Should exist immediately
		_, ok := shortCache.Get("expire_key")
		if !ok {
			t.Error("key should exist immediately")
		}

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		_, ok = shortCache.Get("expire_key")
		if ok {
			t.Error("key should be expired")
		}
	})

	t.Run("handles concurrent access", func(t *testing.T) {
		cache := NewMemoryCache(100, 1*time.Hour).(*MemoryCache)
		done := make(chan bool, 20)

		// Concurrent writes
		for i := 0; i < 10; i++ {
			go func(idx int) {
				cache.Set(fmt.Sprintf("key_%d", idx), fmt.Sprintf("value_%d", idx), 100)
				done <- true
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 10; i++ {
			go func(idx int) {
				val, ok := cache.Get(fmt.Sprintf("key_%d", idx))
				_ = val
				_ = ok
				done <- true
			}(i)
		}

		// Wait for all operations
		for i := 0; i < 20; i++ {
			<-done
		}
	})

	t.Run("evicts multiple entries when needed", func(t *testing.T) {
		cache := NewMemoryCache(5, 1*time.Hour).(*MemoryCache)

		// Fill to capacity
		for i := 0; i < 5; i++ {
			cache.Set(fmt.Sprintf("key_%d", i), "value", 100)
		}

		// Add 3 more entries - should evict 3 oldest
		cache.Set("new1", "value1", 100)
		cache.Set("new2", "value2", 100)
		cache.Set("new3", "value3", 100)

		// Verify new entries exist
		_, ok := cache.Get("new1")
		if !ok {
			t.Error("new1 should exist")
		}
		_, ok = cache.Get("new2")
		if !ok {
			t.Error("new2 should exist")
		}
		_, ok = cache.Get("new3")
		if !ok {
			t.Error("new3 should exist")
		}

		// Verify cache size is at capacity
		cache.mu.RLock()
		size := len(cache.entries)
		cache.mu.RUnlock()
		if size > 5 {
			t.Errorf("cache size should be at most 5, got %d", size)
		}
	})

	t.Run("cleanup removes expired entries", func(t *testing.T) {
		// Use longer TTL so key2 doesn't expire during test
		cache := NewMemoryCache(10, 1*time.Hour).(*MemoryCache)

		// Set entries
		cache.Set("key1", "value1", 100)
		cache.Set("key2", "value2", 100)

		// Manually expire one entry
		cache.mu.Lock()
		if entry, ok := cache.entries["key1"]; ok {
			entry.expiresAt = time.Now().Add(-time.Hour)
			cache.entries["key1"] = entry
		}
		cache.mu.Unlock()

		// Get should remove expired entry
		_, ok := cache.Get("key1")
		if ok {
			t.Error("expired key1 should be removed")
		}

		// key2 should still exist (not expired)
		val, ok := cache.Get("key2")
		if !ok {
			t.Error("key2 should still exist")
		}
		if val != "value2" {
			t.Errorf("expected value2, got %s", val)
		}
	})

	t.Run("handles zero maxSize", func(t *testing.T) {
		cache := NewMemoryCache(0, 1*time.Hour).(*MemoryCache)
		cache.Set("key1", "value1", 100)

		val, ok := cache.Get("key1")
		if !ok {
			t.Error("should find key1")
		}
		if val != "value1" {
			t.Errorf("expected value1, got %s", val)
		}
	})

	t.Run("handles zero TTL", func(t *testing.T) {
		cache := NewMemoryCache(10, 0).(*MemoryCache)
		cache.Set("key1", "value1", 100)

		// Should use default TTL (24 hours)
		val, ok := cache.Get("key1")
		if !ok {
			t.Error("should find key1")
		}
		if val != "value1" {
			t.Errorf("expected value1, got %s", val)
		}
	})
}
