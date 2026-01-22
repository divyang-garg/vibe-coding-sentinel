// LLM Cache - Progressive Depth Analysis
//
// This package implements progressive depth analysis for general code quality checks.
// Unlike the main package version, this focuses on standard code analysis workflows
// without complex cost optimization features.
//
// Depth Levels:
//   - quick: Brief summary, 3-5 findings - Fast feedback
//   - medium: Moderate analysis with examples - Standard reviews
//   - deep: Comprehensive with all details - Thorough analysis
//
// Analysis Types Supported:
//   - security: Security vulnerabilities and best practices
//   - performance: Performance issues and optimization opportunities
//   - maintainability: Code quality and maintainability metrics
//   - architecture: Design patterns and architectural concerns
//
// Differences from Main Package:
//   - Simpler implementation without dynamic model selection
//   - Uses configured model (no cost optimization)
//   - Focuses on code quality analysis rather than cost optimization
//
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
)

// analyzeWithProgressiveDepth performs analysis with progressive depth levels
func analyzeWithProgressiveDepth(ctx context.Context, config *LLMConfig, fileContent string, analysisType string, depth string, projectID string, validationID string) (string, error) {
	// Generate cache key
	fileHash := calculateFileHash(fileContent)
	cacheKey := generateLLMCacheKey(fileHash, analysisType, depth)

	// Check cache first
	if cached, ok := getCachedLLMResponse(cacheKey, config); ok {
		LogInfo(ctx, "Cache hit for analysis: %s (depth: %s)", analysisType, depth)
		return cached, nil
	}

	// Generate prompt based on depth
	prompt := generatePrompt(analysisType, depth, fileContent)

	// Call LLM
	response, tokensUsed, err := callLLM(ctx, config, prompt, "progressive_analysis")
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	// Cache the response
	setCachedLLMResponse(cacheKey, response, tokensUsed, config)

	return response, nil
}

// generatePrompt generates a prompt based on analysis type and depth
// This is a wrapper that delegates to the unified prompt builder
func generatePrompt(analysisType string, depth string, fileContent string) string {
	return GeneratePrompt(analysisType, depth, fileContent)
}
