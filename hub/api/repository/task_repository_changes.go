// Package repository - Task change tracking data access operations
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"context"
	"fmt"
	"sentinel-hub-api/models"
)

// SaveChange saves a task change to the database
func (r *TaskRepositoryImpl) SaveChange(ctx context.Context, change *models.TaskChange) error {
	query := `
		INSERT INTO task_changes (id, task_id, change_type, old_values, new_values, changed_by, changed_at, justification)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		change.ID, change.TaskID, change.ChangeType, change.OldValues, change.NewValues,
		change.ChangedBy, change.ChangedAt, change.Justification)
	if err != nil {
		return fmt.Errorf("failed to save change %s for task %s: %w", change.ID, change.TaskID, err)
	}
	return nil
}

// FindChanges retrieves change history for a task
func (r *TaskRepositoryImpl) FindChanges(ctx context.Context, taskID string) ([]models.TaskChange, error) {
	query := "SELECT id, task_id, change_type, old_values, new_values, changed_by, changed_at, justification FROM task_changes WHERE task_id = $1 ORDER BY changed_at DESC"
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query changes for task %s: %w", taskID, err)
	}
	defer rows.Close()

	var changes []models.TaskChange
	for rows.Next() {
		var c models.TaskChange
		err := rows.Scan(&c.ID, &c.TaskID, &c.ChangeType, &c.OldValues, &c.NewValues, &c.ChangedBy, &c.ChangedAt, &c.Justification)
		if err != nil {
			return nil, fmt.Errorf("failed to scan change row: %w", err)
		}
		changes = append(changes, c)
	}

	return changes, nil
}
