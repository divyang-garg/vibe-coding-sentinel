// LLM Cache - Enhanced Caching Functions
// Implements enhanced caching for comprehensive analysis and business context
// Phase 14D: Enhanced caching with comprehensive analysis and business context caching
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ComprehensiveAnalysisCache represents a cached comprehensive analysis result
type ComprehensiveAnalysisCache struct {
	Result    *ComprehensiveAnalysisReport
	CreatedAt time.Time
	ExpiresAt time.Time
}

// BusinessContextCache represents cached business context (rules, entities, journeys)
type BusinessContextCache struct {
	Rules     []interface{} // BusinessRule type from knowledge items
	Entities  []interface{} // Entity type from knowledge items
	Journeys  []interface{} // Journey type from knowledge items
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Cache metrics tracking
var (
	analysisResultCache  = sync.Map{} // map[string]*ComprehensiveAnalysisCache
	businessContextCache = sync.Map{} // map[string]*BusinessContextCache
	cacheHitCounter      = sync.Map{} // map[string]int64 (projectID -> hits)
	cacheMissCounter     = sync.Map{} // map[string]int64 (projectID -> misses)
	cacheMetricsMutex    sync.RWMutex
)

// Phase 14D: Model selection savings tracking
var (
	modelSelectionSavingsCounter  = sync.Map{} // map[string]float64 (projectID -> total savings)
	cheaperModelSelectedCounter   = sync.Map{} // map[string]int64 (projectID -> count)
	expensiveModelSelectedCounter = sync.Map{} // map[string]int64 (projectID -> count)
	cacheSizeCounter              = sync.Map{} // map[string]int64 (projectID -> cache size)
)

// generateFeatureHash generates a hash from feature name and codebase path
func generateFeatureHash(feature, codebasePath string) string {
	input := fmt.Sprintf("%s:%s", feature, codebasePath)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// getCachedAnalysisResult retrieves a cached comprehensive analysis result
func getCachedAnalysisResult(projectID, featureHash, depth, mode string, config *LLMConfig) (*ComprehensiveAnalysisReport, bool) {
	// Check if caching is enabled
	if config != nil && !config.CostOptimization.UseCache {
		recordCacheMiss(projectID)
		return nil, false
	}

	cacheKey := fmt.Sprintf("analysis:%s:%s:%s:%s", projectID, featureHash, depth, mode)
	cached, ok := analysisResultCache.Load(cacheKey)
	if !ok {
		recordCacheMiss(projectID)
		return nil, false
	}

	entry := cached.(*ComprehensiveAnalysisCache)
	if time.Now().After(entry.ExpiresAt) {
		// Entry expired, remove it
		// Phase 14D: Decrement cache size counter
		if val, ok := cacheSizeCounter.Load(projectID); ok {
			size := val.(int64)
			if size > 0 {
				cacheSizeCounter.Store(projectID, size-1)
			}
		}
		analysisResultCache.Delete(cacheKey)
		recordCacheMiss(projectID)
		return nil, false
	}

	recordCacheHit(projectID)
	return entry.Result, true
}

// setCachedAnalysisResult stores a comprehensive analysis result in cache
func setCachedAnalysisResult(projectID, featureHash, depth, mode string, result *ComprehensiveAnalysisReport, config *LLMConfig) {
	if config != nil && !config.CostOptimization.UseCache {
		return
	}

	cacheKey := fmt.Sprintf("analysis:%s:%s:%s:%s", projectID, featureHash, depth, mode)

	// Phase 14D: Check if entry already exists to avoid double counting
	_, exists := analysisResultCache.Load(cacheKey)

	ttl := 24 * time.Hour // default
	if config != nil && config.CostOptimization.CacheTTLHours > 0 {
		ttl = time.Duration(config.CostOptimization.CacheTTLHours) * time.Hour
	}

	analysisResultCache.Store(cacheKey, &ComprehensiveAnalysisCache{
		Result:    result,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	})

	// Phase 14D: Increment cache size counter if new entry
	if !exists {
		if val, ok := cacheSizeCounter.Load(projectID); ok {
			size := val.(int64)
			cacheSizeCounter.Store(projectID, size+1)
		} else {
			cacheSizeCounter.Store(projectID, int64(1))
		}
	}
}

// getCachedBusinessContext retrieves cached business context
func getCachedBusinessContext(projectID, codebaseHash string, config *LLMConfig) (map[string]interface{}, bool) {
	if config != nil && !config.CostOptimization.UseCache {
		return nil, false
	}

	cacheKey := fmt.Sprintf("business:%s:%s", projectID, codebaseHash)
	cached, ok := businessContextCache.Load(cacheKey)
	if !ok {
		return nil, false
	}

	entry := cached.(*BusinessContextCache)
	if time.Now().After(entry.ExpiresAt) {
		// Phase 14D: Decrement cache size counter
		if val, ok := cacheSizeCounter.Load(projectID); ok {
			size := val.(int64)
			if size > 0 {
				cacheSizeCounter.Store(projectID, size-1)
			}
		}
		businessContextCache.Delete(cacheKey)
		return nil, false
	}

	result := map[string]interface{}{
		"rules":    entry.Rules,
		"entities": entry.Entities,
		"journeys": entry.Journeys,
	}
	return result, true
}

// setCachedBusinessContext stores business context in cache
func setCachedBusinessContext(projectID, codebaseHash string, rules, entities, journeys []interface{}, config *LLMConfig) {
	if config != nil && !config.CostOptimization.UseCache {
		return
	}

	cacheKey := fmt.Sprintf("business:%s:%s", projectID, codebaseHash)

	// Phase 14D: Check if entry already exists to avoid double counting
	_, exists := businessContextCache.Load(cacheKey)

	ttl := 24 * time.Hour // default
	if config != nil && config.CostOptimization.CacheTTLHours > 0 {
		ttl = time.Duration(config.CostOptimization.CacheTTLHours) * time.Hour
	}

	businessContextCache.Store(cacheKey, &BusinessContextCache{
		Rules:     rules,
		Entities:  entities,
		Journeys:  journeys,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	})

	// Phase 14D: Increment cache size counter if new entry
	if !exists {
		if val, ok := cacheSizeCounter.Load(projectID); ok {
			size := val.(int64)
			cacheSizeCounter.Store(projectID, size+1)
		} else {
			cacheSizeCounter.Store(projectID, int64(1))
		}
	}
}

// recordCacheHit records a cache hit for metrics
func recordCacheHit(projectID string) {
	cacheMetricsMutex.Lock()
	defer cacheMetricsMutex.Unlock()

	var hits int64
	if val, ok := cacheHitCounter.Load(projectID); ok {
		hits = val.(int64)
	}
	cacheHitCounter.Store(projectID, hits+1)
}

// recordCacheMiss records a cache miss for metrics
func recordCacheMiss(projectID string) {
	cacheMetricsMutex.Lock()
	defer cacheMetricsMutex.Unlock()

	var misses int64
	if val, ok := cacheMissCounter.Load(projectID); ok {
		misses = val.(int64)
	}
	cacheMissCounter.Store(projectID, misses+1)
}

// getCacheHitRate calculates cache hit rate for a project
func getCacheHitRate(projectID string) float64 {
	cacheMetricsMutex.RLock()
	defer cacheMetricsMutex.RUnlock()

	var hits, misses int64
	if val, ok := cacheHitCounter.Load(projectID); ok {
		hits = val.(int64)
	}
	if val, ok := cacheMissCounter.Load(projectID); ok {
		misses = val.(int64)
	}

	total := hits + misses
	if total == 0 {
		return 0.0
	}

	return float64(hits) / float64(total)
}

// trackModelSelectionSavings tracks savings from model selection decisions
// Phase 14D: New function to track actual cost savings
func trackModelSelectionSavings(projectID string, savings float64, isCheaperModel bool) {
	if savings <= 0 {
		return
	}

	// Track total savings
	if val, ok := modelSelectionSavingsCounter.Load(projectID); ok {
		currentSavings := val.(float64)
		modelSelectionSavingsCounter.Store(projectID, currentSavings+savings)
	} else {
		modelSelectionSavingsCounter.Store(projectID, savings)
	}

	// Track model selection counts
	if isCheaperModel {
		if val, ok := cheaperModelSelectedCounter.Load(projectID); ok {
			count := val.(int64)
			cheaperModelSelectedCounter.Store(projectID, count+1)
		} else {
			cheaperModelSelectedCounter.Store(projectID, int64(1))
		}
	} else {
		if val, ok := expensiveModelSelectedCounter.Load(projectID); ok {
			count := val.(int64)
			expensiveModelSelectedCounter.Store(projectID, count+1)
		} else {
			expensiveModelSelectedCounter.Store(projectID, int64(1))
		}
	}
}

// getModelSelectionSavings returns the total savings from model selection for a project
// Phase 14D: New function to retrieve tracked savings
func getModelSelectionSavings(projectID string) float64 {
	if val, ok := modelSelectionSavingsCounter.Load(projectID); ok {
		return val.(float64)
	}
	return 0.0
}

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
					if val, ok := cacheSizeCounter.Load(projectID); ok {
						size := val.(int64)
						if size > 0 {
							cacheSizeCounter.Store(projectID, size-1)
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
					if val, ok := cacheSizeCounter.Load(projectID); ok {
						size := val.(int64)
						if size > 0 {
							cacheSizeCounter.Store(projectID, size-1)
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
