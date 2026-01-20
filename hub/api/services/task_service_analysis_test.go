// Package services provides testing for task service analysis.
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"errors"
	"testing"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskService_AnalyzeTaskImpact(t *testing.T) {
	t.Run("success_with_analyzer", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task := &models.Task{
			ID:     "task-1",
			Title:  "Test Task",
			Status: models.TaskStatusPending,
		}

		dependencies := []models.TaskDependency{
			{ID: "dep-1", TaskID: "task-1", DependsOnTaskID: "task-2"},
		}

		expectedAnalysis := &models.TaskImpactAnalysis{
			ID:         "impact-1",
			TaskID:     "task-1",
			ChangeType: "task_change",
			RiskLevel:  "medium",
		}

		change := models.TaskChange{
			TaskID:     "task-1",
			ChangeType: "status_change",
			Field:      "status",
			NewValue:   "in_progress",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return(dependencies, nil)
		mockImpactAnalyzer.On("AnalyzeImpact", mock.Anything, "task-1", "task_change", mock.AnythingOfType("[]models.Task"), dependencies).Return(expectedAnalysis, nil)

		// When
		analysis, err := service.AnalyzeTaskImpact(context.Background(), "task-1", change)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, expectedAnalysis.ID, analysis.ID)
		assert.Equal(t, expectedAnalysis.TaskID, analysis.TaskID)
		mockRepo.AssertExpectations(t)
		mockImpactAnalyzer.AssertExpectations(t)
	})

	t.Run("success_without_analyzer_fallback", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		service := NewTaskService(mockRepo, nil, nil)

		task := &models.Task{
			ID:     "task-1",
			Title:  "Test Task",
			Status: models.TaskStatusPending,
		}

		dependencies := []models.TaskDependency{}

		change := models.TaskChange{
			TaskID:     "task-1",
			ChangeType: "status_change",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return(dependencies, nil)

		// When
		analysis, err := service.AnalyzeTaskImpact(context.Background(), "task-1", change)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, "task-1", analysis.TaskID)
		assert.Equal(t, "task_change", analysis.ChangeType)
		assert.Equal(t, "medium", analysis.RiskLevel)
		assert.Contains(t, analysis.AffectedTasks, "Test Task")
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error_empty_id", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		change := models.TaskChange{
			ChangeType: "status_change",
		}

		// When
		analysis, err := service.AnalyzeTaskImpact(context.Background(), "", change)

		// Then
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Contains(t, err.Error(), "task ID is required")
		mockRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("task_not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		change := models.TaskChange{
			TaskID:     "task-1",
			ChangeType: "status_change",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(nil, nil)

		// When
		analysis, err := service.AnalyzeTaskImpact(context.Background(), "task-1", change)

		// Then
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Contains(t, err.Error(), "task not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error_finding_dependencies", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task := &models.Task{ID: "task-1"}

		change := models.TaskChange{
			TaskID:     "task-1",
			ChangeType: "status_change",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return(nil, errors.New("database error"))

		// When
		analysis, err := service.AnalyzeTaskImpact(context.Background(), "task-1", change)

		// Then
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Contains(t, err.Error(), "failed to get dependencies")
		mockRepo.AssertExpectations(t)
	})

	t.Run("analyzer_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task := &models.Task{ID: "task-1"}

		dependencies := []models.TaskDependency{}

		change := models.TaskChange{
			TaskID:     "task-1",
			ChangeType: "status_change",
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-1").Return(dependencies, nil)
		mockImpactAnalyzer.On("AnalyzeImpact", mock.Anything, "task-1", "task_change", mock.AnythingOfType("[]models.Task"), dependencies).Return(nil, errors.New("analysis error"))

		// When
		analysis, err := service.AnalyzeTaskImpact(context.Background(), "task-1", change)

		// Then
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Contains(t, err.Error(), "impact analysis failed")
		mockRepo.AssertExpectations(t)
		mockImpactAnalyzer.AssertExpectations(t)
	})
}

func TestTaskService_GetTaskImpactAnalysis(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task := &models.Task{
			ID:     "task-1",
			Title:  "Test Task",
			Status: models.TaskStatusPending,
		}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)

		// When
		analysis, err := service.GetTaskImpactAnalysis(context.Background(), "task-1")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, "task-1", analysis.TaskID)
		assert.Equal(t, "analysis", analysis.ChangeType)
		assert.Equal(t, "task", analysis.ImpactScope)
		assert.Equal(t, "low", analysis.RiskLevel)
		assert.Contains(t, analysis.AffectedTasks, "task-1")
		assert.NotEmpty(t, analysis.AnalyzedAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error_empty_id", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		// When
		analysis, err := service.GetTaskImpactAnalysis(context.Background(), "")

		// Then
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Contains(t, err.Error(), "task ID is required")
		mockRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("task_not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(nil, nil)

		// When
		analysis, err := service.GetTaskImpactAnalysis(context.Background(), "task-1")

		// Then
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Contains(t, err.Error(), "task not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(nil, errors.New("database error"))

		// When
		analysis, err := service.GetTaskImpactAnalysis(context.Background(), "task-1")

		// Then
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Contains(t, err.Error(), "failed to find task")
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetTaskExecutionPlan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		task1 := &models.Task{ID: "task-1", Title: "Task 1"}
		task2 := &models.Task{ID: "task-2", Title: "Task 2"}

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(task1, nil)
		mockRepo.On("FindByID", mock.Anything, "task-2").Return(task2, nil)

		// When
		plan, err := service.GetTaskExecutionPlan(context.Background(), []string{"task-1", "task-2"})

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Equal(t, 2, len(plan.Tasks))
		assert.Equal(t, 1, len(plan.Batches))
		assert.Equal(t, 2, len(plan.Batches[0]))
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error_empty_task_ids", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		// When
		plan, err := service.GetTaskExecutionPlan(context.Background(), []string{})

		// Then
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "no task IDs provided")
		mockRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("task_not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(nil, nil)

		// When
		plan, err := service.GetTaskExecutionPlan(context.Background(), []string{"task-1"})

		// Then
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "failed to find task")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindByID", mock.Anything, "task-1").Return(nil, errors.New("database error"))

		// When
		plan, err := service.GetTaskExecutionPlan(context.Background(), []string{"task-1"})

		// Then
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "failed to find task")
		mockRepo.AssertExpectations(t)
	})
}
