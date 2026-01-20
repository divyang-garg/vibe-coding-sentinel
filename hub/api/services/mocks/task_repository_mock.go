// Package mocks provides mock implementations for testing
// Complies with CODING_STANDARDS.md: Test utilities max 250 lines
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"sentinel-hub-api/models"
)

// MockTaskRepository implements TaskRepository for testing
type MockTaskRepository struct {
	mock.Mock
	tasks         map[string]*models.Task
	dependencies  map[string]*models.TaskDependency
	verifications map[string]*models.TaskVerification
}

// NewMockTaskRepository creates a new mock task repository
func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks:         make(map[string]*models.Task),
		dependencies:  make(map[string]*models.TaskDependency),
		verifications: make(map[string]*models.TaskVerification),
	}
}

// Save saves a task
func (m *MockTaskRepository) Save(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	if args.Error(0) == nil && task != nil {
		m.tasks[task.ID] = task
	}
	return args.Error(0)
}

// FindByID finds a task by ID
func (m *MockTaskRepository) FindByID(ctx context.Context, id string) (*models.Task, error) {
	args := m.Called(ctx, id)
	if task, ok := m.tasks[id]; ok {
		return task, args.Error(1)
	}
	if args.Get(0) != nil {
		return args.Get(0).(*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

// Update updates a task
func (m *MockTaskRepository) Update(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	if args.Error(0) == nil && task != nil {
		m.tasks[task.ID] = task
	}
	return args.Error(0)
}

// Delete deletes a task
func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		delete(m.tasks, id)
	}
	return args.Error(0)
}

// FindByProjectID finds tasks by project ID
func (m *MockTaskRepository) FindByProjectID(ctx context.Context, projectID string, req models.ListTasksRequest) ([]models.Task, int, error) {
	args := m.Called(ctx, projectID, req)
	var tasks []models.Task
	for _, task := range m.tasks {
		if task.ProjectID == projectID {
			tasks = append(tasks, *task)
		}
	}
	if args.Get(0) != nil {
		tasks = args.Get(0).([]models.Task)
	}
	return tasks, args.Int(1), args.Error(2)
}

// SaveDependency saves a task dependency
func (m *MockTaskRepository) SaveDependency(ctx context.Context, dependency *models.TaskDependency) error {
	args := m.Called(ctx, dependency)
	if args.Error(0) == nil && dependency != nil {
		depID := dependency.ID
		if depID == "" {
			depID = dependency.TaskID + "_" + dependency.DependsOnTaskID
		}
		m.dependencies[depID] = dependency
	}
	return args.Error(0)
}

// FindDependencies finds dependencies for a task
func (m *MockTaskRepository) FindDependencies(ctx context.Context, taskID string) ([]models.TaskDependency, error) {
	args := m.Called(ctx, taskID)
	var deps []models.TaskDependency
	for _, dep := range m.dependencies {
		if dep.TaskID == taskID {
			deps = append(deps, *dep)
		}
	}
	if args.Get(0) != nil {
		deps = args.Get(0).([]models.TaskDependency)
	}
	return deps, args.Error(1)
}

// FindDependents finds tasks that depend on the given task
func (m *MockTaskRepository) FindDependents(ctx context.Context, taskID string) ([]models.TaskDependency, error) {
	args := m.Called(ctx, taskID)
	var deps []models.TaskDependency
	for _, dep := range m.dependencies {
		if dep.DependsOnTaskID == taskID {
			deps = append(deps, *dep)
		}
	}
	if args.Get(0) != nil {
		deps = args.Get(0).([]models.TaskDependency)
	}
	return deps, args.Error(1)
}

// DeleteDependency deletes a dependency
func (m *MockTaskRepository) DeleteDependency(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		delete(m.dependencies, id)
	}
	return args.Error(0)
}

// SaveVerification saves a task verification
func (m *MockTaskRepository) SaveVerification(ctx context.Context, verification *models.TaskVerification) error {
	args := m.Called(ctx, verification)
	if args.Error(0) == nil && verification != nil {
		m.verifications[verification.ID] = verification
	}
	return args.Error(0)
}
