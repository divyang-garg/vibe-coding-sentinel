// Package main - LLM cache metrics tracking
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"sync"
)

// Cache metrics tracking
var (
	cacheHitCounter      = sync.Map{} // map[string]int64 (projectID -> hits)
	cacheMissCounter     = sync.Map{} // map[string]int64 (projectID -> misses)
	cacheMetricsMutex   sync.RWMutex
	modelSelectionSavingsCounter  = sync.Map{} // map[string]float64 (projectID -> total savings)
	cheaperModelSelectedCounter   = sync.Map{} // map[string]int64 (projectID -> count)
	expensiveModelSelectedCounter = sync.Map{} // map[string]int64 (projectID -> count)
	cacheSizeCounter              = sync.Map{} // map[string]int64 (projectID -> cache size)
)

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
