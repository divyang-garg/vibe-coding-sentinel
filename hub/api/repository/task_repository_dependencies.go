// Package repository - Task dependency data access operations
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"context"
	"sentinel-hub-api/models"
)

// SaveDependency saves a task dependency to the database
func (r *TaskRepositoryImpl) SaveDependency(ctx context.Context, dep *models.TaskDependency) error {
	query := `
		INSERT INTO task_dependencies (id, task_id, depends_on_task_id, dependency_type, confidence, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query, dep.ID, dep.TaskID, dep.DependsOnTaskID, dep.DependencyType, dep.Confidence, dep.CreatedAt)
	return err
}

// FindDependencies retrieves dependencies for a task
func (r *TaskRepositoryImpl) FindDependencies(ctx context.Context, taskID string) ([]models.TaskDependency, error) {
	query := "SELECT id, task_id, depends_on_task_id, dependency_type, confidence, created_at FROM task_dependencies WHERE task_id = $1"
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []models.TaskDependency
	for rows.Next() {
		var dep models.TaskDependency
		err := rows.Scan(&dep.ID, &dep.TaskID, &dep.DependsOnTaskID, &dep.DependencyType, &dep.Confidence, &dep.CreatedAt)
		if err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}

	return deps, nil
}

// DeleteDependency removes a task dependency
func (r *TaskRepositoryImpl) DeleteDependency(ctx context.Context, id string) error {
	query := "DELETE FROM task_dependencies WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// FindDependents finds tasks that depend on the given task (where taskID is the depends_on_task_id)
func (r *TaskRepositoryImpl) FindDependents(ctx context.Context, taskID string) ([]models.TaskDependency, error) {
	query := `
		SELECT id, task_id, depends_on_task_id, dependency_type, confidence, created_at
		FROM task_dependencies
		WHERE depends_on_task_id = $1`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []models.TaskDependency
	for rows.Next() {
		var dep models.TaskDependency
		err := rows.Scan(&dep.ID, &dep.TaskID, &dep.DependsOnTaskID, &dep.DependencyType, &dep.Confidence, &dep.CreatedAt)
		if err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return deps, nil
}
