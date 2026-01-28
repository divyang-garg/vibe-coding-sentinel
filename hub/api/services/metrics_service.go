// Package services - Metrics service
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"

	"sentinel-hub-api/models"
	"sentinel-hub-api/pkg"
	"sentinel-hub-api/repository"
)

// MetricsServiceImpl implements MetricsService interface
type MetricsServiceImpl struct {
	llmUsageRepo repository.LLMUsageRepository
}

// NewMetricsService creates a new metrics service
func NewMetricsService(llmUsageRepo repository.LLMUsageRepository) MetricsService {
	return &MetricsServiceImpl{
		llmUsageRepo: llmUsageRepo,
	}
}

// GetCacheMetrics retrieves cache metrics for a project
func (s *MetricsServiceImpl) GetCacheMetrics(ctx context.Context, projectID string) (*models.CacheMetricsResponse, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Get cache metrics from shared pkg package
	hits := pkg.GetCacheHits(projectID)
	misses := pkg.GetCacheMisses(projectID)
	hitRate := pkg.GetCacheHitRate(projectID)
	cacheSize := pkg.GetCacheSize(projectID)
	total := hits + misses

	return &models.CacheMetricsResponse{
		ProjectID:     projectID,
		HitRate:       hitRate,
		Hits:          hits,
		Misses:        misses,
		TotalRequests: total,
		CacheSize:     cacheSize,
	}, nil
}

// GetCostMetrics retrieves cost metrics for a project
func (s *MetricsServiceImpl) GetCostMetrics(ctx context.Context, projectID string) (*models.CostMetricsResponse, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Get model selection metrics from shared pkg package
	modelSelectionSavings := pkg.GetModelSelectionSavings(projectID)
	cheaperModelCount := pkg.GetCheaperModelCount(projectID)
	expensiveModelCount := pkg.GetExpensiveModelCount(projectID)

	// Get total cost and tokens from database using LLMUsageRepository
	// Query all usage records for the project (with reasonable limit)
	usages, _, err := s.llmUsageRepo.GetUsageByProject(ctx, projectID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to query LLM usage: %w", err)
	}

	// Calculate totals from usage records
	var totalCost float64
	var totalTokens int64
	for _, usage := range usages {
		totalCost += usage.EstimatedCost
		totalTokens += int64(usage.TokensUsed)
	}

	return &models.CostMetricsResponse{
		ProjectID:             projectID,
		TotalCost:             totalCost,
		TotalTokens:           totalTokens,
		ModelSelectionSavings: modelSelectionSavings,
		CheaperModelCount:     cheaperModelCount,
		ExpensiveModelCount:   expensiveModelCount,
	}, nil
}
