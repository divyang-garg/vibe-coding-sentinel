// Package utils provides core task integration CRUD functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package utils

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/pkg"
	"sentinel-hub-api/pkg/database"
)

// GetChangeRequestByID retrieves a change request by ID from database
func GetChangeRequestByID(ctx context.Context, id string) (*ChangeRequest, error) {
	if id == "" {
		return nil, fmt.Errorf("change request ID is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	query := `
		SELECT id, project_id, status, implementation_status, type
		FROM change_requests
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, id)

	var cr ChangeRequest
	err := row.Scan(&cr.ID, &cr.ProjectID, &cr.Status, &cr.ImplementationStatus, &cr.Type)
	if err != nil {
		return nil, HandleNotFoundError(err, "change request", id)
	}

	return &cr, nil
}

// GetTask retrieves a task by ID from database
func GetTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	query := `
		SELECT id, status, version
		FROM tasks
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, taskID)

	var task Task
	err := row.Scan(&task.ID, &task.Status, &task.Version)
	if err != nil {
		return nil, HandleNotFoundError(err, "task", taskID)
	}

	return &task, nil
}

// UpdateTask updates a task in the database
func UpdateTask(ctx context.Context, taskID string, req UpdateTaskRequest) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	// Get current task to verify version
	currentTask, err := GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current task: %w", err)
	}

	// Verify version matches (optimistic locking)
	if req.Version > 0 && req.Version != currentTask.Version {
		return nil, fmt.Errorf("task version mismatch: expected %d, got %d", currentTask.Version, req.Version)
	}

	// Build update query dynamically based on provided fields
	updates := []string{"version = version + 1", "updated_at = NOW()"}
	args := []interface{}{}
	argIndex := 1

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

	// Build final query
	setClause := ""
	for i, update := range updates {
		if i > 0 {
			setClause += ", "
		}
		setClause += update
	}

	query := fmt.Sprintf(`
		UPDATE tasks
		SET %s
		WHERE id = $%d`, setClause, argIndex)
	args = append(args, taskID)

	// Add version check if provided
	if req.Version > 0 {
		argIndex++
		query += fmt.Sprintf(" AND version = $%d", argIndex)
		args = append(args, req.Version)
	}

	result, err := database.ExecWithTimeout(ctx, db, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("task not found or version mismatch")
	}

	// Return updated task
	return GetTask(ctx, taskID)
}

// CreateTask creates a new task in the database
func CreateTask(ctx context.Context, projectID string, req CreateTaskRequest) (*Task, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("task title is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	taskID := GenerateEntityID()
	now := time.Now().UTC()

	query := `
		INSERT INTO tasks (id, project_id, source, title, description, status, priority, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, status, version
	`

	status := TaskStatusPending
	if req.Priority == "" {
		req.Priority = TaskPriorityMedium
	}

	row := database.QueryRowWithTimeout(ctx, db, query,
		taskID, projectID, req.Source, req.Title, req.Description,
		status, req.Priority, now, now, 1,
	)

	var task Task
	err := row.Scan(&task.ID, &task.Status, &task.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &task, nil
}

// ListTasks lists tasks for a project from database with pagination and filtering
func ListTasks(ctx context.Context, projectID string, req ListTasksRequest) (*ListTasksResponse, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	// Build WHERE clause
	whereClause := "WHERE project_id = $1"
	args := []interface{}{projectID}
	argCount := 1

	// Apply filters
	status := req.StatusFilter
	if status == "" {
		status = req.Status
	}
	if status != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	priority := req.PriorityFilter
	if priority == "" {
		priority = req.Priority
	}
	if priority != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, priority)
	}

	if !req.IncludeArchived {
		whereClause += " AND archived_at IS NULL"
	}

	// Set default pagination
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Build query with pagination
	argCount++
	query := fmt.Sprintf(`
		SELECT id, status, version
		FROM tasks %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)
	args = append(args, req.Limit, req.Offset)

	rows, err := database.QueryWithTimeout(ctx, db, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Status, &task.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tasks: %w", err)
	}

	return &ListTasksResponse{
		Tasks: tasks,
	}, nil
}

// GetKnowledgeItemByID retrieves a knowledge item by ID from database
func GetKnowledgeItemByID(ctx context.Context, id string) (*KnowledgeItem, error) {
	if id == "" {
		return nil, fmt.Errorf("knowledge item ID is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	query := `
		SELECT id, status
		FROM knowledge_items
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, id)

	var ki KnowledgeItem
	err := row.Scan(&ki.ID, &ki.Status)
	if err != nil {
		return nil, HandleNotFoundError(err, "knowledge item", id)
	}

	return &ki, nil
}

// GetTestRequirementByID retrieves a test requirement by ID from database
func GetTestRequirementByID(ctx context.Context, id string) (*TestRequirement, error) {
	if id == "" {
		return nil, fmt.Errorf("test requirement ID is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	query := `
		SELECT id, rule_title, description
		FROM test_requirements
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, id)

	var tr TestRequirement
	err := row.Scan(&tr.ID, &tr.RuleTitle, &tr.Description)
	if err != nil {
		return nil, HandleNotFoundError(err, "test requirement", id)
	}

	return &tr, nil
}

// GetComprehensiveValidationByID retrieves a comprehensive validation by ID from database
func GetComprehensiveValidationByID(ctx context.Context, id string) (*ComprehensiveValidation, error) {
	if id == "" {
		return nil, fmt.Errorf("comprehensive validation ID is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	query := `
		SELECT validation_id, project_id, feature
		FROM comprehensive_validations
		WHERE validation_id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, id)

	var cv ComprehensiveValidation
	err := row.Scan(&cv.ID, &cv.ProjectID, &cv.Feature)
	if err != nil {
		return nil, HandleNotFoundError(err, "comprehensive validation", id)
	}

	return &cv, nil
}

// LogError logs an error using the proper logging package
func LogError(ctx context.Context, format string, args ...interface{}) {
	pkg.LogError(ctx, format, args...)
}
