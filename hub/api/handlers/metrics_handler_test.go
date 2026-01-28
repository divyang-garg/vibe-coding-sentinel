// Package handlers - Unit tests for metrics handler
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockMetricsService implements MetricsService for testing
type mockMetricsService struct {
	mock.Mock
}

func (m *mockMetricsService) GetCacheMetrics(ctx context.Context, projectID string) (*models.CacheMetricsResponse, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CacheMetricsResponse), args.Error(1)
}

func (m *mockMetricsService) GetCostMetrics(ctx context.Context, projectID string) (*models.CostMetricsResponse, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CostMetricsResponse), args.Error(1)
}

func TestMetricsHandler_GetCacheMetrics_Success(t *testing.T) {
	mockService := new(mockMetricsService)
	handler := NewMetricsHandler(mockService)

	expectedResponse := &models.CacheMetricsResponse{
		ProjectID:     "project_123",
		HitRate:       0.75,
		Hits:          75,
		Misses:        25,
		TotalRequests: 100,
		CacheSize:     50,
	}

	mockService.On("GetCacheMetrics", mock.Anything, "project_123").Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/api/v1/metrics/cache", nil)
	ctx := context.WithValue(req.Context(), "project", &models.Project{ID: "project_123"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCacheMetrics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.CacheMetricsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.ProjectID, response.ProjectID)
	assert.Equal(t, expectedResponse.HitRate, response.HitRate)
	assert.Equal(t, expectedResponse.Hits, response.Hits)
	mockService.AssertExpectations(t)
}

func TestMetricsHandler_GetCacheMetrics_NoProjectContext(t *testing.T) {
	mockService := new(mockMetricsService)
	handler := NewMetricsHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/metrics/cache", nil)
	w := httptest.NewRecorder()

	handler.GetCacheMetrics(w, req)

	// WriteErrorResponse converts ValidationError to 400, not 401
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetCacheMetrics")
}

func TestMetricsHandler_GetCacheMetrics_ServiceError(t *testing.T) {
	mockService := new(mockMetricsService)
	handler := NewMetricsHandler(mockService)

	mockService.On("GetCacheMetrics", mock.Anything, "project_123").Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/api/v1/metrics/cache", nil)
	ctx := context.WithValue(req.Context(), "project", &models.Project{ID: "project_123"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCacheMetrics(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestMetricsHandler_GetCostMetrics_Success(t *testing.T) {
	mockService := new(mockMetricsService)
	handler := NewMetricsHandler(mockService)

	expectedResponse := &models.CostMetricsResponse{
		ProjectID:             "project_123",
		TotalCost:             125.50,
		TotalTokens:           50000,
		ModelSelectionSavings: 25.30,
		CheaperModelCount:     45,
		ExpensiveModelCount:   5,
	}

	mockService.On("GetCostMetrics", mock.Anything, "project_123").Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/api/v1/metrics/cost", nil)
	ctx := context.WithValue(req.Context(), "project", &models.Project{ID: "project_123"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCostMetrics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.CostMetricsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.ProjectID, response.ProjectID)
	assert.Equal(t, expectedResponse.TotalCost, response.TotalCost)
	assert.Equal(t, expectedResponse.TotalTokens, response.TotalTokens)
	mockService.AssertExpectations(t)
}

func TestMetricsHandler_GetCostMetrics_NoProjectContext(t *testing.T) {
	mockService := new(mockMetricsService)
	handler := NewMetricsHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/metrics/cost", nil)
	w := httptest.NewRecorder()

	handler.GetCostMetrics(w, req)

	// WriteErrorResponse converts ValidationError to 400, not 401
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetCostMetrics")
}

func TestMetricsHandler_GetCostMetrics_ServiceError(t *testing.T) {
	mockService := new(mockMetricsService)
	handler := NewMetricsHandler(mockService)

	mockService.On("GetCostMetrics", mock.Anything, "project_123").Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/api/v1/metrics/cost", nil)
	ctx := context.WithValue(req.Context(), "project", &models.Project{ID: "project_123"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCostMetrics(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
