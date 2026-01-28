// Package models provides additional tests for task models
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTask_CanTransitionTo_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		current   TaskStatus
		newStatus TaskStatus
		expected  bool
	}{
		{
			name:      "pending to in_progress",
			current:   TaskStatusPending,
			newStatus: TaskStatusInProgress,
			expected:  true,
		},
		{
			name:      "pending to cancelled",
			current:   TaskStatusPending,
			newStatus: TaskStatusCancelled,
			expected:  true,
		},
		{
			name:      "pending to completed (invalid)",
			current:   TaskStatusPending,
			newStatus: TaskStatusCompleted,
			expected:  false,
		},
		{
			name:      "in_progress to completed",
			current:   TaskStatusInProgress,
			newStatus: TaskStatusCompleted,
			expected:  true,
		},
		{
			name:      "in_progress to blocked",
			current:   TaskStatusInProgress,
			newStatus: TaskStatusBlocked,
			expected:  true,
		},
		{
			name:      "in_progress to cancelled",
			current:   TaskStatusInProgress,
			newStatus: TaskStatusCancelled,
			expected:  true,
		},
		{
			name:      "blocked to in_progress",
			current:   TaskStatusBlocked,
			newStatus: TaskStatusInProgress,
			expected:  true,
		},
		{
			name:      "blocked to cancelled",
			current:   TaskStatusBlocked,
			newStatus: TaskStatusCancelled,
			expected:  true,
		},
		{
			name:      "completed to in_progress (terminal state)",
			current:   TaskStatusCompleted,
			newStatus: TaskStatusInProgress,
			expected:  false,
		},
		{
			name:      "cancelled to in_progress (terminal state)",
			current:   TaskStatusCancelled,
			newStatus: TaskStatusInProgress,
			expected:  false,
		},
		{
			name:      "unknown status",
			current:   TaskStatus("unknown"),
			newStatus: TaskStatusInProgress,
			expected:  false,
		},
		{
			name:      "pending to blocked (invalid)",
			current:   TaskStatusPending,
			newStatus: TaskStatusBlocked,
			expected:  false,
		},
		{
			name:      "blocked to completed (invalid)",
			current:   TaskStatusBlocked,
			newStatus: TaskStatusCompleted,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.current}
			result := task.CanTransitionTo(tt.newStatus)
			assert.Equal(t, tt.expected, result, "CanTransitionTo(%s -> %s) = %v, want %v", tt.current, tt.newStatus, result, tt.expected)
		})
	}
}

func TestTask_Structure(t *testing.T) {
	t.Run("creates task with all fields", func(t *testing.T) {
		now := time.Now()
		lineNum := 42
		assignedTo := "user-123"
		effort := 5

		task := &Task{
			ID:                     1,
			ProjectID:              "proj-123",
			Source:                 "scanner",
			Title:                  "Test Task",
			Description:            "Test Description",
			FilePath:               "test.go",
			LineNumber:             &lineNum,
			Status:                 TaskStatusInProgress,
			Priority:               TaskPriorityHigh,
			AssignedTo:             &assignedTo,
			EstimatedEffort:        &effort,
			ActualEffort:           &effort,
			Tags:                   []string{"bug", "critical"},
			VerificationConfidence: 0.95,
			CreatedAt:              now,
			UpdatedAt:              now,
			CompletedAt:            &now,
			VerifiedAt:             &now,
			ArchivedAt:             nil,
			Version:                1,
		}

		assert.Equal(t, 1, task.ID)
		assert.Equal(t, "proj-123", task.ProjectID)
		assert.Equal(t, "Test Task", task.Title)
		assert.Equal(t, TaskStatusInProgress, task.Status)
		assert.Equal(t, TaskPriorityHigh, task.Priority)
		assert.NotNil(t, task.LineNumber)
		assert.Equal(t, 42, *task.LineNumber)
	})
}
