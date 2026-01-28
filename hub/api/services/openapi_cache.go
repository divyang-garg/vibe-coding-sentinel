// Package services provides caching for OpenAPI contract parsing
// Complies with CODING_STANDARDS.md: Business services max 400 lines per file
package services

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// ContractCache provides caching for parsed OpenAPI contracts
type ContractCache struct {
	mu        sync.RWMutex
	contracts map[string]*CachedContract
	ttl       time.Duration
}

// CachedContract represents a cached contract with expiration
type CachedContract struct {
	Contract  *OpenAPIContract
	ExpiresAt time.Time
	FilePath  string
	FileModTime time.Time
}

var (
	globalCache *ContractCache
	cacheOnce   sync.Once
)

// GetContractCache returns the global contract cache instance
func GetContractCache() *ContractCache {
	cacheOnce.Do(func() {
		globalCache = &ContractCache{
			contracts: make(map[string]*CachedContract),
			ttl:       5 * time.Minute, // Default TTL: 5 minutes
		}
	})
	return globalCache
}

// GetCachedContract retrieves cached contract or parses new one
// Checks file modification time to invalidate cache if file changed
func GetCachedContract(ctx context.Context, filePath string) (*OpenAPIContract, error) {
	cache := GetContractCache()

	// Check cache first
	cache.mu.RLock()
	cached, exists := cache.contracts[filePath]
	cache.mu.RUnlock()

	if exists {
		// Check if cache is still valid
		now := time.Now()
		if now.Before(cached.ExpiresAt) {
			// Check file modification time
			fileInfo, err := os.Stat(filePath)
			if err == nil {
				if fileInfo.ModTime().Equal(cached.FileModTime) || fileInfo.ModTime().Before(cached.FileModTime) {
					// Cache is valid and file hasn't changed
					return cached.Contract, nil
				}
			}
		}
	}

	// Cache miss or expired - parse new contract
	contract, err := ParseOpenAPIContract(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract: %w", err)
	}

	// Get file modification time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return contract, nil // Return contract even if stat fails
	}

	// Store in cache
	cache.mu.Lock()
	cache.contracts[filePath] = &CachedContract{
		Contract:    contract,
		ExpiresAt:   time.Now().Add(cache.ttl),
		FilePath:    filePath,
		FileModTime: fileInfo.ModTime(),
	}
	cache.mu.Unlock()

	return contract, nil
}

// SetTTL sets the cache TTL (time-to-live)
func (c *ContractCache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}

// ClearCache clears all cached contracts
func (c *ContractCache) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.contracts = make(map[string]*CachedContract)
}

// RemoveCache removes a specific contract from cache
func (c *ContractCache) RemoveCache(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.contracts, filePath)
}

// GetCacheStats returns cache statistics
func (c *ContractCache) GetCacheStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	validCount := 0
	expiredCount := 0

	for _, cached := range c.contracts {
		if now.Before(cached.ExpiresAt) {
			validCount++
		} else {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"total_contracts": len(c.contracts),
		"valid_contracts": validCount,
		"expired_contracts": expiredCount,
		"ttl_seconds": c.ttl.Seconds(),
	}
}

// CleanupExpired removes expired entries from cache
func (c *ContractCache) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for filePath, cached := range c.contracts {
		if now.After(cached.ExpiresAt) {
			delete(c.contracts, filePath)
		}
	}
}
