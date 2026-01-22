// Package services - Organization service core operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"sentinel-hub-api/models"
	"strings"
	"time"
)

// OrganizationService defines the interface for organization-related business operations
type OrganizationService interface {
	// Organization operations
	CreateOrganization(ctx context.Context, req models.CreateOrganizationRequest) (*models.Organization, error)
	GetOrganization(ctx context.Context, id string) (*models.Organization, error)
	ListOrganizations(ctx context.Context) ([]models.Organization, error)
	UpdateOrganization(ctx context.Context, id string, req models.UpdateOrganizationRequest) (*models.Organization, error)
	DeleteOrganization(ctx context.Context, id string) error

	// Project operations within organizations
	CreateProject(ctx context.Context, orgID string, req models.CreateProjectRequest) (*models.Project, error)
	GetProject(ctx context.Context, id string) (*models.Project, error)
	ListProjects(ctx context.Context, orgID string) ([]models.Project, error)
	UpdateProject(ctx context.Context, id string, req models.UpdateProjectRequest) (*models.Project, error)
	DeleteProject(ctx context.Context, id string) error

	// API key management
	GenerateAPIKey(ctx context.Context, projectID string) (string, error)
	ValidateAPIKey(ctx context.Context, apiKey string) (*models.Project, error)
	RevokeAPIKey(ctx context.Context, projectID string) error
}

// OrganizationServiceImpl implements OrganizationService
type OrganizationServiceImpl struct {
	orgRepo     OrganizationRepository
	projectRepo ProjectRepository
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(orgRepo OrganizationRepository, projectRepo ProjectRepository) OrganizationService {
	return &OrganizationServiceImpl{
		orgRepo:     orgRepo,
		projectRepo: projectRepo,
	}
}

// OrganizationRepository defines the interface for organization data access
type OrganizationRepository interface {
	Save(ctx context.Context, org *models.Organization) error
	FindByID(ctx context.Context, id string) (*models.Organization, error)
	FindAll(ctx context.Context) ([]models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id string) error
}

// ProjectRepository defines the interface for project data access
type ProjectRepository interface {
	Save(ctx context.Context, project *models.Project) error
	FindByID(ctx context.Context, id string) (*models.Project, error)
	FindByOrganizationID(ctx context.Context, orgID string) ([]models.Project, error)
	FindByAPIKey(ctx context.Context, apiKey string) (*models.Project, error)         // Legacy: for migration support
	FindByAPIKeyHash(ctx context.Context, apiKeyHash string) (*models.Project, error) // Secure: hash-based lookup
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id string) error
}

// CreateOrganization creates a new organization with validation
func (s *OrganizationServiceImpl) CreateOrganization(ctx context.Context, req models.CreateOrganizationRequest) (*models.Organization, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	if len(req.Name) < 2 || len(req.Name) > 100 {
		return nil, fmt.Errorf("organization name must be between 2 and 100 characters")
	}

	// Check if organization with same name already exists
	orgs, err := s.orgRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing organizations: %w", err)
	}

	for _, org := range orgs {
		if strings.EqualFold(org.Name, req.Name) {
			return nil, fmt.Errorf("organization with name '%s' already exists", req.Name)
		}
	}

	org := &models.Organization{
		ID:          generateOrganizationID(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.orgRepo.Save(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to save organization: %w", err)
	}

	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationServiceImpl) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	if id == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	org, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

// ListOrganizations retrieves all organizations
func (s *OrganizationServiceImpl) ListOrganizations(ctx context.Context) ([]models.Organization, error) {
	orgs, err := s.orgRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return orgs, nil
}

// UpdateOrganization updates an organization with validation
func (s *OrganizationServiceImpl) UpdateOrganization(ctx context.Context, id string, req models.UpdateOrganizationRequest) (*models.Organization, error) {
	if id == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	org, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, fmt.Errorf("organization not found")
	}

	// Validate and apply updates
	if req.Name != "" {
		if len(req.Name) < 2 || len(req.Name) > 100 {
			return nil, fmt.Errorf("organization name must be between 2 and 100 characters")
		}

		// Check for name conflicts
		orgs, err := s.orgRepo.FindAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing organizations: %w", err)
		}

		for _, existingOrg := range orgs {
			if existingOrg.ID != id && strings.EqualFold(existingOrg.Name, req.Name) {
				return nil, fmt.Errorf("organization with name '%s' already exists", req.Name)
			}
		}

		org.Name = req.Name
	}

	if req.Description != "" {
		org.Description = req.Description
	}

	org.UpdatedAt = time.Now()

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return org, nil
}

// DeleteOrganization deletes an organization and all its projects
func (s *OrganizationServiceImpl) DeleteOrganization(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("organization ID is required")
	}

	// Check if organization exists
	org, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return fmt.Errorf("organization not found")
	}

	// Delete all projects in the organization first
	projects, err := s.projectRepo.FindByOrganizationID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get organization projects: %w", err)
	}

	for _, project := range projects {
		if err := s.projectRepo.Delete(ctx, project.ID); err != nil {
			return fmt.Errorf("failed to delete project %s: %w", project.ID, err)
		}
	}

	// Delete the organization
	if err := s.orgRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

// Helper functions
func generateOrganizationID() string {
	return time.Now().Format("20060102150405") + "_org"
}
