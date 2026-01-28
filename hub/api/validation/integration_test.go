// Package validation - Integration tests for validation framework
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package validation

import (
	"testing"
)

func TestValidateCreateTaskRequest_Integration(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid task creation",
			data: map[string]interface{}{
				"title":  "Test Task",
				"status": "pending",
			},
			wantErr: false,
		},
		{
			name: "missing required title",
			data: map[string]interface{}{
				"status": "pending",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			data: map[string]interface{}{
				"title":  "Test Task",
				"status": "invalid_status",
			},
			wantErr: true,
		},
		{
			name: "title too long",
			data: map[string]interface{}{
				"title":  string(make([]byte, 501)), // 501 characters
				"status": "pending",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateTaskRequest(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateTaskRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreateProjectRequest_Integration(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid project creation",
			data: map[string]interface{}{
				"name": "Test Project",
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			data:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "invalid characters in name",
			data: map[string]interface{}{
				"name": "Test@Project#123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateProjectRequest(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateProjectRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNoSQLInjection_Integration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"safe input", "normal text", false},
		{"SQL injection", "'; DROP TABLE users; --", true},
		{"SELECT statement", "SELECT * FROM users", true},
		{"UNION attack", "UNION SELECT password FROM users", true},
		{"INSERT attack", "INSERT INTO users VALUES", true},
		{"UPDATE attack", "UPDATE users SET", true},
		{"DELETE attack", "DELETE FROM users", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoSQLInjection("field", tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNoSQLInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
