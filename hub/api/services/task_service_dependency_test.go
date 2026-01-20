// Package services provides testing for task service dependencies.
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskService_AddDependency(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task1 := &models.Task{ID: "task-1", ProjectID: "project-123"}
		task2 := &models.Task{ID: "task-2", ProjectID: "project-123"}

		req := models.AddDependencyRequest{
			DependsOnTaskID: "task-2",
			DependencyType:  "finish_to_start",
			Confidence:      0.9,
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task1, nil)
		mockRepo.On("FindByID", mock.Anything, "task-2").Return(task2, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return([]models.TaskDependency{}, nil)
		mockDepAnalyzer.On("DetectCycles", mock.Anything, mock.AnythingOfType("[]models.TaskDependency")).Return([][]string{}, nil)
		mockRepo.On("SaveDependency", mock.Anything, mock.AnythingOfType("*models.TaskDependency")).Return(nil)

		// When
		dependency, err := service.AddDependency(context.Background(), "task-1", req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, dependency)
		assert.Equal(t, "task-1", dependency.TaskID)
		assert.Equal(t, "task-2", dependency.DependsOnTaskID)
		assert.Equal(t, req.DependencyType, dependency.DependencyType)
		assert.Equal(t, req.Confidence, dependency.Confidence)
		mockRepo.AssertExpectations(t)
		mockDepAnalyzer.AssertExpectations(t)
	})

	t.Run("validation_error_empty_task_id", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.AddDependencyRequest{
			DependsOnTaskID: "task-2",
		}

		// When
		dependency, err := service.AddDependency(context.Background(), "", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, dependency)
		assert.Contains(t, err.Error(), "task ID is required")
		mockRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("validation_error_empty_depends_on", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.AddDependencyRequest{}

		// When
		dependency, err := service.AddDependency(context.Background(), "task-1", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, dependency)
		assert.Contains(t, err.Error(), "depends on task ID is required")
		mockRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("task_not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.AddDependencyRequest{
			DependsOnTaskID: "task-2",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(nil, nil)

		// When
		dependency, err := service.AddDependency(context.Background(), "task-1", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, dependency)
		assert.Contains(t, err.Error(), "task not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("depends_on_task_not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task1 := &models.Task{ID: "task-1", ProjectID: "project-123"}

		req := models.AddDependencyRequest{
			DependsOnTaskID: "task-2",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task1, nil)
		mockRepo.On("FindByID", mock.Anything, "task-2").Return(nil, nil)

		// When
		dependency, err := service.AddDependency(context.Background(), "task-1", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, dependency)
		assert.Contains(t, err.Error(), "depends on task not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("circular_dependency_detected", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task1 := &models.Task{ID: "task-1", ProjectID: "project-123"}
		task2 := &models.Task{ID: "task-2", ProjectID: "project-123"}

		req := models.AddDependencyRequest{
			DependsOnTaskID: "task-2",
		}

		cycles := [][]string{{"task-1", "task-2", "task-1"}}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task1, nil)
		mockRepo.On("FindByID", mock.Anything, "task-2").Return(task2, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return([]models.TaskDependency{}, nil)
		mockDepAnalyzer.On("DetectCycles", mock.Anything, mock.AnythingOfType("[]models.TaskDependency")).Return(cycles, nil)

		// When
		dependency, err := service.AddDependency(context.Background(), "task-1", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, dependency)
		assert.Contains(t, err.Error(), "dependency would create a cycle")
		mockRepo.AssertExpectations(t)
		mockDepAnalyzer.AssertExpectations(t)
	})

	t.Run("repository_error_saving_dependency", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task1 := &models.Task{ID: "task-1", ProjectID: "project-123"}
		task2 := &models.Task{ID: "task-2", ProjectID: "project-123"}

		req := models.AddDependencyRequest{
			DependsOnTaskID: "task-2",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task1, nil)
		mockRepo.On("FindByID", mock.Anything, "task-2").Return(task2, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return([]models.TaskDependency{}, nil)
		mockDepAnalyzer.On("DetectCycles", mock.Anything, mock.AnythingOfType("[]models.TaskDependency")).Return([][]string{}, nil)
		mockRepo.On("SaveDependency", mock.Anything, mock.AnythingOfType("*models.TaskDependency")).Return(errors.New("database error"))

		// When
		dependency, err := service.AddDependency(context.Background(), "task-1", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, dependency)
		assert.Contains(t, err.Error(), "failed to save dependency")
		mockRepo.AssertExpectations(t)
		mockDepAnalyzer.AssertExpectations(t)
	})
}

func TestTaskService_GetDependencies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		dependencies := []models.TaskDependency{
			{ID: "dep-1", TaskID: "task-1", DependsOnTaskID: "task-2"},
			{ID: "dep-2", TaskID: "task-1", DependsOnTaskID: "task-3"},
		}

		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return(dependencies, nil)

		// When
		response, err := service.GetDependencies(context.Background(), "task-1")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Graph)
		assert.Equal(t, 2, len(response.Graph.Dependencies))
		assert.True(t, response.IsValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no_dependencies", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return([]models.TaskDependency{}, nil)

		// When
		response, err := service.GetDependencies(context.Background(), "task-1")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 0, len(response.Graph.Dependencies))
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return(nil, errors.New("database error"))

		// When
		response, err := service.GetDependencies(context.Background(), "task-1")

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get dependencies")
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_AnalyzeDependencies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		dependencies := []models.TaskDependency{
			{ID: "dep-1", TaskID: "task-1", DependsOnTaskID: "task-2"},
		}

		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return(dependencies, nil)

		// When
		graph, err := service.AnalyzeDependencies(context.Background(), "task-1")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, graph)
		assert.Equal(t, 1, len(graph.Dependencies))
		assert.True(t, graph.IsValid)
		assert.NotEmpty(t, graph.GeneratedAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return(nil, errors.New("database error"))

		// When
		graph, err := service.AnalyzeDependencies(context.Background(), "task-1")

		// Then
		assert.Error(t, err)
		assert.Nil(t, graph)
		assert.Contains(t, err.Error(), "failed to get dependencies")
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_VerifyTask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task := &models.Task{
			ID:     "task-1",
			Status: models.TaskStatusCompleted,
		}

		req := models.VerifyTaskRequest{
			Status:     "verified",
			Confidence: 0.95,
			VerifiedBy: "user@example.com",
			VerifiedAt: time.Now(),
			Notes:      "Task completed successfully",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
		mockRepo.On("SaveVerification", mock.Anything, mock.AnythingOfType("*models.TaskVerification")).Return(nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)

		// When
		response, err := service.VerifyTask(context.Background(), "task-1", req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Task)
		assert.NotNil(t, response.Verification)
		assert.Equal(t, req.Confidence, response.Task.VerificationConfidence)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error_empty_id", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.VerifyTaskRequest{
			Status: "verified",
		}

		// When
		response, err := service.VerifyTask(context.Background(), "", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "task ID is required")
		mockRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("task_not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.VerifyTaskRequest{
			Status: "verified",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(nil, nil)

		// When
		response, err := service.VerifyTask(context.Background(), "task-1", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "task not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error_saving_verification", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task := &models.Task{ID: "task-1"}

		req := models.VerifyTaskRequest{
			Status:     "verified",
			VerifiedAt: time.Now(),
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
		mockRepo.On("SaveVerification", mock.Anything, mock.AnythingOfType("*models.TaskVerification")).Return(errors.New("database error"))

		// When
		response, err := service.VerifyTask(context.Background(), "task-1", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to save verification")
		mockRepo.AssertExpectations(t)
	})
}
