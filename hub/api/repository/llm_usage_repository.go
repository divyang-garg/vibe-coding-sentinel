// Package repository provides LLM usage tracking repository implementation
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/models"
)

// LLMUsageRepository defines the interface for LLM usage data operations
type LLMUsageRepository interface {
	SaveUsage(ctx context.Context, usage *models.LLMUsage) error
	GetUsageByProject(ctx context.Context, projectID string, limit, offset int) ([]*models.LLMUsage, int, error)
	GetUsageByValidationID(ctx context.Context, validationID string) ([]*models.LLMUsage, error)
}

// LLMUsageRepositoryImpl implements LLMUsageRepository
type LLMUsageRepositoryImpl struct {
	db Database
}

// NewLLMUsageRepository creates a new LLM usage repository instance
func NewLLMUsageRepository(db Database) LLMUsageRepository {
	return &LLMUsageRepositoryImpl{db: db}
}

// SaveUsage saves LLM usage tracking data to the database
func (r *LLMUsageRepositoryImpl) SaveUsage(ctx context.Context, usage *models.LLMUsage) error {
	if usage == nil {
		return fmt.Errorf("usage cannot be nil")
	}

	query := `
		INSERT INTO llm_usage (id, project_id, validation_id, provider, model, tokens_used, estimated_cost, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			tokens_used = EXCLUDED.tokens_used,
			estimated_cost = EXCLUDED.estimated_cost
	`

	var createdAt time.Time
	if usage.CreatedAt != "" {
		var err error
		createdAt, err = time.Parse(time.RFC3339, usage.CreatedAt)
		if err != nil {
			createdAt = time.Now()
		}
	} else {
		createdAt = time.Now()
	}

	_, err := r.db.Exec(ctx, query,
		usage.ID,
		usage.ProjectID,
		usage.ValidationID,
		usage.Provider,
		usage.Model,
		usage.TokensUsed,
		usage.EstimatedCost,
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save LLM usage: %w", err)
	}

	return nil
}

// GetUsageByProject retrieves LLM usage records for a project with pagination
func (r *LLMUsageRepositoryImpl) GetUsageByProject(ctx context.Context, projectID string, limit, offset int) ([]*models.LLMUsage, int, error) {
	if projectID == "" {
		return nil, 0, fmt.Errorf("project ID is required")
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM llm_usage WHERE project_id = $1`
	var totalCount int
	row := r.db.QueryRow(ctx, countQuery, projectID)
	if err := row.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count LLM usage: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, project_id, validation_id, provider, model, tokens_used, estimated_cost, created_at
		FROM llm_usage
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query LLM usage: %w", err)
	}
	defer rows.Close()

	var usages []*models.LLMUsage
	for rows.Next() {
		usage := &models.LLMUsage{}
		var createdAt time.Time
		var validationID *string

		err := rows.Scan(
			&usage.ID,
			&usage.ProjectID,
			&validationID,
			&usage.Provider,
			&usage.Model,
			&usage.TokensUsed,
			&usage.EstimatedCost,
			&createdAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan LLM usage: %w", err)
		}

		if validationID != nil {
			usage.ValidationID = *validationID
		}
		usage.CreatedAt = createdAt.Format(time.RFC3339)

		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating LLM usage: %w", err)
	}

	return usages, totalCount, nil
}

// GetUsageByValidationID retrieves LLM usage records for a specific validation
func (r *LLMUsageRepositoryImpl) GetUsageByValidationID(ctx context.Context, validationID string) ([]*models.LLMUsage, error) {
	if validationID == "" {
		return nil, fmt.Errorf("validation ID is required")
	}

	query := `
		SELECT id, project_id, validation_id, provider, model, tokens_used, estimated_cost, created_at
		FROM llm_usage
		WHERE validation_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, validationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query LLM usage: %w", err)
	}
	defer rows.Close()

	var usages []*models.LLMUsage
	for rows.Next() {
		usage := &models.LLMUsage{}
		var createdAt time.Time
		var valID *string

		err := rows.Scan(
			&usage.ID,
			&usage.ProjectID,
			&valID,
			&usage.Provider,
			&usage.Model,
			&usage.TokensUsed,
			&usage.EstimatedCost,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan LLM usage: %w", err)
		}

		if valID != nil {
			usage.ValidationID = *valID
		}
		usage.CreatedAt = createdAt.Format(time.RFC3339)

		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating LLM usage: %w", err)
	}

	return usages, nil
}
