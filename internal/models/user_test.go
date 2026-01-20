// Package models provides unit tests for data models
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		wantErr  bool
		errField string
	}{
		{
			name: "valid user",
			user: User{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "hashedpassword",
				Role:     RoleUser,
			},
			wantErr: false,
		},
		{
			name: "missing email",
			user: User{
				Name:     "Test User",
				Password: "hashedpassword",
				Role:     RoleUser,
			},
			wantErr:  true,
			errField: "email",
		},
		{
			name: "missing name",
			user: User{
				Email:    "test@example.com",
				Password: "hashedpassword",
				Role:     RoleUser,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "empty email",
			user: User{
				Email:    "",
				Name:     "Test User",
				Password: "hashedpassword",
				Role:     RoleUser,
			},
			wantErr:  true,
			errField: "email",
		},
		{
			name: "empty name",
			user: User{
				Email:    "test@example.com",
				Name:     "",
				Password: "hashedpassword",
				Role:     RoleUser,
			},
			wantErr:  true,
			errField: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				var validationErr *ValidationError
				assert.ErrorAs(t, err, &validationErr)
				if tt.errField != "" {
					assert.Equal(t, tt.errField, validationErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTask_Validate(t *testing.T) {
	tests := []struct {
		name     string
		task     Task
		wantErr  bool
		errField string
	}{
		{
			name: "valid task",
			task: Task{
				ProjectID: "project-123",
				Title:     "Implement user authentication",
				Status:    TaskStatusPending,
				Priority:  TaskPriorityMedium,
			},
			wantErr: false,
		},
		{
			name: "missing project_id",
			task: Task{
				Title:    "Implement user authentication",
				Status:   TaskStatusPending,
				Priority: TaskPriorityMedium,
			},
			wantErr:  true,
			errField: "project_id",
		},
		{
			name: "missing title",
			task: Task{
				ProjectID: "project-123",
				Status:    TaskStatusPending,
				Priority:  TaskPriorityMedium,
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "empty project_id",
			task: Task{
				ProjectID: "",
				Title:     "Implement user authentication",
				Status:    TaskStatusPending,
				Priority:  TaskPriorityMedium,
			},
			wantErr:  true,
			errField: "project_id",
		},
		{
			name: "empty title",
			task: Task{
				ProjectID: "project-123",
				Title:     "",
				Status:    TaskStatusPending,
				Priority:  TaskPriorityMedium,
			},
			wantErr:  true,
			errField: "title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				var validationErr *ValidationError
				assert.ErrorAs(t, err, &validationErr)
				if tt.errField != "" {
					assert.Equal(t, tt.errField, validationErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTask_IsCompleted(t *testing.T) {
	tests := []struct {
		name string
		task Task
		want bool
	}{
		{
			name: "completed task",
			task: Task{Status: TaskStatusCompleted},
			want: true,
		},
		{
			name: "pending task",
			task: Task{Status: TaskStatusPending},
			want: false,
		},
		{
			name: "in progress task",
			task: Task{Status: TaskStatusInProgress},
			want: false,
		},
		{
			name: "blocked task",
			task: Task{Status: TaskStatusBlocked},
			want: false,
		},
		{
			name: "cancelled task",
			task: Task{Status: TaskStatusCancelled},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.task.IsCompleted())
		})
	}
}

func TestTask_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		current  TaskStatus
		next     TaskStatus
		expected bool
	}{
		// From Pending
		{"pending to in_progress", TaskStatusPending, TaskStatusInProgress, true},
		{"pending to cancelled", TaskStatusPending, TaskStatusCancelled, true},
		{"pending to completed", TaskStatusPending, TaskStatusCompleted, false},
		{"pending to blocked", TaskStatusPending, TaskStatusBlocked, false},

		// From In Progress
		{"in_progress to completed", TaskStatusInProgress, TaskStatusCompleted, true},
		{"in_progress to blocked", TaskStatusInProgress, TaskStatusBlocked, true},
		{"in_progress to cancelled", TaskStatusInProgress, TaskStatusCancelled, true},
		{"in_progress to pending", TaskStatusInProgress, TaskStatusPending, false},

		// From Blocked
		{"blocked to in_progress", TaskStatusBlocked, TaskStatusInProgress, true},
		{"blocked to cancelled", TaskStatusBlocked, TaskStatusCancelled, true},
		{"blocked to completed", TaskStatusBlocked, TaskStatusCompleted, false},

		// Terminal states (cannot transition)
		{"completed to any", TaskStatusCompleted, TaskStatusPending, false},
		{"cancelled to any", TaskStatusCancelled, TaskStatusInProgress, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{Status: tt.current}
			assert.Equal(t, tt.expected, task.CanTransitionTo(tt.next))
		})
	}
}

func TestErrorTypes(t *testing.T) {
	t.Run("ValidationError", func(t *testing.T) {
		err := &ValidationError{
			Field:   "email",
			Value:   "invalid-email",
			Message: "invalid email format",
		}

		assert.Equal(t, "validation failed for field 'email': invalid email format", err.Error())
		assert.Equal(t, "email", err.Field)
		assert.Equal(t, "invalid email format", err.Message)
	})

	t.Run("NotFoundError", func(t *testing.T) {
		err := &NotFoundError{
			Resource: "user",
			ID:       123,
		}

		assert.Equal(t, "user with id 123 not found", err.Error())
	})

	t.Run("NotFoundError without ID", func(t *testing.T) {
		err := &NotFoundError{
			Resource: "user",
		}

		assert.Equal(t, "user not found", err.Error())
	})

	t.Run("AuthenticationError", func(t *testing.T) {
		err := &AuthenticationError{
			Message: "invalid credentials",
		}

		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("AuthorizationError", func(t *testing.T) {
		err := &AuthorizationError{
			Message: "access denied",
		}

		assert.Equal(t, "access denied", err.Error())
	})

	t.Run("NotImplementedError", func(t *testing.T) {
		err := &NotImplementedError{
			Feature: "advanced analytics",
		}

		assert.Equal(t, "feature not implemented: advanced analytics", err.Error())
	})

	t.Run("RateLimitError", func(t *testing.T) {
		resetTime := time.Now()
		err := &RateLimitError{
			Message:    "rate limit exceeded",
			RetryAfter: 60,
			ResetTime:  resetTime,
		}

		assert.Equal(t, "rate limit exceeded", err.Error())
		assert.Equal(t, 60, err.RetryAfter)
		assert.Equal(t, resetTime, err.ResetTime)
	})
}
