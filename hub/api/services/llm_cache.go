// Phase 14A: LLM Cost Optimization
// Implements caching and progressive depth for LLM calls
// Phase 14D: Enhanced caching with comprehensive analysis and business context caching

package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
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

	// Use CacheTTLHours from config if available, otherwise use default
	ttl := llmCacheTTL // default 24h
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
	// Combine all inputs
	input := fmt.Sprintf("%s:%s:%s", fileHash, analysisType, prompt)

	// Hash the input
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// analyzeWithProgressiveDepth performs analysis with progressive depth levels
func analyzeWithProgressiveDepth(ctx context.Context, config *LLMConfig, fileContent string, analysisType string, depth string, projectID string, validationID string) (string, error) {
	// Level 1: Fast checks (pattern-based, AST-based) - No LLM
	if depth == "surface" {
		// Use pattern matching and AST analysis only - no LLM required
		return "", nil
	}

	// Level 2: Medium-depth (cheaper models) - Only if Level 1 finds issues
	if depth == "medium" {
		// Check cache first
		fileHash := calculateFileHash(fileContent)
		cacheKey := generateLLMCacheKey(fileHash, analysisType, "medium")

		if cached, ok := getCachedLLMResponse(cacheKey, config); ok {
			return cached, nil
		}

		// Phase 14D: Select model based on task type and depth (respects user config and cost limits)
		// Estimate tokens (rough: 1 token ≈ 4 characters)
		estimatedTokens := len(fileContent) / 4
		selectedModel, err := selectModelWithDepth(ctx, analysisType, config, depth, estimatedTokens, projectID)
		if err != nil {
			return "", fmt.Errorf("failed to select model: %w", err)
		}

		// Use selected model (may be cheaper model for medium depth)
		originalModel := config.Model
		config.Model = selectedModel

		prompt := generatePrompt(analysisType, depth, fileContent)
		response, tokensUsed, err := callLLMWithDepth(ctx, config, prompt, analysisType, depth, projectID)
		if err != nil {
			config.Model = originalModel // Restore original
			return "", err
		}

		// Cache the response with actual token count
		setCachedLLMResponse(cacheKey, response, tokensUsed, config)

		// Track usage
		usage := &LLMUsage{
			ProjectID:     projectID,
			ValidationID:  validationID,
			Provider:      config.Provider,
			Model:         config.Model,
			TokensUsed:    tokensUsed,
			EstimatedCost: calculateEstimatedCost(config.Provider, config.Model, tokensUsed),
		}
		if err := trackUsage(ctx, usage); err != nil {
			// Log but don't fail
			LogWarn(ctx, "Failed to track LLM usage: %v", err)
		}

		config.Model = originalModel // Restore original
		return response, nil
	}

	// Level 3: Deep analysis (expensive models) - Only if Level 2 finds issues
	if depth == "deep" {
		// Check cache first
		fileHash := calculateFileHash(fileContent)
		cacheKey := generateLLMCacheKey(fileHash, analysisType, "deep")

		if cached, ok := getCachedLLMResponse(cacheKey, config); ok {
			return cached, nil
		}

		// Phase 14D: Select model based on task type and depth (respects user config and cost limits)
		// Estimate tokens (rough: 1 token ≈ 4 characters)
		estimatedTokens := len(fileContent) / 4
		selectedModel, err := selectModelWithDepth(ctx, analysisType, config, depth, estimatedTokens, projectID)
		if err != nil {
			return "", fmt.Errorf("failed to select model: %w", err)
		}

		// Use selected model (may be high-accuracy model for deep depth)
		originalModel := config.Model
		config.Model = selectedModel

		prompt := generatePrompt(analysisType, depth, fileContent)
		response, tokensUsed, err := callLLMWithDepth(ctx, config, prompt, analysisType, depth, projectID)
		if err != nil {
			config.Model = originalModel // Restore original
			return "", err
		}

		// Cache the response with actual token count
		setCachedLLMResponse(cacheKey, response, tokensUsed, config)

		// Track usage
		usage := &LLMUsage{
			ProjectID:     projectID,
			ValidationID:  validationID,
			Provider:      config.Provider,
			Model:         config.Model,
			TokensUsed:    tokensUsed,
			EstimatedCost: calculateEstimatedCost(config.Provider, config.Model, tokensUsed),
		}
		if err := trackUsage(ctx, usage); err != nil {
			// Log but don't fail
			LogWarn(ctx, "Failed to track LLM usage: %v", err)
		}

		config.Model = originalModel // Restore original
		return response, nil
	}

	return "", fmt.Errorf("invalid depth: %s", depth)
}

// Helper functions

func calculateFileHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func generatePrompt(analysisType string, depth string, fileContent string) string {
	// Generate structured prompt based on analysis type and depth
	var systemPrompt, userPrompt string

	switch analysisType {
	case "semantic_analysis":
		systemPrompt = "You are an expert code analyzer specializing in semantic analysis. Analyze code for logic errors, edge cases, and potential bugs."
		userPrompt = fmt.Sprintf(`Analyze the following code with %s depth for semantic issues:

%s

Provide your analysis in JSON format with the following structure:
{
  "issues": [
    {
      "type": "error_type",
      "line": "line_number",
      "description": "detailed description",
      "severity": "low|medium|high"
    }
  ]
}`, depth, fileContent)
	case "business_logic":
		systemPrompt = "You are an expert business logic analyzer. Analyze code for business rule compliance and logic correctness."
		userPrompt = fmt.Sprintf(`Analyze the following business logic code with %s depth:

%s

Identify any violations of business rules, missing validations, or incorrect logic flows.`, depth, fileContent)
	case "error_handling":
		systemPrompt = "You are an expert in error handling analysis. Analyze code for proper error handling patterns."
		userPrompt = fmt.Sprintf(`Analyze the following code for error handling with %s depth:

%s

Identify missing error handling, improper error propagation, or error handling anti-patterns.`, depth, fileContent)
	default:
		systemPrompt = "You are an expert code analyzer."
		userPrompt = fmt.Sprintf(`Analyze the following code for %s with %s depth:

%s`, analysisType, depth, fileContent)
	}

	// Combine into structured prompt
	return fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
}

// =============================================================================
// Phase 14D: Enhanced Caching Functions
// =============================================================================

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
