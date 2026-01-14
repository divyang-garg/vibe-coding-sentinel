// Package repository contains data access implementations for the Sentinel Hub API.
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"context"
	"sentinel-hub-api/models"
	"time"
)

// TaskRepositoryImpl implements TaskRepository
type TaskRepositoryImpl struct {
	db Database
}

// NewTaskRepository creates a new task repository instance
func NewTaskRepository(db Database) *TaskRepositoryImpl {
	return &TaskRepositoryImpl{db: db}
}

// Database defines the interface for database operations
type Database interface {
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	BeginTx(ctx context.Context) (Transaction, error)
}

// Row represents a single database row
type Row interface {
	Scan(dest ...interface{}) error
}

// Rows represents multiple database rows
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

// Result represents the result of a database operation
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Transaction represents a database transaction
type Transaction interface {
	Database
	Commit() error
	Rollback() error
}

// Save saves a task to the database
func (r *TaskRepositoryImpl) Save(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (id, project_id, source, title, description, file_path, line_number,
		                  status, priority, assigned_to, estimated_effort, verification_confidence,
		                  created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			source = EXCLUDED.source,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			file_path = EXCLUDED.file_path,
			line_number = EXCLUDED.line_number,
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			assigned_to = EXCLUDED.assigned_to,
			estimated_effort = EXCLUDED.estimated_effort,
			verification_confidence = EXCLUDED.verification_confidence,
			updated_at = EXCLUDED.updated_at,
			version = EXCLUDED.version
		WHERE tasks.version = EXCLUDED.version - 1`

	_, err := r.db.Exec(ctx, query,
		task.ID, task.ProjectID, task.Source, task.Title, task.Description,
		task.FilePath, task.LineNumber, string(task.Status), string(task.Priority), task.AssignedTo,
		task.EstimatedEffort, task.VerificationConfidence, task.CreatedAt, task.UpdatedAt, task.Version)

	return err
}

// FindByID retrieves a task by ID
func (r *TaskRepositoryImpl) FindByID(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, project_id, source, title, description, file_path, line_number,
		       status, priority, assigned_to, estimated_effort, actual_effort,
		       verification_confidence, created_at, updated_at, completed_at, verified_at, archived_at, version
		FROM tasks WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)

	var task models.Task
	var completedAt, verifiedAt, archivedAt *time.Time

	err := row.Scan(
		&task.ID, &task.ProjectID, &task.Source, &task.Title, &task.Description,
		&task.FilePath, &task.LineNumber, &task.Status, &task.Priority, &task.AssignedTo,
		&task.EstimatedEffort, &task.ActualEffort, &task.VerificationConfidence,
		&task.CreatedAt, &task.UpdatedAt, &completedAt, &verifiedAt, &archivedAt, &task.Version,
	)

	if err != nil {
		return nil, err
	}

	task.CompletedAt = completedAt
	task.VerifiedAt = verifiedAt
	task.ArchivedAt = archivedAt

	return &task, nil
}

// FindByProjectID retrieves tasks for a project with filtering and pagination
func (r *TaskRepositoryImpl) FindByProjectID(ctx context.Context, projectID string, filters models.ListTasksRequest) ([]models.Task, int, error) {
	whereClause := "WHERE project_id = $1"
	args := []interface{}{projectID}
	argCount := 1

	if filters.Status != "" {
		argCount++
		whereClause += " AND status = $" + string(rune('0'+argCount))
		args = append(args, filters.Status)
	}

	if filters.Priority != "" {
		argCount++
		whereClause += " AND priority = $" + string(rune('0'+argCount))
		args = append(args, filters.Priority)
	}

	if filters.AssignedTo != "" {
		argCount++
		whereClause += " AND assigned_to = $" + string(rune('0'+argCount))
		args = append(args, filters.AssignedTo)
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM tasks " + whereClause
	row := r.db.QueryRow(ctx, countQuery, args...)
	var total int
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, project_id, source, title, description, file_path, line_number,
		       status, priority, assigned_to, estimated_effort, actual_effort,
		       verification_confidence, created_at, updated_at, completed_at, verified_at, archived_at, version
		FROM tasks ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + string(rune('0'+argCount+1)) + ` OFFSET $` + string(rune('0'+argCount+2))

	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var completedAt, verifiedAt, archivedAt *time.Time

		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.Source, &task.Title, &task.Description,
			&task.FilePath, &task.LineNumber, &task.Status, &task.Priority, &task.AssignedTo,
			&task.EstimatedEffort, &task.ActualEffort, &task.VerificationConfidence,
			&task.CreatedAt, &task.UpdatedAt, &completedAt, &verifiedAt, &archivedAt, &task.Version,
		)
		if err != nil {
			return nil, 0, err
		}

		task.CompletedAt = completedAt
		task.VerifiedAt = verifiedAt
		task.ArchivedAt = archivedAt

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// Update updates a task in the database
func (r *TaskRepositoryImpl) Update(ctx context.Context, task *models.Task) error {
	query := `
		UPDATE tasks SET
			source = $2, title = $3, description = $4, file_path = $5, line_number = $6,
			status = $7, priority = $8, assigned_to = $9, estimated_effort = $10,
			actual_effort = $11, verification_confidence = $12, updated_at = $13,
			completed_at = $14, verified_at = $15, archived_at = $16, version = $17
		WHERE id = $1 AND version = $18`

	_, err := r.db.Exec(ctx, query,
		task.ID, task.Source, task.Title, task.Description, task.FilePath, task.LineNumber,
		string(task.Status), string(task.Priority), task.AssignedTo, task.EstimatedEffort,
		task.ActualEffort, task.VerificationConfidence, task.UpdatedAt, task.CompletedAt,
		task.VerifiedAt, task.ArchivedAt, task.Version, task.Version-1)

	return err
}

// Delete removes a task from the database
func (r *TaskRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM tasks WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}
