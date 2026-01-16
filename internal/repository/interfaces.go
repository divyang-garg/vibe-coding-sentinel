// Package repository defines data access interfaces
// Complies with CODING_STANDARDS.md: Interface definitions
package repository

import (
	"context"

	"github.com/divyang-garg/sentinel-hub-api/internal/models"
)

// UserRepository defines user data access methods
type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

// TaskRepository defines task data access methods
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) (*models.Task, error)
	GetByID(ctx context.Context, id int) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter *TaskFilter) ([]*models.Task, error)
	GetByProjectID(ctx context.Context, projectID string, limit, offset int) ([]*models.Task, error)
}

// TaskFilter represents filtering options for task queries
type TaskFilter struct {
	Status     *string
	Priority   *string
	AssignedTo *string
	ProjectID  *string
	Limit      int
	Offset     int
}
