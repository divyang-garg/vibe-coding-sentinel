// Package llm provides advanced cost optimization algorithms
// Complies with CODING_STANDARDS.md: Cost optimization max 250 lines
package llm

import (
	"math"
)

// CostOptimizer handles intelligent model selection based on cost constraints
type CostOptimizer struct {
	ProviderRates map[string]map[string]float64 // provider -> model -> cost per 1K tokens
	BudgetLimits  map[string]float64            // provider -> max cost per request
}

// NewCostOptimizer creates a cost optimizer with current pricing
func NewCostOptimizer() *CostOptimizer {
	return &CostOptimizer{
		ProviderRates: map[string]map[string]float64{
			"openai": {
				"gpt-4":         0.03,  // --.03 per 1K tokens
				"gpt-3.5-turbo": 0.002, // --.002 per 1K tokens
			},
			"anthropic": {
				"claude-3-opus":   0.015,   // --.015 per 1K tokens
				"claude-3-sonnet": 0.003,   // --.003 per 1K tokens
				"claude-3-haiku":  0.00025, // --.00025 per 1K tokens
			},
			"azure": {
				"gpt-4":         0.03,
				"gpt-3.5-turbo": 0.002,
			},
		},
		BudgetLimits: make(map[string]float64),
	}
}

// OptimizeModelSelection selects the best model within cost constraints
func (co *CostOptimizer) OptimizeModelSelection(taskType string, provider string, estimatedTokens int, costConfig *CostOptimizationConfig) string {
	// Extract cost constraints
	maxCostPerRequest := 0.10 // default --.10
	if costConfig != nil && costConfig.MaxCostPerRequest > 0 {
		maxCostPerRequest = costConfig.MaxCostPerRequest
	}

	// Get available models for provider
	models, exists := co.ProviderRates[provider]
	if !exists {
		return "gpt-3.5-turbo" // fallback
	}

	// Calculate cost for each model
	bestModel := ""
	lowestCost := math.MaxFloat64

	for model, ratePer1K := range models {
		estimatedCost := (float64(estimatedTokens) / 1000.0) * ratePer1K

		// Check if within budget
		if estimatedCost <= maxCostPerRequest && estimatedCost < lowestCost {
			// Additional logic based on task type
			if co.isModelSuitableForTask(model, taskType) {
				lowestCost = estimatedCost
				bestModel = model
			}
		}
	}

	if bestModel == "" {
		// No model within budget, return cheapest available
		for model, ratePer1K := range models {
			estimatedCost := (float64(estimatedTokens) / 1000.0) * ratePer1K
			if estimatedCost < lowestCost {
				lowestCost = estimatedCost
				bestModel = model
			}
		}
	}

	return bestModel
}

// isModelSuitableForTask checks if a model is appropriate for the task
func (co *CostOptimizer) isModelSuitableForTask(model string, taskType string) bool {
	switch taskType {
	case "code_analysis", "security_audit", "performance_analysis":
		// Require capable models for complex analysis
		return model == "gpt-4" || model == "claude-3-opus" || model == "claude-3-sonnet"
	case "documentation", "explanation":
		// Balanced models acceptable
		return true
	case "quick_analysis", "syntax_check":
		// Any model acceptable for simple tasks
		return true
	default:
		return true
	}
}

// EstimateTokens provides a rough token count estimation
func EstimateTokens(text string) int {
	// Rough estimation: 1 token ≈ 4 characters for English text
	// More accurate for code: ~1 token per 3-4 characters
	charCount := len(text)
	if charCount == 0 {
		return 0
	}

	// Code has higher token density than natural language
	// Estimate based on common patterns
	tokenEstimate := int(math.Ceil(float64(charCount) / 3.5))
	return tokenEstimate
}
