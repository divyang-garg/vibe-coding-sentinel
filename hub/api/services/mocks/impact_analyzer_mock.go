// Package mocks provides mock implementations for testing
// Complies with CODING_STANDARDS.md: Test utilities max 250 lines
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"sentinel-hub-api/models"
	"sentinel-hub-api/repository"
)

// MockImpactAnalyzer implements ImpactAnalyzer for testing
// This wraps the repository.ImpactAnalyzerImpl to allow mocking
type MockImpactAnalyzer struct {
	mock.Mock
	*repository.ImpactAnalyzerImpl
}

// NewMockImpactAnalyzer creates a new mock impact analyzer
func NewMockImpactAnalyzer() *MockImpactAnalyzer {
	return &MockImpactAnalyzer{
		ImpactAnalyzerImpl: repository.NewImpactAnalyzer(),
	}
}

// AnalyzeImpact analyzes impact of task changes (mockable)
func (m *MockImpactAnalyzer) AnalyzeImpact(ctx context.Context, taskID string, changeType string, tasks []models.Task, dependencies []models.TaskDependency) (*models.TaskImpactAnalysis, error) {
	args := m.Called(ctx, taskID, changeType, tasks, dependencies)
	if args.Get(0) != nil {
		return args.Get(0).(*models.TaskImpactAnalysis), args.Error(1)
	}
	// Fallback to real implementation if not mocked
	return m.ImpactAnalyzerImpl.AnalyzeImpact(ctx, taskID, changeType, tasks, dependencies)
}
