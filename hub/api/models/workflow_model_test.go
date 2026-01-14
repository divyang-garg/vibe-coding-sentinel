// Package models - Workflow model tests
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import (
	"testing"
)

// TestWorkflowStatus tests WorkflowStatus enum functionality
func TestWorkflowStatus_IsValid(t *testing.T) {
	tests := []struct {
		status WorkflowStatus
		want   bool
	}{
		{WorkflowStatusPending, true},
		{WorkflowStatusRunning, true},
		{WorkflowStatusCompleted, true},
		{WorkflowStatusFailed, true},
		{WorkflowStatusCancelled, true},
		{WorkflowStatus("invalid"), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.want {
			t.Errorf("WorkflowStatus.IsValid() = %v, want %v", got, tt.want)
		}
	}
}

func TestWorkflowStatus_MarshalJSON(t *testing.T) {
	status := WorkflowStatusRunning
	data, err := status.MarshalJSON()
	if err != nil {
		t.Fatalf("WorkflowStatus.MarshalJSON() error = %v", err)
	}
	if string(data) != `"running"` {
		t.Errorf("WorkflowStatus.MarshalJSON() = %v, want %v", string(data), `"running"`)
	}
}

func TestWorkflowStatus_UnmarshalJSON(t *testing.T) {
	var status WorkflowStatus
	data := []byte(`"completed"`)
	err := status.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("WorkflowStatus.UnmarshalJSON() error = %v", err)
	}
	if status != WorkflowStatusCompleted {
		t.Errorf("WorkflowStatus.UnmarshalJSON() = %v, want %v", status, WorkflowStatusCompleted)
	}
}

// TestStepStatus tests StepStatus enum functionality
func TestStepStatus_IsValid(t *testing.T) {
	tests := []struct {
		status StepStatus
		want   bool
	}{
		{StepStatusPending, true},
		{StepStatusRunning, true},
		{StepStatusCompleted, true},
		{StepStatusFailed, true},
		{StepStatusSkipped, true},
		{StepStatus("invalid"), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.want {
			t.Errorf("StepStatus.IsValid() = %v, want %v", got, tt.want)
		}
	}
}

func TestStepStatus_MarshalJSON(t *testing.T) {
	status := StepStatusCompleted
	data, err := status.MarshalJSON()
	if err != nil {
		t.Fatalf("StepStatus.MarshalJSON() error = %v", err)
	}
	if string(data) != `"completed"` {
		t.Errorf("StepStatus.MarshalJSON() = %v, want %v", string(data), `"completed"`)
	}
}

func TestStepStatus_UnmarshalJSON(t *testing.T) {
	var status StepStatus
	data := []byte(`"failed"`)
	err := status.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("StepStatus.UnmarshalJSON() error = %v", err)
	}
	if status != StepStatusFailed {
		t.Errorf("StepStatus.UnmarshalJSON() = %v, want %v", status, StepStatusFailed)
	}
}

// TestAgentStatus tests AgentStatus enum functionality
func TestAgentStatus_IsValid(t *testing.T) {
	tests := []struct {
		status AgentStatus
		want   bool
	}{
		{AgentStatusActive, true},
		{AgentStatusInactive, true},
		{AgentStatusDisconnected, true},
		{AgentStatus("invalid"), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.want {
			t.Errorf("AgentStatus.IsValid() = %v, want %v", got, tt.want)
		}
	}
}

// TestWorkflowStatus_JSONSerialization tests WorkflowStatus JSON marshaling/unmarshaling
func TestWorkflowStatus_JSONSerialization(t *testing.T) {
	status := WorkflowStatusRunning

	// Test MarshalJSON
	data, err := status.MarshalJSON()
	if err != nil {
		t.Fatalf("WorkflowStatus.MarshalJSON() error = %v", err)
	}
	if string(data) != `"running"` {
		t.Errorf("WorkflowStatus.MarshalJSON() = %v, want %v", string(data), `"running"`)
	}

	// Test UnmarshalJSON
	var unmarshaled WorkflowStatus
	err = unmarshaled.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("WorkflowStatus.UnmarshalJSON() error = %v", err)
	}
	if unmarshaled != WorkflowStatusRunning {
		t.Errorf("Unmarshaled status = %v, want %v", unmarshaled, WorkflowStatusRunning)
	}
}

// TestStepStatus_JSONSerialization tests StepStatus JSON marshaling/unmarshaling
func TestStepStatus_JSONSerialization(t *testing.T) {
	status := StepStatusCompleted

	// Test MarshalJSON
	data, err := status.MarshalJSON()
	if err != nil {
		t.Fatalf("StepStatus.MarshalJSON() error = %v", err)
	}
	if string(data) != `"completed"` {
		t.Errorf("StepStatus.MarshalJSON() = %v, want %v", string(data), `"completed"`)
	}

	// Test UnmarshalJSON
	var unmarshaled StepStatus
	err = unmarshaled.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("StepStatus.UnmarshalJSON() error = %v", err)
	}
	if unmarshaled != StepStatusCompleted {
		t.Errorf("Unmarshaled status = %v, want %v", unmarshaled, StepStatusCompleted)
	}
}

// TestAgentStatus_JSONSerialization tests AgentStatus JSON marshaling/unmarshaling
func TestAgentStatus_JSONSerialization(t *testing.T) {
	status := AgentStatusActive

	// Test MarshalJSON
	data, err := status.MarshalJSON()
	if err != nil {
		t.Fatalf("AgentStatus.MarshalJSON() error = %v", err)
	}
	if string(data) != `"active"` {
		t.Errorf("AgentStatus.MarshalJSON() = %v, want %v", string(data), `"active"`)
	}

	// Test UnmarshalJSON
	var unmarshaled AgentStatus
	err = unmarshaled.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("AgentStatus.UnmarshalJSON() error = %v", err)
	}
	if unmarshaled != AgentStatusActive {
		t.Errorf("Unmarshaled status = %v, want %v", unmarshaled, AgentStatusActive)
	}
}
