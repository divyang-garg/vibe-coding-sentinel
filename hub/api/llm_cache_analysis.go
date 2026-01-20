// Package main - LLM analysis result caching
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ComprehensiveAnalysisCache represents a cached comprehensive analysis result
type ComprehensiveAnalysisCache struct {
	Result    *ComprehensiveAnalysisReport
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Cache metrics tracking
var (
	analysisResultCache = sync.Map{} // map[string]*ComprehensiveAnalysisCache
)

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
