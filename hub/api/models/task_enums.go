// Package models - Task-related enums and types
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import (
	"encoding/json"
	"fmt"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// String returns the string representation of TaskStatus
func (s TaskStatus) String() string {
	return string(s)
}

// IsValid checks if the TaskStatus is valid
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted, TaskStatusBlocked, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler
func (s TaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements json.Unmarshaler
func (s *TaskStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = TaskStatus(str)
	if !s.IsValid() {
		return fmt.Errorf("invalid task status: %s", str)
	}
	return nil
}

// TaskPriority represents the priority of a task
type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityMedium   TaskPriority = "medium"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

// String returns the string representation of TaskPriority
func (p TaskPriority) String() string {
	return string(p)
}

// IsValid checks if the TaskPriority is valid
func (p TaskPriority) IsValid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityCritical:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler
func (p TaskPriority) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(p))
}

// UnmarshalJSON implements json.Unmarshaler
func (p *TaskPriority) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*p = TaskPriority(str)
	if !p.IsValid() {
		return fmt.Errorf("invalid task priority: %s", str)
	}
	return nil
}

// VerificationStatus represents the status of a task verification
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusFailed   VerificationStatus = "failed"
)

// String returns the string representation of VerificationStatus
func (s VerificationStatus) String() string {
	return string(s)
}

// IsValid checks if the VerificationStatus is valid
func (s VerificationStatus) IsValid() bool {
	switch s {
	case VerificationStatusPending, VerificationStatusVerified, VerificationStatusFailed:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler
func (s VerificationStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements json.Unmarshaler
func (s *VerificationStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = VerificationStatus(str)
	if !s.IsValid() {
		return fmt.Errorf("invalid verification status: %s", str)
	}
	return nil
}
