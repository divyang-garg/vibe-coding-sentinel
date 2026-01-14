// Package models contains organization-related data models.
// This file defines all organization domain entities following the data-only principle.
package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// AgentStatus represents the status of a registered agent
type AgentStatus string

const (
	AgentStatusActive       AgentStatus = "active"
	AgentStatusInactive     AgentStatus = "inactive"
	AgentStatusDisconnected AgentStatus = "disconnected"
)

// String returns the string representation of AgentStatus
func (s AgentStatus) String() string {
	return string(s)
}

// IsValid checks if the AgentStatus is valid
func (s AgentStatus) IsValid() bool {
	switch s {
	case AgentStatusActive, AgentStatusInactive, AgentStatusDisconnected:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler
func (s AgentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements json.Unmarshaler
func (s *AgentStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = AgentStatus(str)
	if !s.IsValid() {
		return fmt.Errorf("invalid agent status: %s", str)
	}
	return nil
}

// Organization represents a customer organization
type Organization struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Project represents a project within an organization
type Project struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Team represents a team within an organization
type Team struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	LeadUserID  string    `json:"lead_user_id,omitempty"`
	Settings    JSONMap   `json:"settings,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TeamMember represents a member of a team
type TeamMember struct {
	ID         string     `json:"id"`
	TeamID     string     `json:"team_id"`
	UserID     string     `json:"user_id"`
	Role       string     `json:"role"` // 'member', 'lead', 'admin'
	JoinedAt   time.Time  `json:"joined_at"`
	LastActive *time.Time `json:"last_active,omitempty"`
}

// RegisteredAgent represents an agent registered with the hub
type RegisteredAgent struct {
	ID        string      `json:"id"`
	OrgID     string      `json:"org_id"`
	ProjectID string      `json:"project_id"`
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	LastSeen  time.Time   `json:"last_seen"`
	Status    AgentStatus `json:"status"`
}

// CreateOrganizationRequest represents a request to create an organization
type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateOrganizationRequest represents a request to update an organization
type UpdateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name string `json:"name"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name string `json:"name"`
}
