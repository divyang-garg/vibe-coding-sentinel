// Package main - LLM business context caching
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"fmt"
	"sync"
	"time"

	"sentinel-hub-api/pkg"
)

// BusinessContextCache represents cached business context (rules, entities, journeys)
type BusinessContextCache struct {
	Rules     []interface{} // BusinessRule type from knowledge items
	Entities  []interface{} // Entity type from knowledge items
	Journeys  []interface{} // Journey type from knowledge items
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Cache for business context
var (
	businessContextCache = sync.Map{} // map[string]*BusinessContextCache
)

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
		if val, ok := pkg.CacheSizeCounter.Load(projectID); ok {
			size := val.(int64)
			if size > 0 {
				pkg.CacheSizeCounter.Store(projectID, size-1)
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
		if val, ok := pkg.CacheSizeCounter.Load(projectID); ok {
			size := val.(int64)
			pkg.CacheSizeCounter.Store(projectID, size+1)
		} else {
			pkg.CacheSizeCounter.Store(projectID, int64(1))
		}
	}
}
