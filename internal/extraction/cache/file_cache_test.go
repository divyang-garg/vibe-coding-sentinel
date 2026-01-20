// Package cache provides file-based caching for extraction results
package cache

import (
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
}
