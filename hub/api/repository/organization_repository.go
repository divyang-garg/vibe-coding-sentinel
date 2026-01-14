// Package repository contains organization and project repository implementations.
package repository

import (
	"context"
	"sentinel-hub-api/models"
)

// OrganizationRepositoryImpl implements organization data access
type OrganizationRepositoryImpl struct {
	db Database
}

// NewOrganizationRepository creates a new organization repository instance
func NewOrganizationRepository(db Database) *OrganizationRepositoryImpl {
	return &OrganizationRepositoryImpl{db: db}
}

// Save saves an organization to the database
func (r *OrganizationRepositoryImpl) Save(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (id, name, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name
		WHERE organizations.id = EXCLUDED.id`

	_, err := r.db.Exec(ctx, query, org.ID, org.Name, org.CreatedAt)
	return err
}

// FindByID retrieves an organization by ID
func (r *OrganizationRepositoryImpl) FindByID(ctx context.Context, id string) (*models.Organization, error) {
	query := "SELECT id, name, created_at FROM organizations WHERE id = $1"

	var org models.Organization
	err := r.db.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &org, nil
}

// FindAll retrieves all organizations
func (r *OrganizationRepositoryImpl) FindAll(ctx context.Context) ([]models.Organization, error) {
	query := "SELECT id, name, created_at FROM organizations ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var org models.Organization
		err := rows.Scan(&org.ID, &org.Name, &org.CreatedAt)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// Update updates an organization
func (r *OrganizationRepositoryImpl) Update(ctx context.Context, org *models.Organization) error {
	return r.Save(ctx, org)
}

// Delete deletes an organization
func (r *OrganizationRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM organizations WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ProjectRepositoryImpl implements project data access
type ProjectRepositoryImpl struct {
	db Database
}

// NewProjectRepository creates a new project repository instance
func NewProjectRepository(db Database) *ProjectRepositoryImpl {
	return &ProjectRepositoryImpl{db: db}
}

// Save saves a project to the database
func (r *ProjectRepositoryImpl) Save(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO projects (id, org_id, name, api_key, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			api_key = EXCLUDED.api_key
		WHERE projects.id = EXCLUDED.id`

	_, err := r.db.Exec(ctx, query, project.ID, project.OrgID, project.Name, project.APIKey, project.CreatedAt)
	return err
}

// FindByID retrieves a project by ID
func (r *ProjectRepositoryImpl) FindByID(ctx context.Context, id string) (*models.Project, error) {
	query := "SELECT id, org_id, name, api_key, created_at FROM projects WHERE id = $1"

	var project models.Project
	var apiKey *string

	err := r.db.QueryRow(ctx, query, id).Scan(&project.ID, &project.OrgID, &project.Name, &apiKey, &project.CreatedAt)
	if err != nil {
		return nil, err
	}

	if apiKey != nil {
		project.APIKey = *apiKey
	}

	return &project, nil
}

// FindByOrganizationID retrieves projects by organization ID
func (r *ProjectRepositoryImpl) FindByOrganizationID(ctx context.Context, orgID string) ([]models.Project, error) {
	query := "SELECT id, org_id, name, api_key, created_at FROM projects WHERE org_id = $1 ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		var apiKey *string

		err := rows.Scan(&project.ID, &project.OrgID, &project.Name, &apiKey, &project.CreatedAt)
		if err != nil {
			return nil, err
		}

		if apiKey != nil {
			project.APIKey = *apiKey
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// FindByAPIKey retrieves a project by API key
func (r *ProjectRepositoryImpl) FindByAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	query := "SELECT id, org_id, name, api_key, created_at FROM projects WHERE api_key = $1"

	var project models.Project
	err := r.db.QueryRow(ctx, query, apiKey).Scan(&project.ID, &project.OrgID, &project.Name, &project.APIKey, &project.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// Update updates a project
func (r *ProjectRepositoryImpl) Update(ctx context.Context, project *models.Project) error {
	return r.Save(ctx, project)
}

// Delete deletes a project
func (r *ProjectRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM projects WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}
