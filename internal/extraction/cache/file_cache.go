// Package cache provides file-based caching for extraction results
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileCache implements extraction.Cache using file system
type FileCache struct {
	dir string
	mu  sync.RWMutex
}

// NewFileCache creates a new file-based cache
func NewFileCache(cacheDir string) (*FileCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &FileCache{dir: cacheDir}, nil
}

// Get retrieves a cached value
func (c *FileCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := filepath.Join(c.dir, key+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", false
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", false
	}

	if time.Now().After(entry.ExpiresAt) {
		os.Remove(filePath) // Clean up expired entry
		return "", false
	}

	return entry.Response, true
}

// Set stores a value in cache
func (c *FileCache) Set(key string, value string, tokensUsed int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := cacheEntry{
		Response:   value,
		TokensUsed: tokensUsed,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return // Silently fail on marshal error
	}

	filePath := filepath.Join(c.dir, key+".json")
	os.WriteFile(filePath, data, 0644)
}

type cacheEntry struct {
	Response   string    `json:"response"`
	TokensUsed int       `json:"tokens_used"`
	ExpiresAt  time.Time `json:"expires_at"`
}
