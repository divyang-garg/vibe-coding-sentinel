// Package services - Unit tests for metrics service
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"testing"

	"sentinel-hub-api/models"
	"sentinel-hub-api/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockLLMUsageRepository implements LLMUsageRepository interface for testing
type mockLLMUsageRepository struct {
	mock.Mock
}

func (m *mockLLMUsageRepository) SaveUsage(ctx context.Context, usage *models.LLMUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *mockLLMUsageRepository) GetUsageByProject(ctx context.Context, projectID string, limit, offset int) ([]*models.LLMUsage, int, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.LLMUsage), args.Int(1), args.Error(2)
}

func (m *mockLLMUsageRepository) GetUsageByValidationID(ctx context.Context, validationID string) ([]*models.LLMUsage, error) {
	args := m.Called(ctx, validationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LLMUsage), args.Error(1)
}

func TestMetricsService_GetCacheMetrics_Success(t *testing.T) {
	// Set up test data in shared cache metrics
	projectID := "test_project_123"
	pkg.CacheHitCounter.Store(projectID, int64(75))
	pkg.CacheMissCounter.Store(projectID, int64(25))
	pkg.CacheSizeCounter.Store(projectID, int64(50))

	mockRepo := new(mockLLMUsageRepository)
	service := NewMetricsService(mockRepo)

	result, err := service.GetCacheMetrics(context.Background(), projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, 0.75, result.HitRate) // 75/(75+25) = 0.75
	assert.Equal(t, int64(75), result.Hits)
	assert.Equal(t, int64(25), result.Misses)
	assert.Equal(t, int64(100), result.TotalRequests)
	assert.Equal(t, int64(50), result.CacheSize)

	// Cleanup
	pkg.CacheHitCounter.Delete(projectID)
	pkg.CacheMissCounter.Delete(projectID)
	pkg.CacheSizeCounter.Delete(projectID)
}

func TestMetricsService_GetCacheMetrics_NoData(t *testing.T) {
	mockRepo := new(mockLLMUsageRepository)
	service := NewMetricsService(mockRepo)

	result, err := service.GetCacheMetrics(context.Background(), "nonexistent_project")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "nonexistent_project", result.ProjectID)
	assert.Equal(t, 0.0, result.HitRate)
	assert.Equal(t, int64(0), result.Hits)
	assert.Equal(t, int64(0), result.Misses)
}

func TestMetricsService_GetCacheMetrics_ContextCancellation(t *testing.T) {
	mockRepo := new(mockLLMUsageRepository)
	service := NewMetricsService(mockRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := service.GetCacheMetrics(ctx, "project_123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, context.Canceled, err)
}

func TestMetricsService_GetCostMetrics_Success(t *testing.T) {
	projectID := "test_project_123"

	// Set up test data in shared cache metrics
	pkg.ModelSelectionSavingsCounter.Store(projectID, 25.30)
	pkg.CheaperModelSelectedCounter.Store(projectID, int64(45))
	pkg.ExpensiveModelSelectedCounter.Store(projectID, int64(5))

	// Set up mock repository
	mockRepo := new(mockLLMUsageRepository)
	usages := []*models.LLMUsage{
		{
			ID:            "usage_1",
			ProjectID:     projectID,
			TokensUsed:    10000,
			EstimatedCost: 50.25,
		},
		{
			ID:            "usage_2",
			ProjectID:     projectID,
			TokensUsed:    20000,
			EstimatedCost: 75.25,
		},
	}
	mockRepo.On("GetUsageByProject", context.Background(), projectID, 10000, 0).Return(usages, 2, nil)

	service := NewMetricsService(mockRepo)

	result, err := service.GetCostMetrics(context.Background(), projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, 125.50, result.TotalCost)         // 50.25 + 75.25
	assert.Equal(t, int64(30000), result.TotalTokens) // 10000 + 20000
	assert.Equal(t, 25.30, result.ModelSelectionSavings)
	assert.Equal(t, int64(45), result.CheaperModelCount)
	assert.Equal(t, int64(5), result.ExpensiveModelCount)

	mockRepo.AssertExpectations(t)

	// Cleanup
	pkg.ModelSelectionSavingsCounter.Delete(projectID)
	pkg.CheaperModelSelectedCounter.Delete(projectID)
	pkg.ExpensiveModelSelectedCounter.Delete(projectID)
}

func TestMetricsService_GetCostMetrics_NoUsageData(t *testing.T) {
	projectID := "test_project_empty"

	mockRepo := new(mockLLMUsageRepository)
	mockRepo.On("GetUsageByProject", context.Background(), projectID, 10000, 0).Return([]*models.LLMUsage{}, 0, nil)

	service := NewMetricsService(mockRepo)

	result, err := service.GetCostMetrics(context.Background(), projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, 0.0, result.TotalCost)
	assert.Equal(t, int64(0), result.TotalTokens)

	mockRepo.AssertExpectations(t)
}

func TestMetricsService_GetCostMetrics_RepositoryError(t *testing.T) {
	projectID := "test_project_error"

	mockRepo := new(mockLLMUsageRepository)
	mockRepo.On("GetUsageByProject", context.Background(), projectID, 10000, 0).Return(nil, 0, assert.AnError)

	service := NewMetricsService(mockRepo)

	result, err := service.GetCostMetrics(context.Background(), projectID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to query LLM usage")

	mockRepo.AssertExpectations(t)
}

func TestMetricsService_GetCostMetrics_ContextCancellation(t *testing.T) {
	mockRepo := new(mockLLMUsageRepository)
	service := NewMetricsService(mockRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := service.GetCostMetrics(ctx, "project_123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, context.Canceled, err)
}
