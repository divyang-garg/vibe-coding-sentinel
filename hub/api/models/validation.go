// Package models provides validation utilities for model types.
package models

import (
	"fmt"
	"strings"
)

// ValidateUser validates a User model
func ValidateUser(u *User) error {
	if u.ID == "" {
		return fmt.Errorf("user ID is required")
	}
	if u.Email == "" {
		return fmt.Errorf("user email is required")
	}
	if !strings.Contains(u.Email, "@") {
		return fmt.Errorf("invalid email format")
	}
	if u.Name == "" {
		return fmt.Errorf("user name is required")
	}
	if u.Role != "" && u.Role != UserRoleAdmin && u.Role != UserRoleManager && u.Role != UserRoleDeveloper && u.Role != UserRoleViewer {
		return fmt.Errorf("invalid role: %s", u.Role)
	}
	return nil
}

// ValidateTask validates a Task model
func ValidateTask(t *Task) error {
	if t.ID == "" {
		return fmt.Errorf("task ID is required")
	}
	if t.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if t.Title == "" {
		return fmt.Errorf("task title is required")
	}
	if !t.Status.IsValid() {
		return fmt.Errorf("invalid task status: %s", t.Status)
	}
	if !t.Priority.IsValid() {
		return fmt.Errorf("invalid task priority: %s", t.Priority)
	}
	if t.VerificationConfidence < 0 || t.VerificationConfidence > 1 {
		return fmt.Errorf("verification confidence must be between 0 and 1")
	}
	return nil
}

// ValidateDocument validates a Document model
func ValidateDocument(d *Document) error {
	if d.ID == "" {
		return fmt.Errorf("document ID is required")
	}
	if d.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if d.Name == "" {
		return fmt.Errorf("document name is required")
	}
	if !d.Status.IsValid() {
		return fmt.Errorf("invalid document status: %s", d.Status)
	}
	if d.Size < 0 {
		return fmt.Errorf("document size cannot be negative")
	}
	if d.Progress < 0 || d.Progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}
	return nil
}

// ValidateOrganization validates an Organization model
func ValidateOrganization(o *Organization) error {
	if o.ID == "" {
		return fmt.Errorf("organization ID is required")
	}
	if o.Name == "" {
		return fmt.Errorf("organization name is required")
	}
	return nil
}

// ValidateProject validates a Project model
func ValidateProject(p *Project) error {
	if p.ID == "" {
		return fmt.Errorf("project ID is required")
	}
	if p.OrgID == "" {
		return fmt.Errorf("organization ID is required")
	}
	if p.Name == "" {
		return fmt.Errorf("project name is required")
	}
	return nil
}

// ValidateCreateUserRequest validates a CreateUserRequest
func ValidateCreateUserRequest(req CreateUserRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("invalid email format")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Role != UserRoleAdmin && req.Role != UserRoleManager && req.Role != UserRoleDeveloper && req.Role != UserRoleViewer {
		return fmt.Errorf("invalid role: %s", req.Role)
	}
	return nil
}

// ValidateUpdateUserRequest validates an UpdateUserRequest
func ValidateUpdateUserRequest(req UpdateUserRequest) error {
	if req.Email != nil && *req.Email != "" && !strings.Contains(*req.Email, "@") {
		return fmt.Errorf("invalid email format")
	}
	if req.Name != nil && *req.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if req.Role != nil && *req.Role != UserRoleAdmin && *req.Role != UserRoleManager && *req.Role != UserRoleDeveloper && *req.Role != UserRoleViewer {
		return fmt.Errorf("invalid role: %s", *req.Role)
	}
	return nil
}
