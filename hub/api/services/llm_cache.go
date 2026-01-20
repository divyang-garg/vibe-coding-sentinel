// LLM Cache - Basic Cache Operations
// Implements basic LLM response caching with TTL
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// LLMCacheEntry represents a cached LLM response
type LLMCacheEntry struct {
	Response   string
	TokensUsed int
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// LLM cache with TTL
var llmCache = make(map[string]*LLMCacheEntry)
var llmCacheMutex sync.RWMutex
var llmCacheTTL = 24 * time.Hour

const maxCacheSize = 1000 // Maximum number of cache entries

// getCachedLLMResponse retrieves a cached LLM response if available
// Phase 14D: Now respects UseCache config flag
func getCachedLLMResponse(cacheKey string, config *LLMConfig) (string, bool) {
	// Check if caching is enabled
	if config != nil && !config.CostOptimization.UseCache {
		return "", false
	}

	llmCacheMutex.RLock()
	defer llmCacheMutex.RUnlock()

	entry, ok := llmCache[cacheKey]
	if !ok {
		return "", false
	}

	// Check if cache entry is expired
	if time.Now().After(entry.ExpiresAt) {
		// Entry expired, remove it
		llmCacheMutex.RUnlock()
		llmCacheMutex.Lock()
		delete(llmCache, cacheKey)
		llmCacheMutex.Unlock()
		llmCacheMutex.RLock()
		return "", false
	}

	return entry.Response, true
}

// cleanupLLMCache removes expired entries and enforces size limit
func cleanupLLMCache() {
	llmCacheMutex.Lock()
	defer llmCacheMutex.Unlock()

	// Remove expired entries
	now := time.Now()
	for key, entry := range llmCache {
		if now.After(entry.ExpiresAt) {
			delete(llmCache, key)
		}
	}

	// If still too large, remove oldest entries
	if len(llmCache) > maxCacheSize {
		// Convert to slice and sort by CreatedAt
		type cacheEntry struct {
			key   string
			entry *LLMCacheEntry
		}
		entries := make([]cacheEntry, 0, len(llmCache))

		for key, entry := range llmCache {
			entries = append(entries, cacheEntry{key, entry})
		}

		// Sort by CreatedAt (oldest first)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].entry.CreatedAt.Before(entries[j].entry.CreatedAt)
		})

		// Remove oldest entries
		toRemove := len(llmCache) - maxCacheSize
		for i := 0; i < toRemove; i++ {
			delete(llmCache, entries[i].key)
		}
	}
}

// setCachedLLMResponse stores an LLM response in cache
// Phase 14D: Now respects UseCache config flag and uses CacheTTLHours from config
func setCachedLLMResponse(cacheKey string, response string, tokensUsed int, config *LLMConfig) {
	// Check if caching is enabled
	if config != nil && !config.CostOptimization.UseCache {
		return
	}

	llmCacheMutex.Lock()
	defer llmCacheMutex.Unlock()

	// Cleanup if needed before adding new entry
	if len(llmCache) >= maxCacheSize {
		llmCacheMutex.Unlock()
		cleanupLLMCache()
		llmCacheMutex.Lock()
	}

	// Determine TTL from config or use default
	ttl := llmCacheTTL
	if config != nil && config.CostOptimization.CacheTTLHours > 0 {
		ttl = time.Duration(config.CostOptimization.CacheTTLHours) * time.Hour
	}

	llmCache[cacheKey] = &LLMCacheEntry{
		Response:   response,
		TokensUsed: tokensUsed,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(ttl),
	}
}

// generateLLMCacheKey generates a cache key from file hash, analysis type, and prompt
func generateLLMCacheKey(fileHash string, analysisType string, prompt string) string {
	input := fmt.Sprintf("%s:%s:%s", fileHash, analysisType, prompt)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// calculateFileHash creates a hash of file content for caching
func calculateFileHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
