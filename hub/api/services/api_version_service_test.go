// Package services provides unit tests for API version service.
// Complies with CODING_STANDARDS.md: Test files max 500 lines
package services

import (
	"context"
	"testing"
	"time"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
)

func TestAPIVersionServiceImpl_CreateAPIVersion(t *testing.T) {
	tests := []struct {
		name    string
		req     models.APIVersion
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid API version creation",
			req: models.APIVersion{
				Version:     "1.0.0",
				Description: "Initial API version",
				ReleasedAt:  time.Now(),
				IsActive:    true,
				Changelog:   []string{"Initial release"},
				Endpoints:   []string{"/api/v1/tasks", "/api/v1/users"},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			req: models.APIVersion{
				Description: "Test version",
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "duplicate version",
			req: models.APIVersion{
				Version:     "1.0.0",
				Description: "Duplicate version",
			},
			wantErr: true,
			errMsg:  "already exists",
		},
	}

	service := NewAPIVersionService()

	// Create first version for duplicate test
	firstVersion := models.APIVersion{
		Version:     "0.9.0",
		Description: "First version",
		ReleasedAt:  time.Now(),
		IsActive:    true,
	}
	_, err := service.CreateAPIVersion(context.Background(), firstVersion)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CreateAPIVersion(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Version, result.Version)
				assert.Equal(t, tt.req.Description, result.Description)
				assert.NotEmpty(t, result.ID)
			}
		})
	}
}

func TestAPIVersionServiceImpl_GetAPIVersion(t *testing.T) {
	service := NewAPIVersionService()

	// Create a version first
	version := models.APIVersion{
		Version:     "2.0.0",
		Description: "Test version",
		ReleasedAt:  time.Now(),
		IsActive:    true,
		Endpoints:   []string{"/api/v2/test"},
	}

	created, err := service.CreateAPIVersion(context.Background(), version)
	assert.NoError(t, err)

	// Test getting the version
	result, err := service.GetAPIVersion(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, "2.0.0", result.Version)
	assert.Equal(t, "Test version", result.Description)
	assert.True(t, result.IsActive)
}

func TestAPIVersionServiceImpl_GetAPIVersion_NotFound(t *testing.T) {
	service := NewAPIVersionService()

	result, err := service.GetAPIVersion(context.Background(), "non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestAPIVersionServiceImpl_ListAPIVersions(t *testing.T) {
	service := NewAPIVersionService()

	// Create multiple versions
	versions := []models.APIVersion{
		{
			Version:     "1.0.0",
			Description: "Version 1",
			ReleasedAt:  time.Now().Add(-24 * time.Hour),
		},
		{
			Version:     "1.1.0",
			Description: "Version 1.1",
			ReleasedAt:  time.Now().Add(-12 * time.Hour),
		},
		{
			Version:     "2.0.0",
			Description: "Version 2",
			ReleasedAt:  time.Now(),
		},
	}

	for _, v := range versions {
		_, err := service.CreateAPIVersion(context.Background(), v)
		assert.NoError(t, err)
	}

	// Test listing all versions
	result, err := service.ListAPIVersions(context.Background())
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Verify all versions are present
	versionMap := make(map[string]*models.APIVersion)
	for _, v := range result {
		versionMap[v.Version] = v
	}

	assert.Contains(t, versionMap, "1.0.0")
	assert.Contains(t, versionMap, "1.1.0")
	assert.Contains(t, versionMap, "2.0.0")

	assert.Equal(t, "Version 1", versionMap["1.0.0"].Description)
	assert.Equal(t, "Version 2", versionMap["2.0.0"].Description)
}

func TestAPIVersionServiceImpl_GetVersionCompatibility(t *testing.T) {
	tests := []struct {
		name          string
		fromVersion   string
		toVersion     string
		setupVersions bool
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "valid compatibility check",
			fromVersion:   "1.0.0",
			toVersion:     "1.1.0",
			setupVersions: true,
			wantErr:       false,
		},
		{
			name:        "missing from version",
			fromVersion: "",
			toVersion:   "1.1.0",
			wantErr:     true,
			errMsg:      "both from_version and to_version are required",
		},
		{
			name:        "missing to version",
			fromVersion: "1.0.0",
			toVersion:   "",
			wantErr:     true,
			errMsg:      "both from_version and to_version are required",
		},
		{
			name:        "version not found",
			fromVersion: "9.9.9",
			toVersion:   "1.1.0",
			wantErr:     true,
			errMsg:      "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAPIVersionService()

			if tt.setupVersions {
				// Create test versions
				v1 := models.APIVersion{
					Version:     "1.0.0",
					Description: "Version 1.0",
					ReleasedAt:  time.Now().Add(-48 * time.Hour),
				}
				v2 := models.APIVersion{
					Version:     "1.1.0",
					Description: "Version 1.1",
					ReleasedAt:  time.Now(),
				}

				_, err := service.CreateAPIVersion(context.Background(), v1)
				assert.NoError(t, err)
				_, err = service.CreateAPIVersion(context.Background(), v2)
				assert.NoError(t, err)
			}

			result, err := service.GetVersionCompatibility(context.Background(), tt.fromVersion, tt.toVersion)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				assert.Equal(t, tt.fromVersion, resultMap["from_version"])
				assert.Equal(t, tt.toVersion, resultMap["to_version"])
				assert.Contains(t, resultMap, "compatibility")
				assert.Contains(t, resultMap, "breaking_changes")
				assert.Contains(t, resultMap, "recommendations")
				assert.Contains(t, resultMap, "analyzed_at")
			}
		})
	}
}

func TestAPIVersionServiceImpl_CreateVersionMigration(t *testing.T) {
	tests := []struct {
		name    string
		req     models.VersionMigration
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid migration creation",
			req: models.VersionMigration{
				FromVersion: "1.0.0",
				ToVersion:   "1.1.0",
				Description: "Migration to 1.1.0",
				Changes:     []models.VersionChange{{Type: "feature", Description: "Added new endpoint"}},
				Status:      "planned",
			},
			wantErr: false,
		},
		{
			name: "missing from version",
			req: models.VersionMigration{
				ToVersion:   "1.1.0",
				Description: "Test migration",
			},
			wantErr: true,
			errMsg:  "both from_version and to_version are required",
		},
		{
			name: "missing to version",
			req: models.VersionMigration{
				FromVersion: "1.0.0",
				Description: "Test migration",
			},
			wantErr: true,
			errMsg:  "both from_version and to_version are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAPIVersionService()

			result, err := service.CreateVersionMigration(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.FromVersion, result.FromVersion)
				assert.Equal(t, tt.req.ToVersion, result.ToVersion)
				assert.Equal(t, tt.req.Description, result.Description)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, "planned", result.Status)
				assert.NotZero(t, result.CreatedAt)
			}
		})
	}
}
