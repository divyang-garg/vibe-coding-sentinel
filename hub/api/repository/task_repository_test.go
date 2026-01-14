// Package repository provides comprehensive testing for repository implementations.
package repository

import (
	"context"
	"testing"
	"time"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	args = append([]interface{}{ctx, query}, args...)
	call := m.Called(args...)
	return call.Get(0).(Row)
}

func (m *MockDatabase) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	args = append([]interface{}{ctx, query}, args...)
	call := m.Called(args...)
	return call.Get(0).(Rows), call.Error(1)
}

func (m *MockDatabase) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	args = append([]interface{}{ctx, query}, args...)
	call := m.Called(args...)
	return call.Get(0).(Result), call.Error(1)
}

func (m *MockDatabase) BeginTx(ctx context.Context) (Transaction, error) {
	call := m.Called(ctx)
	return call.Get(0).(Transaction), call.Error(1)
}

// MockRow implements the Row interface
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	call := m.Called(dest...)
	return call.Error(0)
}

// MockRows implements the Rows interface
type MockRows struct {
	mock.Mock
	rows  []models.Task
	index int
}

func (m *MockRows) Next() bool {
	if m.index < len(m.rows) {
		m.index++
		return true
	}
	return false
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.index == 0 || m.index > len(m.rows) {
		return nil
	}

	task := m.rows[m.index-1]
	dest[0] = task.ID
	dest[1] = task.ProjectID
	dest[2] = task.Source
	dest[3] = task.Title
	dest[4] = task.Description
	dest[5] = task.FilePath
	dest[6] = task.LineNumber
	dest[7] = task.Status
	dest[8] = task.Priority
	dest[9] = task.AssignedTo
	dest[10] = task.EstimatedEffort
	dest[11] = task.ActualEffort
	dest[12] = task.VerificationConfidence
	dest[13] = task.CreatedAt
	dest[14] = task.UpdatedAt
	dest[15] = task.CompletedAt
	dest[16] = task.VerifiedAt
	dest[17] = task.ArchivedAt
	dest[18] = task.Version

	return nil
}

func (m *MockRows) Close() error {
	return nil
}

func (m *MockRows) Err() error {
	return nil
}

// MockResult implements the Result interface
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	call := m.Called()
	return call.Get(0).(int64), call.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	call := m.Called()
	return call.Get(0).(int64), call.Error(1)
}

func TestTaskRepository_Save(t *testing.T) {
	mockDB := &MockDatabase{}
	mockResult := &MockResult{}

	repo := NewTaskRepository(mockDB)

	task := &models.Task{
		ID:          "task-123",
		ProjectID:   "project-456",
		Source:      "cursor",
		Title:       "Test Task",
		Description: "A test task",
		Status:      "pending",
		Priority:    "medium",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	ctx := context.Background()

	mockDB.On("Exec", ctx, mock.AnythingOfType("string"),
		task.ID, task.ProjectID, task.Source, task.Title, task.Description,
		task.FilePath, task.LineNumber, task.Status, task.Priority, task.AssignedTo,
		task.EstimatedEffort, task.VerificationConfidence, task.CreatedAt, task.UpdatedAt, task.Version).
		Return(mockResult, nil)

	err := repo.Save(ctx, task)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestTaskRepository_FindByID(t *testing.T) {
	mockDB := &MockDatabase{}
	mockRow := &MockRow{}

	repo := NewTaskRepository(mockDB)

	expectedTask := &models.Task{
		ID:          "task-123",
		ProjectID:   "project-456",
		Source:      "cursor",
		Title:       "Test Task",
		Description: "A test task",
		Status:      "pending",
		Priority:    "medium",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	ctx := context.Background()

	mockDB.On("QueryRow", ctx, mock.AnythingOfType("string"), "task-123").
		Return(mockRow)

	mockRow.On("Scan",
		&expectedTask.ID, &expectedTask.ProjectID, &expectedTask.Source, &expectedTask.Title, &expectedTask.Description,
		&expectedTask.FilePath, mock.AnythingOfType("*int"), &expectedTask.Status, &expectedTask.Priority,
		mock.AnythingOfType("**string"), mock.AnythingOfType("**int"), mock.AnythingOfType("**int"),
		&expectedTask.VerificationConfidence, &expectedTask.CreatedAt, &expectedTask.UpdatedAt,
		mock.AnythingOfType("**time.Time"), mock.AnythingOfType("**time.Time"), mock.AnythingOfType("**time.Time"),
		&expectedTask.Version).
		Run(func(args mock.Arguments) {
			// Simulate scanning the task data
			dest := args.Get(0).([]interface{})
			*dest[0].(*string) = expectedTask.ID
			*dest[1].(*string) = expectedTask.ProjectID
			*dest[2].(*string) = expectedTask.Source
			*dest[3].(*string) = expectedTask.Title
			*dest[4].(*string) = expectedTask.Description
			*dest[7].(*string) = string(expectedTask.Status)
			*dest[8].(*string) = string(expectedTask.Priority)
			*dest[12].(*float64) = expectedTask.VerificationConfidence
			*dest[13].(*time.Time) = expectedTask.CreatedAt
			*dest[14].(*time.Time) = expectedTask.UpdatedAt
			*dest[18].(*int) = expectedTask.Version
		}).
		Return(nil)

	task, err := repo.FindByID(ctx, "task-123")

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask.ID, task.ID)
	assert.Equal(t, expectedTask.ProjectID, task.ProjectID)
	assert.Equal(t, expectedTask.Title, task.Title)

	mockDB.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestTaskRepository_FindByProjectID(t *testing.T) {
	mockDB := &MockDatabase{}
	mockRows := &MockRows{}

	repo := NewTaskRepository(mockDB)

	tasks := []models.Task{
		{
			ID:          "task-1",
			ProjectID:   "project-123",
			Source:      "cursor",
			Title:       "Task 1",
			Description: "First task",
			Status:      "pending",
			Priority:    "high",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     1,
		},
		{
			ID:          "task-2",
			ProjectID:   "project-123",
			Source:      "manual",
			Title:       "Task 2",
			Description: "Second task",
			Status:      "completed",
			Priority:    "low",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     1,
		},
	}

	mockRows.rows = tasks

	ctx := context.Background()
	req := models.ListTasksRequest{Limit: 10, Offset: 0}

	mockDB.On("Query", ctx, mock.AnythingOfType("string"), "project-123", 10, 0).
		Return(mockRows, nil)

	mockDB.On("QueryRow", ctx, mock.AnythingOfType("string"), "project-123").
		Return(&MockRow{}) // Mock for count query

	resultTasks, total, err := repo.FindByProjectID(ctx, "project-123", req)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(resultTasks))
	assert.Equal(t, int64(0), total) // Mock returns 0 for count
	assert.Equal(t, "task-1", resultTasks[0].ID)
	assert.Equal(t, "task-2", resultTasks[1].ID)

	mockDB.AssertExpectations(t)
}

func TestTaskRepository_Delete(t *testing.T) {
	mockDB := &MockDatabase{}
	mockResult := &MockResult{}

	repo := NewTaskRepository(mockDB)

	ctx := context.Background()

	mockDB.On("Exec", ctx, mock.AnythingOfType("string"), "task-123").
		Return(mockResult, nil)

	err := repo.Delete(ctx, "task-123")

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestTaskRepository_SaveDependency(t *testing.T) {
	mockDB := &MockDatabase{}
	mockResult := &MockResult{}

	repo := NewTaskRepository(mockDB)

	dependency := &models.TaskDependency{
		ID:              "dep-123",
		TaskID:          "task-456",
		DependsOnTaskID: "task-789",
		DependencyType:  "explicit",
		Confidence:      0.9,
		CreatedAt:       time.Now(),
	}

	ctx := context.Background()

	mockDB.On("Exec", ctx, mock.AnythingOfType("string"),
		dependency.ID, dependency.TaskID, dependency.DependsOnTaskID,
		dependency.DependencyType, dependency.Confidence, dependency.CreatedAt).
		Return(mockResult, nil)

	err := repo.SaveDependency(ctx, dependency)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestTaskRepository_SaveVerification(t *testing.T) {
	mockDB := &MockDatabase{}
	mockResult := &MockResult{}

	repo := NewTaskRepository(mockDB)

	verification := &models.TaskVerification{
		ID:               "ver-123",
		TaskID:           "task-456",
		VerificationType: "code_existence",
		Status:           "verified",
		Confidence:       0.95,
		Evidence:         map[string]interface{}{"files_found": []string{"main.go"}},
		RetryCount:       0,
		VerifiedAt:       &time.Time{},
		CreatedAt:        time.Now(),
	}

	ctx := context.Background()

	mockDB.On("Exec", ctx, mock.AnythingOfType("string"),
		verification.ID, verification.TaskID, verification.VerificationType,
		verification.Status, verification.Confidence, verification.Evidence,
		verification.RetryCount, verification.VerifiedAt, verification.CreatedAt).
		Return(mockResult, nil)

	err := repo.SaveVerification(ctx, verification)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}
