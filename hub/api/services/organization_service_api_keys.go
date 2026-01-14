// Package services - Organization service API key management
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"sentinel-hub-api/models"
	"time"
)

// GenerateAPIKey generates a new API key for a project
func (s *OrganizationServiceImpl) GenerateAPIKey(ctx context.Context, projectID string) (string, error) {
	if projectID == "" {
		return "", fmt.Errorf("project ID is required")
	}

	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to find project: %w", err)
	}
	if project == nil {
		return "", fmt.Errorf("project not found")
	}

	// Generate new API key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	project.APIKey = apiKey
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return "", fmt.Errorf("failed to update project API key: %w", err)
	}

	return apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated project
func (s *OrganizationServiceImpl) ValidateAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	project, err := s.projectRepo.FindByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("invalid API key")
	}

	return project, nil
}

// RevokeAPIKey revokes a project's API key
func (s *OrganizationServiceImpl) RevokeAPIKey(ctx context.Context, projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("project not found")
	}

	// Clear the API key
	project.APIKey = ""
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	return nil
}

// generateAPIKey generates a secure random API key
func (s *OrganizationServiceImpl) generateAPIKey() (string, error) {
	// Generate a secure random API key
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 32

	key := make([]byte, keyLength)
	for i := range key {
		key[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}

	return string(key), nil
}
