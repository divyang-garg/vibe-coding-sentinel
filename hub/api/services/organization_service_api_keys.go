// Package services - Organization service API key management
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"sentinel-hub-api/models"
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

	// Generate hash and prefix for secure storage
	hash, prefix := s.hashAPIKey(apiKey)

	// Store hash instead of plaintext (defense-in-depth security)
	project.APIKeyHash = hash
	project.APIKeyPrefix = prefix
	// Clear any old plaintext API key
	project.APIKey = ""

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return "", fmt.Errorf("failed to update project API key: %w", err)
	}

	// Return plaintext key ONLY once (user must save it)
	return apiKey, nil
}

// ValidateAPIKey validates an API key by comparing its hash and returns the associated project
func (s *OrganizationServiceImpl) ValidateAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Generate hash from provided key
	hash, prefix := s.hashAPIKey(apiKey)

	// Find project by hash (secure lookup)
	project, err := s.projectRepo.FindByAPIKeyHash(ctx, hash)
	if err != nil {
		// Fallback: try old plaintext lookup for migration period
		if project == nil {
			oldProject, oldErr := s.projectRepo.FindByAPIKey(ctx, apiKey)
			if oldErr == nil && oldProject != nil {
				// Migrate old key to hash format
				project = oldProject
				project.APIKeyHash = hash
				project.APIKeyPrefix = prefix
				project.APIKey = "" // Clear plaintext
				s.projectRepo.Update(ctx, project)
			}
		}
		if project == nil {
			return nil, fmt.Errorf("failed to validate API key: %w", err)
		}
	}

	if project == nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Additional verification: check prefix matches (fast check before hash comparison)
	if project.APIKeyPrefix != "" && project.APIKeyPrefix != prefix {
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

	// Clear the API key (both hash and plaintext for security)
	project.APIKey = ""
	project.APIKeyHash = ""
	project.APIKeyPrefix = ""
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	return nil
}

// generateAPIKey generates a cryptographically secure random API key
// Uses crypto/rand for secure random number generation
func (s *OrganizationServiceImpl) generateAPIKey() (string, error) {
	// Generate 32 bytes of cryptographically secure random data
	// This provides 256 bits of entropy
	const keyLength = 32

	key := make([]byte, keyLength)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate secure random key: %w", err)
	}

	// Base64 URL encoding produces a URL-safe string with 43-44 characters
	// Using URL encoding to avoid special characters that might cause issues
	apiKey := base64.URLEncoding.EncodeToString(key)

	// Remove padding if present (optional, but cleaner)
	apiKey = strings.TrimRight(apiKey, "=")

	return apiKey, nil
}

// hashAPIKey generates SHA-256 hash and prefix for an API key
// Returns the hash (hex-encoded) and prefix (first 8 characters)
func (s *OrganizationServiceImpl) hashAPIKey(apiKey string) (hash, prefix string) {
	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	hash = hex.EncodeToString(hasher.Sum(nil))

	// Store first 8 characters for identification (without compromising security)
	if len(apiKey) >= 8 {
		prefix = apiKey[:8]
	} else {
		prefix = apiKey
	}
	return hash, prefix
}
