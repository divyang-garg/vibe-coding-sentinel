// Package models - Task model tests
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTaskStatus tests TaskStatus enum functionality
func TestTaskStatus_String(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   string
	}{
		{TaskStatusPending, "pending"},
		{TaskStatusInProgress, "in_progress"},
		{TaskStatusCompleted, "completed"},
		{TaskStatusBlocked, "blocked"},
		{TaskStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("TaskStatus.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestTaskStatus_IsValid(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   bool
	}{
		{TaskStatusPending, true},
		{TaskStatusInProgress, true},
		{TaskStatusCompleted, true},
		{TaskStatusBlocked, true},
		{TaskStatusCancelled, true},
		{TaskStatus("invalid"), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.want {
			t.Errorf("TaskStatus.IsValid() = %v, want %v", got, tt.want)
		}
	}
}

func TestTaskStatus_MarshalJSON(t *testing.T) {
	status := TaskStatusCompleted
	data, err := status.MarshalJSON()
	if err != nil {
		t.Fatalf("TaskStatus.MarshalJSON() error = %v", err)
	}
	if string(data) != `"completed"` {
		t.Errorf("TaskStatus.MarshalJSON() = %v, want %v", string(data), `"completed"`)
	}
}

func TestTaskStatus_UnmarshalJSON(t *testing.T) {
	var status TaskStatus
	data := []byte(`"in_progress"`)
	err := status.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("TaskStatus.UnmarshalJSON() error = %v", err)
	}
	if status != TaskStatusInProgress {
		t.Errorf("TaskStatus.UnmarshalJSON() = %v, want %v", status, TaskStatusInProgress)
	}
}

// TestTaskPriority tests TaskPriority enum functionality
func TestTaskPriority_IsValid(t *testing.T) {
	tests := []struct {
		priority TaskPriority
		want     bool
	}{
		{TaskPriorityLow, true},
		{TaskPriorityMedium, true},
		{TaskPriorityHigh, true},
		{TaskPriorityCritical, true},
		{TaskPriority("invalid"), false},
	}

	for _, tt := range tests {
		if got := tt.priority.IsValid(); got != tt.want {
			t.Errorf("TaskPriority.IsValid() = %v, want %v", got, tt.want)
		}
	}
}

func TestTaskPriority_MarshalJSON(t *testing.T) {
	priority := TaskPriorityHigh
	data, err := priority.MarshalJSON()
	if err != nil {
		t.Fatalf("TaskPriority.MarshalJSON() error = %v", err)
	}
	if string(data) != `"high"` {
		t.Errorf("TaskPriority.MarshalJSON() = %v, want %v", string(data), `"high"`)
	}
}

func TestTaskPriority_UnmarshalJSON(t *testing.T) {
	var priority TaskPriority
	data := []byte(`"critical"`)
	err := priority.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("TaskPriority.UnmarshalJSON() error = %v", err)
	}
	if priority != TaskPriorityCritical {
		t.Errorf("TaskPriority.UnmarshalJSON() = %v, want %v", priority, TaskPriorityCritical)
	}
}

// TestTask_JSONSerialization tests Task model JSON marshaling
func TestTask_JSONSerialization(t *testing.T) {
	task := Task{
		ID:        "task-123",
		ProjectID: "project-abc",
		Source:    "manual",
		Title:     "Implement feature X",
		Status:    TaskStatusPending,
		Priority:  TaskPriorityHigh,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Version:   1,
	}

	// Test MarshalJSON
	data, err := json.Marshal(task)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"status":"pending"`)
	assert.Contains(t, string(data), `"priority":"high"`)

	// Test UnmarshalJSON
	var unmarshaledTask Task
	err = json.Unmarshal(data, &unmarshaledTask)
	assert.NoError(t, err)
	assert.Equal(t, TaskStatusPending, unmarshaledTask.Status)
	assert.Equal(t, TaskPriorityHigh, unmarshaledTask.Priority)
}

// TestValidateTask tests Task model validation
func TestValidateTask(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: Task{
				ID:        "task-123",
				ProjectID: "proj-1",
				Title:     "Test task",
				Status:    TaskStatusPending,
				Priority:  TaskPriorityMedium,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			task: Task{
				ProjectID: "proj-1",
				Title:     "Test task",
				Status:    TaskStatusPending,
				Priority:  TaskPriorityMedium,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			task: Task{
				ID:        "task-123",
				ProjectID: "proj-1",
				Title:     "Test task",
				Status:    "invalid",
				Priority:  TaskPriorityMedium,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTask(&tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
