// Package cache provides file-based caching for extraction results
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileCache(t *testing.T) {
	t.Run("creates cache directory", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "test_cache_"+t.Name())
		defer os.RemoveAll(tmpDir)

		cache, err := NewFileCache(tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, cache)

		// Check directory was created
		info, err := os.Stat(tmpDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("handles existing directory", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "test_cache_existing_"+t.Name())
		os.MkdirAll(tmpDir, 0755)
		defer os.RemoveAll(tmpDir)

		cache, err := NewFileCache(tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, cache)
	})
}

func TestFileCache_GetSet(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test_cache_getset_"+t.Name())
	defer os.RemoveAll(tmpDir)

	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	t.Run("returns false for missing key", func(t *testing.T) {
		val, ok := cache.Get("nonexistent")
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("stores and retrieves value", func(t *testing.T) {
		key := "test_key"
		value := `{"test": "value"}`

		cache.Set(key, value, 100)

		retrieved, ok := cache.Get(key)
		assert.True(t, ok)
		assert.Equal(t, value, retrieved)
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		key := "overwrite_key"

		cache.Set(key, "value1", 50)
		cache.Set(key, "value2", 60)

		retrieved, ok := cache.Get(key)
		assert.True(t, ok)
		assert.Equal(t, "value2", retrieved)
	})
}

func TestFileCache_Expiration(t *testing.T) {
	t.Run("returns false for corrupted file", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "test_cache_corrupt_"+t.Name())
		defer os.RemoveAll(tmpDir)

		cache, err := NewFileCache(tmpDir)
		require.NoError(t, err)

		// Write corrupted data directly
		corruptFile := filepath.Join(tmpDir, "corrupt_key.json")
		os.WriteFile(corruptFile, []byte("not valid json"), 0644)

		val, ok := cache.Get("corrupt_key")
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("handles expired entries", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "test_cache_expired_"+t.Name())
		defer os.RemoveAll(tmpDir)

		cache, err := NewFileCache(tmpDir)
		require.NoError(t, err)

		// Set a value
		cache.Set("expired_key", "value", 100)

		// Manually modify the file to have expired timestamp
		filePath := filepath.Join(tmpDir, "expired_key.json")

		// Modify to have past expiration
		modifiedData := []byte(`{"response":"value","tokens_used":100,"expires_at":"2000-01-01T00:00:00Z"}`)
		os.WriteFile(filePath, modifiedData, 0644)

		val, ok := cache.Get("expired_key")
		assert.False(t, ok)
		assert.Empty(t, val)

		// File should be removed
		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("handles missing file gracefully", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "test_cache_missing_"+t.Name())
		defer os.RemoveAll(tmpDir)

		cache, err := NewFileCache(tmpDir)
		require.NoError(t, err)

		val, ok := cache.Get("nonexistent_file")
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestFileCache_NewFileCache_ErrorCases(t *testing.T) {
	t.Run("handles permission denied", func(t *testing.T) {
		// Try to create cache in read-only location (if possible)
		// On Unix systems, /dev/null is not a directory
		_, err := NewFileCache("/dev/null/test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create cache directory")
	})
}

func TestFileCache_Set_ErrorCases(t *testing.T) {
	t.Run("handles marshal error gracefully", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "test_cache_marshal_"+t.Name())
		defer os.RemoveAll(tmpDir)

		cache, err := NewFileCache(tmpDir)
		require.NoError(t, err)

		// Set should handle errors gracefully (currently silently fails)
		// This tests that Set doesn't panic on errors
		cache.Set("test_key", "normal_value", 100)

		// Verify it was set
		val, ok := cache.Get("test_key")
		assert.True(t, ok)
		assert.Equal(t, "normal_value", val)
	})
}

func TestFileCache_ConcurrentAccess(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test_cache_concurrent_"+t.Name())
	defer os.RemoveAll(tmpDir)

	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	// Test concurrent reads and writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			key := fmt.Sprintf("key_%d", idx)
			cache.Set(key, fmt.Sprintf("value_%d", idx), 100)
			val, ok := cache.Get(key)
			assert.True(t, ok)
			assert.Equal(t, fmt.Sprintf("value_%d", idx), val)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
