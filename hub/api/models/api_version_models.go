// Package models - API versioning data models
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import (
	"time"
)

// APIVersion represents an API version definition
type APIVersion struct {
	ID          string    `json:"id" validate:"required"`
	Version     string    `json:"version" validate:"required"`
	Description string    `json:"description"`
	ReleasedAt  time.Time `json:"released_at"`
	IsActive    bool      `json:"is_active"`
	Changelog   []string  `json:"changelog,omitempty"`
	Endpoints   []string  `json:"endpoints,omitempty"`
}

// VersionMigration represents a migration between API versions
type VersionMigration struct {
	ID          string                 `json:"id" validate:"required"`
	FromVersion string                 `json:"from_version" validate:"required"`
	ToVersion   string                 `json:"to_version" validate:"required"`
	Description string                 `json:"description"`
	Changes     []VersionChange        `json:"changes"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// VersionChange represents a specific change in a version migration
type VersionChange struct {
	Type        string `json:"type"`
	Endpoint    string `json:"endpoint"`
	Description string `json:"description"`
	Breaking    bool   `json:"breaking"`
}
