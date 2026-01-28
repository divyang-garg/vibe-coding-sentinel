// Package main - LLM cache cleanup operations
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"strings"
	"time"

	"sentinel-hub-api/pkg"
)

// cleanupAnalysisCache removes expired analysis cache entries
func cleanupAnalysisCache() {
	now := time.Now()
	analysisResultCache.Range(func(key, value interface{}) bool {
		entry := value.(*ComprehensiveAnalysisCache)
		if now.After(entry.ExpiresAt) {
			// Phase 14D: Extract projectID from cache key and decrement counter
			if keyStr, ok := key.(string); ok {
				// Key format: "analysis:{projectID}:{featureHash}:{depth}:{mode}"
				parts := strings.Split(keyStr, ":")
				if len(parts) >= 2 && parts[0] == "analysis" {
					projectID := parts[1]
					if val, ok := pkg.CacheSizeCounter.Load(projectID); ok {
						size := val.(int64)
						if size > 0 {
							pkg.CacheSizeCounter.Store(projectID, size-1)
						}
					}
				}
			}
			analysisResultCache.Delete(key)
		}
		return true
	})
}

// cleanupBusinessContextCache removes expired business context cache entries
func cleanupBusinessContextCache() {
	now := time.Now()
	businessContextCache.Range(func(key, value interface{}) bool {
		entry := value.(*BusinessContextCache)
		if now.After(entry.ExpiresAt) {
			// Phase 14D: Extract projectID from cache key and decrement counter
			if keyStr, ok := key.(string); ok {
				// Key format: "business:{projectID}:{codebaseHash}"
				parts := strings.Split(keyStr, ":")
				if len(parts) >= 2 && parts[0] == "business" {
					projectID := parts[1]
					if val, ok := pkg.CacheSizeCounter.Load(projectID); ok {
						size := val.(int64)
						if size > 0 {
							pkg.CacheSizeCounter.Store(projectID, size-1)
						}
					}
				}
			}
			businessContextCache.Delete(key)
		}
		return true
	})
}

// startCacheCleanup starts background goroutine for cache cleanup
func startCacheCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupLLMCache()
			cleanupAnalysisCache()
			cleanupBusinessContextCache()
		}
	}()
}
