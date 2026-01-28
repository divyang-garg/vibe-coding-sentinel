// Phase 14E: Task Storage Layer
// Database operations for task management

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// GetTask retrieves a task by ID (with caching)
func GetTask(ctx context.Context, taskID string) (*Task, error) {
	// Check cache first
	if cachedTask, found := GetCachedTask(taskID); found {
		return cachedTask, nil
	}

	query := `
		SELECT id, project_id, source, title, description, file_path, line_number,
		       status, priority, assigned_to, estimated_effort, actual_effort, tags,
		       verification_confidence, version, created_at, updated_at, completed_at, verified_at, archived_at
		FROM tasks
		WHERE id = $1
	`

	row := queryRowWithTimeout(ctx, query, taskID)

	var task Task
	var lineNumber sql.NullInt64
	var assignedTo, description, filePath sql.NullString
	var completedAt, verifiedAt, archivedAt sql.NullTime
	var estimatedEffort, actualEffort sql.NullInt64
	var tags []string

	err := row.Scan(
		&task.ID, &task.ProjectID, &task.Source, &task.Title, &description,
		&filePath, &lineNumber, &task.Status, &task.Priority, &assignedTo,
		&estimatedEffort, &actualEffort, pq.Array(&tags), &task.VerificationConfidence,
		&task.Version, &task.CreatedAt, &task.UpdatedAt, &completedAt, &verifiedAt, &archivedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if description.Valid {
		task.Description = description.String
	}
	if filePath.Valid {
		task.FilePath = filePath.String
	}
	if lineNumber.Valid {
		ln := int(lineNumber.Int64)
		task.LineNumber = &ln
	}
	if assignedTo.Valid {
		task.AssignedTo = &assignedTo.String
	}
	if estimatedEffort.Valid {
		ef := int(estimatedEffort.Int64)
		task.EstimatedEffort = &ef
	}
	if actualEffort.Valid {
		af := int(actualEffort.Int64)
		task.ActualEffort = &af
	}
	task.Tags = tags
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if verifiedAt.Valid {
		task.VerifiedAt = &verifiedAt.Time
	}
	if archivedAt.Valid {
		task.ArchivedAt = &archivedAt.Time
	}

	// Cache the task
	SetCachedTask(taskID, &task)

	return &task, nil
}

// CreateTask creates a new task
func CreateTask(ctx context.Context, projectID string, req CreateTaskRequest) (*Task, error) {
	taskID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO tasks (
			id, project_id, source, title, description, file_path, line_number,
			status, priority, assigned_to, tags, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 1)
		RETURNING id, created_at, updated_at
	`

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	row := queryRowWithTimeout(ctx, query,
		taskID, projectID, req.Source, req.Title, req.Description,
		req.FilePath, req.LineNumber, "pending", priority, req.AssignedTo,
		pq.Array(req.Tags), now, now,
	)

	var createdAt, updatedAt time.Time
	if err := row.Scan(&taskID, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return GetTask(ctx, taskID)
}

// UpdateTask updates an existing task with optimistic locking
func UpdateTask(ctx context.Context, taskID string, req UpdateTaskRequest) (*Task, error) {
	// Get current task to check version
	currentTask, err := GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Check version for optimistic locking
	if currentTask.Version != req.Version {
		return nil, fmt.Errorf("task was modified by another operation (version mismatch: expected %d, got %d)", currentTask.Version, req.Version)
	}

	// Build update query dynamically
	updates := []string{"updated_at = $1", "version = version + 1"}
	args := []interface{}{time.Now()}
	argIndex := 2

	if req.Title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		if *req.Status == "completed" {
			updates = append(updates, fmt.Sprintf("completed_at = $%d", argIndex+1))
			now := time.Now()
			args = append(args, now)
			argIndex++
		}
		argIndex++
	}
	if req.Priority != nil {
		updates = append(updates, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *req.Priority)
		argIndex++
	}
	if req.AssignedTo != nil {
		updates = append(updates, fmt.Sprintf("assigned_to = $%d", argIndex))
		args = append(args, *req.AssignedTo)
		argIndex++
	}
	if req.Tags != nil {
		updates = append(updates, fmt.Sprintf("tags = $%d", argIndex))
		args = append(args, pq.Array(req.Tags))
		argIndex++
	}

	if len(updates) == 2 { // Only updated_at and version
		return currentTask, nil // No changes
	}

	args = append(args, taskID)
	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d",
		strings.Join(updates, ", "), argIndex)

	_, err = execWithTimeout(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Invalidate cache after successful update
	InvalidateTaskCache(taskID)

	return GetTask(ctx, taskID)
}

// ListTasks lists tasks with filters and pagination
func ListTasks(ctx context.Context, projectID string, req ListTasksRequest) (*ListTasksResponse, error) {
	whereClauses := []string{"project_id = $1"}
	args := []interface{}{projectID}
	argIndex := 2

	if req.StatusFilter != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, req.StatusFilter)
		argIndex++
	}

	// Exclude archived tasks by default (unless explicitly included)
	includeArchived := false
	if req.IncludeArchived != nil && *req.IncludeArchived {
		includeArchived = true
	}
	if !includeArchived {
		whereClauses = append(whereClauses, "archived_at IS NULL")
	}
	if req.PriorityFilter != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, req.PriorityFilter)
		argIndex++
	}
	if req.SourceFilter != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("source = $%d", argIndex))
		args = append(args, req.SourceFilter)
		argIndex++
	}
	if req.AssignedTo != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("assigned_to = $%d", argIndex))
		args = append(args, *req.AssignedTo)
		argIndex++
	}
	if len(req.Tags) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("tags && $%d", argIndex))
		args = append(args, pq.Array(req.Tags))
		argIndex++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", whereClause)
	var total int
	row := queryRowWithTimeout(ctx, countQuery, args...)
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Get tasks with pagination
	limit := req.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, project_id, source, title, description, file_path, line_number,
		       status, priority, assigned_to, estimated_effort, actual_effort, tags,
		       verification_confidence, version, created_at, updated_at, completed_at, verified_at, archived_at
		FROM tasks
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := queryWithTimeout(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var lineNumber sql.NullInt64
		var assignedTo, description, filePath sql.NullString
		var completedAt, verifiedAt, archivedAt sql.NullTime
		var estimatedEffort, actualEffort sql.NullInt64
		var tags []string

		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.Source, &task.Title, &description,
			&filePath, &lineNumber, &task.Status, &task.Priority, &assignedTo,
			&estimatedEffort, &actualEffort, pq.Array(&tags), &task.VerificationConfidence,
			&task.Version, &task.CreatedAt, &task.UpdatedAt, &completedAt, &verifiedAt, &archivedAt,
		)
		if err != nil {
			continue
		}

		if description.Valid {
			task.Description = description.String
		}
		if filePath.Valid {
			task.FilePath = filePath.String
		}
		if lineNumber.Valid {
			ln := int(lineNumber.Int64)
			task.LineNumber = &ln
		}
		if assignedTo.Valid {
			task.AssignedTo = &assignedTo.String
		}
		if estimatedEffort.Valid {
			ef := int(estimatedEffort.Int64)
			task.EstimatedEffort = &ef
		}
		if actualEffort.Valid {
			af := int(actualEffort.Int64)
			task.ActualEffort = &af
		}
		task.Tags = tags
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if verifiedAt.Valid {
			task.VerifiedAt = &verifiedAt.Time
		}
		if archivedAt.Valid {
			task.ArchivedAt = &archivedAt.Time
		}

		tasks = append(tasks, task)
	}

	hasNext := offset+limit < total
	hasPrevious := offset > 0

	return &ListTasksResponse{
		Tasks:       tasks,
		Total:       total,
		Limit:       limit,
		Offset:      offset,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}

// DeleteTask soft deletes a task (marks as deleted)
func DeleteTask(ctx context.Context, taskID string) error {
	// Invalidate cache
	InvalidateTaskCache(taskID)

	// For now, we'll do a hard delete. Can be changed to soft delete later
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := execWithTimeout(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}
