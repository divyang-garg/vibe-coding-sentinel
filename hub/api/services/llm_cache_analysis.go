// LLM Cache - Progressive Depth Analysis
// Implements progressive depth analysis for LLM calls
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
func generatePrompt(analysisType string, depth string, fileContent string) string {
	var systemPrompt string
	var userPrompt string

	// System prompt based on analysis type
	switch analysisType {
	case "security":
		systemPrompt = "You are a security analysis expert. Analyze code for security vulnerabilities and best practices."
	case "performance":
		systemPrompt = "You are a performance optimization expert. Analyze code for performance issues and optimization opportunities."
	case "maintainability":
		systemPrompt = "You are a code quality expert. Analyze code for maintainability, readability, and best practices."
	case "architecture":
		systemPrompt = "You are an architecture expert. Analyze code structure, design patterns, and architectural concerns."
	default:
		systemPrompt = "You are a code analysis expert. Analyze the provided code."
	}

	// User prompt based on depth
	switch depth {
	case "quick":
		userPrompt = fmt.Sprintf("Quick analysis: Provide a brief summary of key findings in the following code:\n\n%s", fileContent)
	case "medium":
		userPrompt = fmt.Sprintf("Medium analysis: Analyze the code and provide findings with examples. Focus on the most important issues:\n\n%s", fileContent)
	case "deep":
		userPrompt = fmt.Sprintf("Deep analysis: Perform comprehensive analysis of the code. Include detailed findings, recommendations, and examples:\n\n%s", fileContent)
	default:
		userPrompt = fmt.Sprintf("Analyze the following code:\n\n%s", fileContent)
	}

	// Combine into structured prompt
	return fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
}
