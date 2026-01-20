// Package services provides shared test utilities and helpers
// Complies with CODING_STANDARDS.md: Test utilities max 250 lines
package services

import (
	"time"

	"sentinel-hub-api/models"
	"sentinel-hub-api/repository"
	"sentinel-hub-api/services/mocks"

	"github.com/stretchr/testify/assert"
)

// createTestTask creates a test task with default values
func createTestTask(id, title, status string) *models.Task {
	if id == "" {
		id = "test-task-" + time.Now().Format("20060102150405")
	}
	if title == "" {
		title = "Test Task"
	}
	if status == "" {
		status = string(models.TaskStatusPending)
	}

	return &models.Task{
		ID:        id,
		ProjectID: "test-project-123",
		Title:     title,
		Status:    models.TaskStatus(status),
		Priority:  models.TaskPriorityMedium,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}
}

// createTestDependency creates a test task dependency
func createTestDependency(from, to string) *models.TaskDependency {
	if from == "" {
		from = "task-1"
	}
	if to == "" {
		to = "task-2"
	}

	return &models.TaskDependency{
		ID:              "dep-" + from + "-" + to,
		TaskID:          from,
		DependsOnTaskID: to,
		DependencyType:  "finish_to_start",
		Confidence:      0.9,
		CreatedAt:       time.Now(),
	}
}

// createTestTaskService creates a task service with mocks
func createTestTaskService(mockRepo *mocks.MockTaskRepository, mockDep *mocks.MockDependencyAnalyzer, mockImpact *mocks.MockImpactAnalyzer) TaskService {
	if mockRepo == nil {
		mockRepo = mocks.NewMockTaskRepository()
	}
	if mockDep == nil {
		mockDep = mocks.NewMockDependencyAnalyzer()
	}
	if mockImpact == nil {
		mockImpact = mocks.NewMockImpactAnalyzer()
	}
	// Create concrete implementations for test
	depAnalyzer := &repository.DependencyAnalyzerImpl{}
	impactAnalyzer := &repository.ImpactAnalyzerImpl{}
	return NewTaskService(mockRepo, depAnalyzer, impactAnalyzer)
}

// assertTaskEqual asserts that two tasks are equal (ignoring timestamps)
func assertTaskEqual(t assert.TestingT, expected, actual *models.Task, msgAndArgs ...interface{}) bool {
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return assert.Fail(t, "One task is nil", msgAndArgs...)
	}

	// Compare key fields
	equal := assert.Equal(t, expected.ID, actual.ID, msgAndArgs...)
	equal = assert.Equal(t, expected.ProjectID, actual.ProjectID, msgAndArgs...) && equal
	equal = assert.Equal(t, expected.Title, actual.Title, msgAndArgs...) && equal
	equal = assert.Equal(t, expected.Status, actual.Status, msgAndArgs...) && equal
	equal = assert.Equal(t, expected.Priority, actual.Priority, msgAndArgs...) && equal
	equal = assert.Equal(t, expected.Description, actual.Description, msgAndArgs...) && equal

	return equal
}

// assertDependencyEqual asserts that two dependencies are equal
func assertDependencyEqual(t assert.TestingT, expected, actual *models.TaskDependency, msgAndArgs ...interface{}) bool {
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return assert.Fail(t, "One dependency is nil", msgAndArgs...)
	}

	equal := assert.Equal(t, expected.TaskID, actual.TaskID, msgAndArgs...)
	equal = assert.Equal(t, expected.DependsOnTaskID, actual.DependsOnTaskID, msgAndArgs...) && equal
	equal = assert.Equal(t, expected.DependencyType, actual.DependencyType, msgAndArgs...) && equal
	equal = assert.Equal(t, expected.Confidence, actual.Confidence, msgAndArgs...) && equal

	return equal
}

// createTestCreateTaskRequest creates a test CreateTaskRequest
func createTestCreateTaskRequest() models.CreateTaskRequest {
	return models.CreateTaskRequest{
		ProjectID:   "test-project-123",
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    "high",
		Source:      "manual",
	}
}

// createTestUpdateTaskRequest creates a test UpdateTaskRequest
func createTestUpdateTaskRequest() models.UpdateTaskRequest {
	title := "Updated Title"
	status := "in_progress"
	return models.UpdateTaskRequest{
		Title:  &title,
		Status: &status,
	}
}

// createTestAddDependencyRequest creates a test AddDependencyRequest
func createTestAddDependencyRequest(dependsOnTaskID string) models.AddDependencyRequest {
	if dependsOnTaskID == "" {
		dependsOnTaskID = "task-2"
	}
	return models.AddDependencyRequest{
		DependsOnTaskID: dependsOnTaskID,
		DependencyType:  "finish_to_start",
		Confidence:      0.9,
	}
}

// createTestVerifyTaskRequest creates a test VerifyTaskRequest
func createTestVerifyTaskRequest() models.VerifyTaskRequest {
	return models.VerifyTaskRequest{
		Status:     "verified",
		Confidence: 0.95,
		VerifiedBy: "user@example.com",
		VerifiedAt: time.Now(),
		Notes:      "Task completed successfully",
	}
}

// createTestTaskChange creates a test TaskChange
func createTestTaskChange(changeType, field string, newValue interface{}) models.TaskChange {
	return models.TaskChange{
		TaskID:     "task-1",
		ChangeType: changeType,
		Field:      field,
		NewValue:   newValue,
		ChangedAt:  time.Now(),
	}
}
