// Package repository contains workflow repository implementations.
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sentinel-hub-api/models"
	"time"
)

// WorkflowRepositoryImpl implements workflow data access
type WorkflowRepositoryImpl struct {
	db Database
}

// NewWorkflowRepository creates a new workflow repository instance
func NewWorkflowRepository(db Database) *WorkflowRepositoryImpl {
	return &WorkflowRepositoryImpl{db: db}
}

// Save saves a workflow definition to the database
func (r *WorkflowRepositoryImpl) Save(ctx context.Context, workflow *models.WorkflowDefinition) error {
	stepsJSON, err := json.Marshal(workflow.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	var inputSchemaJSON, outputSchemaJSON sql.NullString
	if workflow.InputSchema != nil {
		inputBytes, err := json.Marshal(workflow.InputSchema)
		if err == nil {
			inputSchemaJSON = sql.NullString{String: string(inputBytes), Valid: true}
		}
	}
	if workflow.OutputSchema != nil {
		outputBytes, err := json.Marshal(workflow.OutputSchema)
		if err == nil {
			outputSchemaJSON = sql.NullString{String: string(outputBytes), Valid: true}
		}
	}

	query := `
		INSERT INTO workflows (id, name, description, version, steps, input_schema, output_schema, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			version = EXCLUDED.version,
			steps = EXCLUDED.steps,
			input_schema = EXCLUDED.input_schema,
			output_schema = EXCLUDED.output_schema,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	if workflow.CreatedAt.IsZero() {
		workflow.CreatedAt = now
	}
	workflow.UpdatedAt = now

	_, err = r.db.Exec(ctx, query,
		workflow.ID, workflow.Name, workflow.Description, workflow.Version,
		string(stepsJSON), inputSchemaJSON, outputSchemaJSON,
		workflow.CreatedAt, workflow.UpdatedAt,
	)

	return err
}

// FindByID retrieves a workflow by ID
func (r *WorkflowRepositoryImpl) FindByID(ctx context.Context, id string) (*models.WorkflowDefinition, error) {
	query := `
		SELECT id, name, description, version, steps, input_schema, output_schema, created_at, updated_at
		FROM workflows WHERE id = $1
	`

	var workflow models.WorkflowDefinition
	var stepsJSON string
	var inputSchemaJSON, outputSchemaJSON sql.NullString

	rowErr := r.db.QueryRow(ctx, query, id).Scan(
		&workflow.ID, &workflow.Name, &workflow.Description, &workflow.Version,
		&stepsJSON, &inputSchemaJSON, &outputSchemaJSON,
		&workflow.CreatedAt, &workflow.UpdatedAt,
	)

	if rowErr != nil {
		return nil, rowErr
	}

	// Unmarshal steps
	if err := json.Unmarshal([]byte(stepsJSON), &workflow.Steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
	}

	// Unmarshal schemas if present
	if inputSchemaJSON.Valid {
		if err := json.Unmarshal([]byte(inputSchemaJSON.String), &workflow.InputSchema); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input_schema: %w", err)
		}
	}
	if outputSchemaJSON.Valid {
		if err := json.Unmarshal([]byte(outputSchemaJSON.String), &workflow.OutputSchema); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output_schema: %w", err)
		}
	}

	return &workflow, nil
}

// List retrieves workflows with pagination
func (r *WorkflowRepositoryImpl) List(ctx context.Context, limit, offset int) ([]models.WorkflowDefinition, int, error) {
	// Count total
	countQuery := "SELECT COUNT(*) FROM workflows"
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, name, description, version, steps, input_schema, output_schema, created_at, updated_at
		FROM workflows
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var workflows []models.WorkflowDefinition
	for rows.Next() {
		var workflow models.WorkflowDefinition
		var stepsJSON string
		var inputSchemaJSON, outputSchemaJSON sql.NullString

		err := rows.Scan(
			&workflow.ID, &workflow.Name, &workflow.Description, &workflow.Version,
			&stepsJSON, &inputSchemaJSON, &outputSchemaJSON,
			&workflow.CreatedAt, &workflow.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Unmarshal steps
		if err := json.Unmarshal([]byte(stepsJSON), &workflow.Steps); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal steps: %w", err)
		}

		// Unmarshal schemas if present
		if inputSchemaJSON.Valid {
			if err := json.Unmarshal([]byte(inputSchemaJSON.String), &workflow.InputSchema); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal input_schema: %w", err)
			}
		}
		if outputSchemaJSON.Valid {
			if err := json.Unmarshal([]byte(outputSchemaJSON.String), &workflow.OutputSchema); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal output_schema: %w", err)
			}
		}

		workflows = append(workflows, workflow)
	}

	return workflows, total, nil
}

// SaveExecution saves a workflow execution to the database
func (r *WorkflowRepositoryImpl) SaveExecution(ctx context.Context, execution *models.WorkflowExecution) error {
	stepResultsJSON, err := json.Marshal(execution.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal step_results: %w", err)
	}

	var inputJSON, outputJSON, errorJSON sql.NullString
	if execution.Input != nil {
		inputBytes, err := json.Marshal(execution.Input)
		if err == nil {
			inputJSON = sql.NullString{String: string(inputBytes), Valid: true}
		}
	}
	if execution.Output != nil {
		outputBytes, err := json.Marshal(execution.Output)
		if err == nil {
			outputJSON = sql.NullString{String: string(outputBytes), Valid: true}
		}
	}
	if execution.Error != nil {
		errorBytes, err := json.Marshal(execution.Error)
		if err == nil {
			errorJSON = sql.NullString{String: string(errorBytes), Valid: true}
		}
	}

	query := `
		INSERT INTO workflow_executions (id, workflow_id, status, input, output, progress, started_at, completed_at, error, step_results)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7, $8, $9::jsonb, $10::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			input = EXCLUDED.input,
			output = EXCLUDED.output,
			progress = EXCLUDED.progress,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			error = EXCLUDED.error,
			step_results = EXCLUDED.step_results
	`

	_, err = r.db.Exec(ctx, query,
		execution.ID, execution.WorkflowID, string(execution.Status),
		inputJSON, outputJSON, execution.Progress,
		execution.StartedAt, execution.CompletedAt, errorJSON,
		string(stepResultsJSON),
	)

	return err
}

// FindExecutionByID retrieves a workflow execution by ID
func (r *WorkflowRepositoryImpl) FindExecutionByID(ctx context.Context, id string) (*models.WorkflowExecution, error) {
	query := `
		SELECT id, workflow_id, status, input, output, progress, started_at, completed_at, error, step_results
		FROM workflow_executions WHERE id = $1
	`

	var execution models.WorkflowExecution
	var statusStr string
	var inputJSON, outputJSON sql.NullString
	var stepResultsJSON string
	var errorStr sql.NullString

	rowErr := r.db.QueryRow(ctx, query, id).Scan(
		&execution.ID, &execution.WorkflowID, &statusStr,
		&inputJSON, &outputJSON, &execution.Progress,
		&execution.StartedAt, &execution.CompletedAt, &errorStr,
		&stepResultsJSON,
	)

	if rowErr != nil {
		return nil, rowErr
	}

	execution.Status = models.WorkflowStatus(statusStr)

	// Handle error field
	if errorStr.Valid && errorStr.String != "" {
		var workflowError models.WorkflowError
		if err := json.Unmarshal([]byte(errorStr.String), &workflowError); err == nil {
			execution.Error = &workflowError
		}
	}

	// Unmarshal JSON fields
	if inputJSON.Valid {
		if err := json.Unmarshal([]byte(inputJSON.String), &execution.Input); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input: %w", err)
		}
	}
	if outputJSON.Valid {
		if err := json.Unmarshal([]byte(outputJSON.String), &execution.Output); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output: %w", err)
		}
	}
	if err := json.Unmarshal([]byte(stepResultsJSON), &execution.Steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal step_results: %w", err)
	}

	return &execution, nil
}
