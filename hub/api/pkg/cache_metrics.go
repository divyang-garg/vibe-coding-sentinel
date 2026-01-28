// Package pkg - Cache metrics tracking
// Shared package for cache metrics that can be used by both main and services
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package pkg

import (
	"sync"
)

// Cache metrics tracking - shared across packages
var (
	CacheHitCounter               = sync.Map{} // map[string]int64 (projectID -> hits)
	CacheMissCounter              = sync.Map{} // map[string]int64 (projectID -> misses)
	CacheMetricsMutex             sync.RWMutex
	ModelSelectionSavingsCounter  = sync.Map{} // map[string]float64 (projectID -> total savings)
	CheaperModelSelectedCounter   = sync.Map{} // map[string]int64 (projectID -> count)
	ExpensiveModelSelectedCounter = sync.Map{} // map[string]int64 (projectID -> count)
	CacheSizeCounter              = sync.Map{} // map[string]int64 (projectID -> cache size)
)

// RecordCacheHit records a cache hit for metrics
func RecordCacheHit(projectID string) {
	CacheMetricsMutex.Lock()
	defer CacheMetricsMutex.Unlock()

	var hits int64
	if val, ok := CacheHitCounter.Load(projectID); ok {
		hits = val.(int64)
	}
	CacheHitCounter.Store(projectID, hits+1)
}

// RecordCacheMiss records a cache miss for metrics
func RecordCacheMiss(projectID string) {
	CacheMetricsMutex.Lock()
	defer CacheMetricsMutex.Unlock()

	var misses int64
	if val, ok := CacheMissCounter.Load(projectID); ok {
		misses = val.(int64)
	}
	CacheMissCounter.Store(projectID, misses+1)
}

// GetCacheHitRate calculates cache hit rate for a project
func GetCacheHitRate(projectID string) float64 {
	CacheMetricsMutex.RLock()
	defer CacheMetricsMutex.RUnlock()

	var hits, misses int64
	if val, ok := CacheHitCounter.Load(projectID); ok {
		hits = val.(int64)
	}
	if val, ok := CacheMissCounter.Load(projectID); ok {
		misses = val.(int64)
	}

	total := hits + misses
	if total == 0 {
		return 0.0
	}

	return float64(hits) / float64(total)
}

// GetCacheHits returns cache hits count for a project
func GetCacheHits(projectID string) int64 {
	CacheMetricsMutex.RLock()
	defer CacheMetricsMutex.RUnlock()

	if val, ok := CacheHitCounter.Load(projectID); ok {
		return val.(int64)
	}
	return 0
}

// GetCacheMisses returns cache misses count for a project
func GetCacheMisses(projectID string) int64 {
	CacheMetricsMutex.RLock()
	defer CacheMetricsMutex.RUnlock()

	if val, ok := CacheMissCounter.Load(projectID); ok {
		return val.(int64)
	}
	return 0
}

// GetCacheSize returns cache size for a project
func GetCacheSize(projectID string) int64 {
	CacheMetricsMutex.RLock()
	defer CacheMetricsMutex.RUnlock()

	if val, ok := CacheSizeCounter.Load(projectID); ok {
		return val.(int64)
	}
	return 0
}

// TrackModelSelectionSavings tracks savings from model selection decisions
func TrackModelSelectionSavings(projectID string, savings float64, isCheaperModel bool) {
	if savings <= 0 {
		return
	}

	// Track total savings
	if val, ok := ModelSelectionSavingsCounter.Load(projectID); ok {
		currentSavings := val.(float64)
		ModelSelectionSavingsCounter.Store(projectID, currentSavings+savings)
	} else {
		ModelSelectionSavingsCounter.Store(projectID, savings)
	}

	// Track model selection counts
	if isCheaperModel {
		if val, ok := CheaperModelSelectedCounter.Load(projectID); ok {
			count := val.(int64)
			CheaperModelSelectedCounter.Store(projectID, count+1)
		} else {
			CheaperModelSelectedCounter.Store(projectID, int64(1))
		}
	} else {
		if val, ok := ExpensiveModelSelectedCounter.Load(projectID); ok {
			count := val.(int64)
			ExpensiveModelSelectedCounter.Store(projectID, count+1)
		} else {
			ExpensiveModelSelectedCounter.Store(projectID, int64(1))
		}
	}
}

// GetModelSelectionSavings returns the total savings from model selection for a project
func GetModelSelectionSavings(projectID string) float64 {
	if val, ok := ModelSelectionSavingsCounter.Load(projectID); ok {
		return val.(float64)
	}
	return 0.0
}

// GetCheaperModelCount returns count of cheaper model selections
func GetCheaperModelCount(projectID string) int64 {
	if val, ok := CheaperModelSelectedCounter.Load(projectID); ok {
		return val.(int64)
	}
	return 0
}

// GetExpensiveModelCount returns count of expensive model selections
func GetExpensiveModelCount(projectID string) int64 {
	if val, ok := ExpensiveModelSelectedCounter.Load(projectID); ok {
		return val.(int64)
	}
	return 0
}
