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
	// Set values through pointers
	*dest[0].(*string) = task.ID
	*dest[1].(*string) = task.ProjectID
	*dest[2].(*string) = task.Source
	*dest[3].(*string) = task.Title
	*dest[4].(*string) = task.Description
	*dest[5].(*string) = task.FilePath
	if task.LineNumber != nil {
		if dest[6].(**int) != nil {
			*dest[6].(**int) = task.LineNumber
		}
	}
	*dest[7].(*models.TaskStatus) = task.Status
	*dest[8].(*models.TaskPriority) = task.Priority
	if task.AssignedTo != nil {
		if dest[9].(**string) != nil {
			*dest[9].(**string) = task.AssignedTo
		}
	}
	if task.EstimatedEffort != nil {
		if dest[10].(**int) != nil {
			*dest[10].(**int) = task.EstimatedEffort
		}
	}
	if task.ActualEffort != nil {
		if dest[11].(**int) != nil {
			*dest[11].(**int) = task.ActualEffort
		}
	}
	*dest[12].(*float64) = task.VerificationConfidence
	*dest[13].(*time.Time) = task.CreatedAt
	*dest[14].(*time.Time) = task.UpdatedAt
	if task.CompletedAt != nil {
		if dest[15].(**time.Time) != nil {
			*dest[15].(**time.Time) = task.CompletedAt
		}
	}
	if task.VerifiedAt != nil {
		if dest[16].(**time.Time) != nil {
			*dest[16].(**time.Time) = task.VerifiedAt
		}
	}
	if task.ArchivedAt != nil {
		if dest[17].(**time.Time) != nil {
			*dest[17].(**time.Time) = task.ArchivedAt
		}
	}
	*dest[18].(*int) = task.Version

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
		task.FilePath, task.LineNumber, string(task.Status), string(task.Priority), task.AssignedTo,
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
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Run(func(args mock.Arguments) {
		// Simulate scanning the task data by setting values through the pointers
		// Non-nullable fields
		*args[0].(*string) = expectedTask.ID
		*args[1].(*string) = expectedTask.ProjectID
		*args[2].(*string) = expectedTask.Source
		*args[3].(*string) = expectedTask.Title
		*args[4].(*string) = expectedTask.Description
		
		// FilePath is a string (not nullable)
		*args[5].(*string) = expectedTask.FilePath
		
		// Nullable int field (LineNumber) - args[6] is **int
		if expectedTask.LineNumber != nil {
			lineNumPtr := args[6].(**int)
			if lineNumPtr != nil {
				*lineNumPtr = expectedTask.LineNumber
			}
		}
		
		*args[7].(*models.TaskStatus) = expectedTask.Status
		*args[8].(*models.TaskPriority) = expectedTask.Priority
		
		// Nullable string field (AssignedTo) - args[9] is **string
		if expectedTask.AssignedTo != nil {
			assignedToPtr := args[9].(**string)
			if assignedToPtr != nil {
				*assignedToPtr = expectedTask.AssignedTo
			}
		}
		
		// Nullable int field (EstimatedEffort) - args[10] is **int
		if expectedTask.EstimatedEffort != nil {
			estEffortPtr := args[10].(**int)
			if estEffortPtr != nil {
				*estEffortPtr = expectedTask.EstimatedEffort
			}
		}
		
		// Nullable int field (ActualEffort) - args[11] is **int
		if expectedTask.ActualEffort != nil {
			actEffortPtr := args[11].(**int)
			if actEffortPtr != nil {
				*actEffortPtr = expectedTask.ActualEffort
			}
		}
		
		*args[12].(*float64) = expectedTask.VerificationConfidence
		*args[13].(*time.Time) = expectedTask.CreatedAt
		*args[14].(*time.Time) = expectedTask.UpdatedAt
		
		// Nullable time fields (completedAt, verifiedAt, archivedAt) - args[15], [16], [17] are **time.Time
		// These can remain nil
		
		*args[18].(*int) = expectedTask.Version
	}).Return(nil)

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

	// Mock for count query
	countRow := &MockRow{}
	countRow.On("Scan", mock.AnythingOfType("*int")).Run(func(args mock.Arguments) {
		*args[0].(*int) = 2 // Total count
	}).Return(nil)
	mockDB.On("QueryRow", ctx, mock.AnythingOfType("string"), "project-123").
		Return(countRow)

	resultTasks, total, err := repo.FindByProjectID(ctx, "project-123", req)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(resultTasks))
	assert.Equal(t, 2, total)
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
