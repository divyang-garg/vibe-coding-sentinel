// Package services provides testing for task service CRUD operations.
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sentinel-hub-api/models"
	"sentinel-hub-api/services/mocks"
)

func TestTaskService_CreateTask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.CreateTaskRequest{
			ProjectID:   "project-123",
			Title:       "Test Task",
			Description: "Test Description",
			Priority:    "high",
		}

		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)

		// When
		task, err := service.CreateTask(context.Background(), req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, req.ProjectID, task.ProjectID)
		assert.Equal(t, req.Title, task.Title)
		assert.Equal(t, req.Description, task.Description)
		assert.Equal(t, models.TaskStatusPending, task.Status)
		assert.Equal(t, models.TaskPriority(req.Priority), task.Priority)
		assert.NotEmpty(t, task.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error_missing_project_id", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.CreateTaskRequest{
			Title: "Test Task",
		}

		// When
		task, err := service.CreateTask(context.Background(), req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "project ID is required")
		mockRepo.AssertNotCalled(t, "Save")
	})

	t.Run("validation_error_missing_title", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.CreateTaskRequest{
			ProjectID: "project-123",
		}

		// When
		task, err := service.CreateTask(context.Background(), req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "title is required")
		mockRepo.AssertNotCalled(t, "Save")
	})

	t.Run("repository_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		service := NewTaskService(mockRepo, nil, nil)
		req := models.CreateTaskRequest{ProjectID: "project-123", Title: "Test Task"}
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*models.Task")).Return(errors.New("database error"))

		// When
		task, err := service.CreateTask(context.Background(), req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "failed to save task")
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetTaskByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		expectedTask := &models.Task{
			ID:        "task-123",
			ProjectID: "project-123",
			Title:     "Test Task",
			Status:    models.TaskStatusPending,
		}

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(expectedTask, nil)

		// When
		task, err := service.GetTaskByID(context.Background(), "task-123")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, expectedTask.ID, task.ID)
		assert.Equal(t, expectedTask.Title, task.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error_empty_id", func(t *testing.T) {
		// Given
		service := NewTaskService(mocks.NewMockTaskRepository(), nil, nil)

		// When
		task, err := service.GetTaskByID(context.Background(), "")

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "task ID is required")
	})

	t.Run("not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(nil, nil)

		// When
		task, err := service.GetTaskByID(context.Background(), "task-123")

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "task not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(nil, errors.New("database error"))

		// When
		task, err := service.GetTaskByID(context.Background(), "task-123")

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "failed to find task")
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		existingTask := &models.Task{
			ID:        "task-123",
			ProjectID: "project-123",
			Title:     "Original Title",
			Status:    models.TaskStatusPending,
			Version:   1,
		}

		newTitle := "Updated Title"
		newStatus := "in_progress"
		req := models.UpdateTaskRequest{
			Title:  &newTitle,
			Status: &newStatus,
		}

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(existingTask, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil).Run(func(args mock.Arguments) {
			task := args.Get(1).(*models.Task)
			assert.Equal(t, newTitle, task.Title)
			assert.Equal(t, models.TaskStatus(newStatus), task.Status)
		})

		// When
		task, err := service.UpdateTask(context.Background(), "task-123", req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, newTitle, task.Title)
		assert.Equal(t, models.TaskStatus(newStatus), task.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error_empty_id", func(t *testing.T) {
		// Given
		service := NewTaskService(mocks.NewMockTaskRepository(), nil, nil)

		// When
		task, err := service.UpdateTask(context.Background(), "", models.UpdateTaskRequest{})

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "task ID is required")
	})

	t.Run("not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(nil, nil)

		req := models.UpdateTaskRequest{}

		// When
		task, err := service.UpdateTask(context.Background(), "task-123", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "task not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_update_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		existingTask := &models.Task{
			ID:     "task-123",
			Status: models.TaskStatusPending,
		}

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(existingTask, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(errors.New("database error"))

		req := models.UpdateTaskRequest{}

		// When
		task, err := service.UpdateTask(context.Background(), "task-123", req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "failed to update task")
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_ListTasks(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		tasks := []models.Task{
			{ID: "task-1", ProjectID: "project-123", Title: "Task 1"},
			{ID: "task-2", ProjectID: "project-123", Title: "Task 2"},
		}

		req := models.ListTasksRequest{
			ProjectID: "project-123",
			Limit:     10,
			Offset:    0,
		}

		mockRepo.On("FindByProjectID", mock.Anything, "project-123", req).Return(tasks, 2, nil)

		// When
		response, err := service.ListTasks(context.Background(), req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 2, len(response.Tasks))
		assert.Equal(t, 2, response.Total)
		assert.Equal(t, 10, response.Limit)
		assert.Equal(t, 0, response.Offset)
		assert.False(t, response.HasMore)
		mockRepo.AssertExpectations(t)
	})

	t.Run("with_pagination", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		tasks := []models.Task{
			{ID: "task-1", ProjectID: "project-123", Title: "Task 1"},
		}

		req := models.ListTasksRequest{
			ProjectID: "project-123",
			Limit:     1,
			Offset:    0,
		}

		mockRepo.On("FindByProjectID", mock.Anything, "project-123", req).Return(tasks, 5, nil)

		// When
		response, err := service.ListTasks(context.Background(), req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, len(response.Tasks))
		assert.Equal(t, 5, response.Total)
		assert.True(t, response.HasMore)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		req := models.ListTasksRequest{
			ProjectID: "project-123",
			Limit:     10,
			Offset:    0,
		}

		mockRepo.On("FindByProjectID", mock.Anything, "project-123", req).Return(nil, 0, errors.New("database error"))

		// When
		response, err := service.ListTasks(context.Background(), req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to list tasks")
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		existingTask := &models.Task{
			ID:     "task-123",
			Status: models.TaskStatusPending,
		}

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(existingTask, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-123").Return([]models.TaskDependency{}, nil)
		mockRepo.On("Delete", mock.Anything, "task-123").Return(nil)

		// When
		err := service.DeleteTask(context.Background(), "task-123")

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error_empty_id", func(t *testing.T) {
		// Given
		service := NewTaskService(mocks.NewMockTaskRepository(), nil, nil)

		// When
		err := service.DeleteTask(context.Background(), "")

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task ID is required")
	})

	t.Run("not_found", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(nil, nil)

		// When
		err := service.DeleteTask(context.Background(), "task-123")

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("with_dependencies", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		existingTask := &models.Task{
			ID:     "task-123",
			Status: models.TaskStatusPending,
		}

		dependencies := []models.TaskDependency{
			{ID: "dep-1", TaskID: "task-123", DependsOnTaskID: "task-456"},
		}

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(existingTask, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-123").Return(dependencies, nil)
		mockRepo.On("DeleteDependency", mock.Anything, "dep-1").Return(nil)
		mockRepo.On("Delete", mock.Anything, "task-123").Return(nil)

		// When
		err := service.DeleteTask(context.Background(), "task-123")

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("dependency_deletion_error", func(t *testing.T) {
		// Given
		mockRepo := mocks.NewMockTaskRepository()
		mockDepAnalyzer := mocks.NewMockDependencyAnalyzer()
		mockImpactAnalyzer := mocks.NewMockImpactAnalyzer()
		service := NewTaskService(mockRepo, mockDepAnalyzer, mockImpactAnalyzer)

		existingTask := &models.Task{
			ID:     "task-123",
			Status: models.TaskStatusPending,
		}

		dependencies := []models.TaskDependency{
			{ID: "dep-1", TaskID: "task-123", DependsOnTaskID: "task-456"},
		}

		mockRepo.On("FindByID", mock.Anything, "task-123").Return(existingTask, nil)
		mockRepo.On("FindDependencies", mock.Anything, "task-123").Return(dependencies, nil)
		mockRepo.On("DeleteDependency", mock.Anything, "dep-1").Return(errors.New("database error"))

		// When
		err := service.DeleteTask(context.Background(), "task-123")

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete dependency")
		mockRepo.AssertExpectations(t)
	})
}
