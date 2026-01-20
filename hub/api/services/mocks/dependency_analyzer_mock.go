// Package mocks provides mock implementations for testing
// Complies with CODING_STANDARDS.md: Test utilities max 250 lines
package mocks

import (
	"context"

	"sentinel-hub-api/models"
	"sentinel-hub-api/repository"

	"github.com/stretchr/testify/mock"
)

// MockDependencyAnalyzer implements DependencyAnalyzer for testing
// This wraps the repository.DependencyAnalyzerImpl to allow mocking
type MockDependencyAnalyzer struct {
	mock.Mock
	*repository.DependencyAnalyzerImpl
}

// NewMockDependencyAnalyzer creates a new mock dependency analyzer
func NewMockDependencyAnalyzer() *MockDependencyAnalyzer {
	return &MockDependencyAnalyzer{
		DependencyAnalyzerImpl: repository.NewDependencyAnalyzer(),
	}
}

// DetectCycles detects cycles in dependencies (mockable)
func (m *MockDependencyAnalyzer) DetectCycles(ctx context.Context, dependencies []models.TaskDependency) ([][]string, error) {
	args := m.Called(ctx, dependencies)
	if args.Get(0) != nil {
		return args.Get(0).([][]string), args.Error(1)
	}
	// Fallback to real implementation if not mocked
	return m.DependencyAnalyzerImpl.DetectCycles(ctx, dependencies)
}
