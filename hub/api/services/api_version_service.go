// Package services provides API versioning business logic.
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/models"
)

// APIVersionServiceImpl implements APIVersionService
type APIVersionServiceImpl struct {
	// In production, this would have repositories for version persistence
	// For now, we'll use in-memory storage for demonstration
	versions   map[string]*models.APIVersion
	migrations map[string]*models.VersionMigration
	nextID     int
}

// NewAPIVersionService creates a new API version service instance
func NewAPIVersionService() APIVersionService {
	return &APIVersionServiceImpl{
		versions:   make(map[string]*models.APIVersion),
		migrations: make(map[string]*models.VersionMigration),
		nextID:     1,
	}
}

// CreateAPIVersion creates a new API version
func (s *APIVersionServiceImpl) CreateAPIVersion(ctx context.Context, req models.APIVersion) (*models.APIVersion, error) {
	if req.Version == "" {
		return nil, fmt.Errorf("version is required")
	}

	// Check if version already exists
	for _, existing := range s.versions {
		if existing.Version == req.Version {
			return nil, fmt.Errorf("version %s already exists", req.Version)
		}
	}

	// Generate ID and set defaults
	req.ID = fmt.Sprintf("ver_%d", s.nextID)
	s.nextID++

	if req.ReleasedAt.IsZero() {
		req.ReleasedAt = time.Now()
	}

	s.versions[req.ID] = &req
	return &req, nil
}

// GetAPIVersion retrieves an API version by ID
func (s *APIVersionServiceImpl) GetAPIVersion(ctx context.Context, id string) (*models.APIVersion, error) {
	version, exists := s.versions[id]
	if !exists {
		return nil, fmt.Errorf("API version not found")
	}
	return version, nil
}

// ListAPIVersions retrieves all API versions
func (s *APIVersionServiceImpl) ListAPIVersions(ctx context.Context) ([]*models.APIVersion, error) {
	versions := make([]*models.APIVersion, 0, len(s.versions))
	for _, version := range s.versions {
		versions = append(versions, version)
	}
	return versions, nil
}

// GetVersionCompatibility checks compatibility between two versions
func (s *APIVersionServiceImpl) GetVersionCompatibility(ctx context.Context, fromVersion, toVersion string) (interface{}, error) {
	if fromVersion == "" || toVersion == "" {
		return nil, fmt.Errorf("both from_version and to_version are required")
	}

	// Find the versions
	var fromVer, toVer *models.APIVersion
	for _, v := range s.versions {
		if v.Version == fromVersion {
			fromVer = v
		}
		if v.Version == toVersion {
			toVer = v
		}
	}

	if fromVer == nil || toVer == nil {
		return nil, fmt.Errorf("one or both versions not found")
	}

	// Simple compatibility check (in production, this would be more sophisticated)
	compatibility := "full"
	if fromVersion != toVersion {
		compatibility = "partial" // Assume partial compatibility for different versions
	}

	// Check for migrations
	var migration *models.VersionMigration
	for _, m := range s.migrations {
		if m.FromVersion == fromVersion && m.ToVersion == toVersion {
			migration = m
			break
		}
	}

	return map[string]interface{}{
		"from_version":        fromVersion,
		"to_version":          toVersion,
		"compatibility":       compatibility,
		"breaking_changes":    []string{}, // Would be populated based on version analysis
		"migration_available": migration != nil,
		"recommendations":     []string{"Review API changes before upgrading"},
		"analyzed_at":         time.Now(),
	}, nil
}

// CreateVersionMigration creates a version migration
func (s *APIVersionServiceImpl) CreateVersionMigration(ctx context.Context, req models.VersionMigration) (*models.VersionMigration, error) {
	if req.FromVersion == "" || req.ToVersion == "" {
		return nil, fmt.Errorf("both from_version and to_version are required")
	}

	// Generate ID and set defaults
	req.ID = fmt.Sprintf("mig_%d", s.nextID)
	s.nextID++

	req.CreatedAt = time.Now()
	if req.Status == "" {
		req.Status = "planned"
	}

	s.migrations[req.ID] = &req
	return &req, nil
}
