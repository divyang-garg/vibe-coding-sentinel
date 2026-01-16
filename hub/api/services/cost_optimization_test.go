// Phase 14D: Cost Optimization Tests
// These tests are in the main package to access unexported functions

package services

import (
	"context"
	"fmt"
	"testing"
)

// TestGetCacheHitRate tests cache hit rate calculation
func TestGetCacheHitRate(t *testing.T) {
	projectID := "test-project-123"

	// Reset counters
	cacheHitCounter.Delete(projectID)
	cacheMissCounter.Delete(projectID)

	// Initially should be 0
	hitRate := getCacheHitRate(projectID)
	if hitRate != 0.0 {
		t.Errorf("Expected hit rate 0.0, got %f", hitRate)
	}

	// Record some hits and misses
	recordCacheHit(projectID)
	recordCacheHit(projectID)
	recordCacheMiss(projectID)

	hitRate = getCacheHitRate(projectID)
	expectedRate := 2.0 / 3.0 // 2 hits out of 3 total
	if hitRate != expectedRate {
		t.Errorf("Expected hit rate %f, got %f", expectedRate, hitRate)
	}
}

// TestGetModelCost tests model cost retrieval
func TestGetModelCost(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		expected float64
	}{
		{"openai", "gpt-4", 0.03},
		{"openai", "gpt-3.5-turbo", 0.0015},
		{"anthropic", "claude-3-haiku", 0.00025},
		{"unknown", "unknown-model", 0.01}, // default fallback
	}

	for _, tt := range tests {
		cost := getModelCost(tt.provider, tt.model)
		if cost != tt.expected {
			t.Errorf("getModelCost(%s, %s) = %f, want %f", tt.provider, tt.model, cost, tt.expected)
		}
	}
}

// TestEstimateCost tests cost estimation
func TestEstimateCost(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		tokens   int
		expected float64
	}{
		{"openai", "gpt-4", 1000, 0.03},
		{"openai", "gpt-3.5-turbo", 2000, 0.003},
		{"anthropic", "claude-3-haiku", 1000, 0.00025},
	}

	for _, tt := range tests {
		cost := estimateCost(tt.provider, tt.model, tt.tokens)
		if cost != tt.expected {
			t.Errorf("estimateCost(%s, %s, %d) = %f, want %f", tt.provider, tt.model, tt.tokens, cost, tt.expected)
		}
	}
}

// TestSelectCheaperModel tests cheaper model selection
func TestSelectCheaperModel(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{"openai", "gpt-3.5-turbo"},
		{"anthropic", "claude-3-haiku"},
		{"azure", "gpt-3.5-turbo"},
		{"unknown", "gpt-3.5-turbo"}, // default fallback
	}

	for _, tt := range tests {
		model, err := selectCheaperModel(tt.provider)
		if err != nil {
			t.Errorf("selectCheaperModel(%s) returned error: %v", tt.provider, err)
		}
		if model != tt.expected {
			t.Errorf("selectCheaperModel(%s) = %s, want %s", tt.provider, model, tt.expected)
		}
	}
}

// TestSelectExpensiveModel tests expensive model selection
func TestSelectExpensiveModel(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{"openai", "gpt-4"},
		{"anthropic", "claude-3-opus"},
		{"azure", "gpt-4"},
		{"unknown", "gpt-4"}, // default fallback
	}

	for _, tt := range tests {
		model, err := selectExpensiveModel(tt.provider)
		if err != nil {
			t.Errorf("selectExpensiveModel(%s) returned error: %v", tt.provider, err)
		}
		if model != tt.expected {
			t.Errorf("selectExpensiveModel(%s) = %s, want %s", tt.provider, model, tt.expected)
		}
	}
}

// TestSelectModelWithDepth tests model selection with depth consideration
func TestSelectModelWithDepth(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		taskType        string
		config          *LLMConfig
		depth           string
		estimatedTokens int
		expectCheaper   bool
		expectExpensive bool
	}{
		{
			name:            "medium depth non-critical should prefer cheaper",
			taskType:        "code_review",
			config:          &LLMConfig{Provider: "openai", CostOptimization: CostOptimizationConfig{MaxCostPerRequest: 0}},
			depth:           "medium",
			estimatedTokens: 1000,
			expectCheaper:   true,
		},
		{
			name:            "deep depth critical should prefer expensive",
			taskType:        "security_analysis",
			config:          &LLMConfig{Provider: "openai", CostOptimization: CostOptimizationConfig{MaxCostPerRequest: 0}},
			depth:           "deep",
			estimatedTokens: 1000,
			expectExpensive: true,
		},
		{
			name:            "cost limit exceeded should use cheaper",
			taskType:        "semantic_analysis",
			config:          &LLMConfig{Provider: "openai", Model: "gpt-4", CostOptimization: CostOptimizationConfig{MaxCostPerRequest: 0.01}},
			depth:           "deep",
			estimatedTokens: 10000, // Would cost $0.30 with gpt-4, exceeds $0.01 limit
			expectCheaper:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := selectModelWithDepth(ctx, tt.taskType, tt.config, tt.depth, tt.estimatedTokens, "test-project")
			if err != nil {
				t.Errorf("selectModelWithDepth() returned error: %v", err)
				return
			}

			if tt.expectCheaper && model != "gpt-3.5-turbo" {
				t.Errorf("Expected cheaper model (gpt-3.5-turbo), got %s", model)
			}
			if tt.expectExpensive && model != "gpt-4" {
				t.Errorf("Expected expensive model (gpt-4), got %s", model)
			}
		})
	}
}

// TestCacheRespectsUseCacheFlag tests that cache respects UseCache config flag
func TestCacheRespectsUseCacheFlag(t *testing.T) {
	projectID := "test-project"
	featureHash := "test-hash"
	depth := "medium"
	mode := "auto"

	// Create config with caching disabled
	config := &LLMConfig{
		CostOptimization: CostOptimizationConfig{
			UseCache: false,
		},
	}

	// Try to get cached result - should return false
	result, ok := getCachedAnalysisResult(projectID, featureHash, depth, mode, config)
	if ok {
		t.Error("Expected cache miss when UseCache is false, got cache hit")
	}
	if result != nil {
		t.Error("Expected nil result when UseCache is false")
	}

	// Try to set cached result - should not cache
	report := &ComprehensiveAnalysisReport{
		ValidationID: "test-validation",
		Feature:      "test-feature",
	}
	setCachedAnalysisResult(projectID, featureHash, depth, mode, report, config)

	// Verify it wasn't cached
	_, ok = getCachedAnalysisResult(projectID, featureHash, depth, mode, config)
	if ok {
		t.Error("Expected cache miss after setting with UseCache=false, got cache hit")
	}
}

// TestCacheTTLHours tests that cache respects CacheTTLHours config
func TestCacheTTLHours(t *testing.T) {
	projectID := "test-project"
	featureHash := "test-hash"
	depth := "medium"
	mode := "auto"

	// Create config with custom TTL
	customTTL := 12 // 12 hours
	config := &LLMConfig{
		CostOptimization: CostOptimizationConfig{
			UseCache:      true,
			CacheTTLHours: customTTL,
		},
	}

	report := &ComprehensiveAnalysisReport{
		ValidationID: "test-validation",
		Feature:      "test-feature",
	}

	setCachedAnalysisResult(projectID, featureHash, depth, mode, report, config)

	// Verify it was cached
	result, ok := getCachedAnalysisResult(projectID, featureHash, depth, mode, config)
	if !ok {
		t.Error("Expected cache hit after setting, got cache miss")
	}
	if result == nil || result.ValidationID != "test-validation" {
		t.Error("Cached result doesn't match")
	}

	// Note: Testing expiration would require manipulating time, which is more complex
	// In production, this would be tested with a time mock
}

// TestCacheMetricsTracking tests cache metrics tracking
func TestCacheMetricsTracking(t *testing.T) {
	projectID := "test-project-metrics"

	// Reset counters
	cacheHitCounter.Delete(projectID)
	cacheMissCounter.Delete(projectID)

	// Record some hits and misses
	recordCacheHit(projectID)
	recordCacheHit(projectID)
	recordCacheMiss(projectID)
	recordCacheHit(projectID)

	hitRate := getCacheHitRate(projectID)
	expectedRate := 3.0 / 4.0 // 3 hits out of 4 total
	if hitRate != expectedRate {
		t.Errorf("Expected hit rate %f, got %f", expectedRate, hitRate)
	}
}

// TestTrackModelSelectionSavings tests model selection savings tracking
func TestTrackModelSelectionSavings(t *testing.T) {
	projectID := "test-project-savings"

	// Reset counters
	modelSelectionSavingsCounter.Delete(projectID)
	cheaperModelSelectedCounter.Delete(projectID)
	expensiveModelSelectedCounter.Delete(projectID)

	// Track some savings
	trackModelSelectionSavings(projectID, 0.05, true)  // $0.05 savings from cheaper model
	trackModelSelectionSavings(projectID, 0.03, true)  // Another $0.03 savings
	trackModelSelectionSavings(projectID, 0.02, false) // $0.02 from other optimization

	// Check total savings
	totalSavings := getModelSelectionSavings(projectID)
	expectedSavings := 0.10 // 0.05 + 0.03 + 0.02
	if totalSavings != expectedSavings {
		t.Errorf("Expected total savings %f, got %f", expectedSavings, totalSavings)
	}

	// Check cheaper model count
	if val, ok := cheaperModelSelectedCounter.Load(projectID); ok {
		count := val.(int64)
		if count != 2 {
			t.Errorf("Expected 2 cheaper model selections, got %d", count)
		}
	} else {
		t.Error("Expected cheaper model counter to exist")
	}
}

// TestBusinessContextCaching tests business context caching
func TestBusinessContextCaching(t *testing.T) {
	projectID := "test-project-bc"
	codebaseHash := "test-hash-123"
	config := &LLMConfig{
		CostOptimization: CostOptimizationConfig{
			UseCache:      true,
			CacheTTLHours: 24,
		},
	}

	// Reset cache
	businessContextCache.Delete(fmt.Sprintf("business:%s:%s", projectID, codebaseHash))
	cacheSizeCounter.Delete(projectID)

	// Create test data
	rules := []interface{}{
		map[string]interface{}{"id": "rule1", "title": "Test Rule 1"},
	}
	entities := []interface{}{
		map[string]interface{}{"id": "entity1", "name": "Test Entity"},
	}
	journeys := []interface{}{
		map[string]interface{}{"id": "journey1", "title": "Test Journey"},
	}

	// Set cached business context
	setCachedBusinessContext(projectID, codebaseHash, rules, entities, journeys, config)

	// Verify cache size was incremented
	if val, ok := cacheSizeCounter.Load(projectID); ok {
		size := val.(int64)
		if size != 1 {
			t.Errorf("Expected cache size 1, got %d", size)
		}
	} else {
		t.Error("Expected cache size counter to exist")
	}

	// Get cached business context
	cached, ok := getCachedBusinessContext(projectID, codebaseHash, config)
	if !ok {
		t.Error("Expected cache hit, got cache miss")
	}
	if cached == nil {
		t.Error("Expected cached data, got nil")
	}
	if cachedRules, ok := cached["rules"].([]interface{}); !ok || len(cachedRules) != 1 {
		t.Error("Cached rules don't match")
	}
}

// TestCacheSizeCounter tests cache size counter accuracy
func TestCacheSizeCounter(t *testing.T) {
	projectID := "test-project-size"
	featureHash := "test-feature-hash"
	depth := "medium"
	mode := "auto"
	config := &LLMConfig{
		CostOptimization: CostOptimizationConfig{
			UseCache: true,
		},
	}

	// Reset counter
	cacheSizeCounter.Delete(projectID)

	// Add first cache entry
	report1 := &ComprehensiveAnalysisReport{
		ValidationID: "validation-1",
		Feature:      "feature-1",
	}
	setCachedAnalysisResult(projectID, featureHash+"-1", depth, mode, report1, config)

	// Verify size is 1
	if val, ok := cacheSizeCounter.Load(projectID); ok {
		size := val.(int64)
		if size != 1 {
			t.Errorf("Expected cache size 1, got %d", size)
		}
	} else {
		t.Error("Expected cache size counter to exist")
	}

	// Add second cache entry
	report2 := &ComprehensiveAnalysisReport{
		ValidationID: "validation-2",
		Feature:      "feature-2",
	}
	setCachedAnalysisResult(projectID, featureHash+"-2", depth, mode, report2, config)

	// Verify size is 2
	if val, ok := cacheSizeCounter.Load(projectID); ok {
		size := val.(int64)
		if size != 2 {
			t.Errorf("Expected cache size 2, got %d", size)
		}
	}

	// Add same entry again (should not increment)
	setCachedAnalysisResult(projectID, featureHash+"-1", depth, mode, report1, config)

	// Verify size is still 2
	if val, ok := cacheSizeCounter.Load(projectID); ok {
		size := val.(int64)
		if size != 2 {
			t.Errorf("Expected cache size to remain 2, got %d", size)
		}
	}
}
