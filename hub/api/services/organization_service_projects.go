// Package services - Organization service project operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"sentinel-hub-api/models"
	"strings"
	"time"
)

// CreateProject creates a new project within an organization
func (s *OrganizationServiceImpl) CreateProject(ctx context.Context, orgID string, req models.CreateProjectRequest) (*models.Project, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	// Verify organization exists
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil || org == nil {
		return nil, fmt.Errorf("organization not found")
	}

	// Check for duplicate project names within organization
	projects, err := s.projectRepo.FindByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing projects: %w", err)
	}

	for _, project := range projects {
		if strings.EqualFold(project.Name, req.Name) {
			return nil, fmt.Errorf("project with name '%s' already exists in this organization", req.Name)
		}
	}

	// Generate API key and hash it for secure storage
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Generate hash and prefix for secure storage
	hash, prefix := s.hashAPIKey(apiKey)

	project := &models.Project{
		ID:           generateProjectID(),
		OrgID:        orgID,
		Name:         req.Name,
		APIKey:       "", // Don't store plaintext - only return it once
		APIKeyHash:   hash,
		APIKeyPrefix: prefix,
		CreatedAt:    time.Now(),
	}

	if err := s.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	// Set APIKey in returned object only (for user to save)
	// This is the only time the plaintext key is available
	project.APIKey = apiKey

	return project, nil
}

// GetProject retrieves a project by ID
func (s *OrganizationServiceImpl) GetProject(ctx context.Context, id string) (*models.Project, error) {
	if id == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}

	return project, nil
}

// ListProjects retrieves projects for an organization
func (s *OrganizationServiceImpl) ListProjects(ctx context.Context, orgID string) ([]models.Project, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	projects, err := s.projectRepo.FindByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}

// UpdateProject updates a project with validation
func (s *OrganizationServiceImpl) UpdateProject(ctx context.Context, id string, req models.UpdateProjectRequest) (*models.Project, error) {
	if id == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}

	// Check for duplicate names within organization (excluding current project)
	projects, err := s.projectRepo.FindByOrganizationID(ctx, project.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing projects: %w", err)
	}

	for _, existingProject := range projects {
		if existingProject.ID != id && strings.EqualFold(existingProject.Name, req.Name) {
			return nil, fmt.Errorf("project with name '%s' already exists in this organization", req.Name)
		}
	}

	project.Name = req.Name
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// DeleteProject deletes a project with validation
func (s *OrganizationServiceImpl) DeleteProject(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("project ID is required")
	}

	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("project not found")
	}

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// Helper function
func generateProjectID() string {
	return time.Now().Format("20060102150405") + "_proj"
}
