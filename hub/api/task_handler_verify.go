// Phase 14E: Task Management API Handlers - Verification Operations
// HTTP handlers for task verification endpoints
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// verifyTaskHandler handles POST /api/v1/tasks/{id}/verify
func verifyTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	// Validate UUID format
	if err := ValidateUUID(taskID); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "id",
			Message: "Invalid task ID format",
			Code:    "invalid_uuid",
		}, http.StatusBadRequest)
		return
	}

	// Verify task exists and project access
	task, err := GetTask(ctx, taskID)
	if err != nil {
		if err.Error() == "task not found: "+taskID {
			WriteErrorResponse(w, &NotFoundError{
				Resource: "task",
				ID:       taskID,
				Message:  "Task not found",
			}, http.StatusNotFound)
			return
		}
		LogErrorWithContext(ctx, err, "Failed to get task")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "get_task",
			Message:       "Failed to get task",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	project, err := getProjectFromContext(ctx)
	if err != nil || task.ProjectID != project.ID {
		WriteErrorResponse(w, &ValidationError{
			Field:   "authorization",
			Message: "Unauthorized",
			Code:    "unauthorized",
		}, http.StatusUnauthorized)
		return
	}

	var req VerifyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to force=false if body is empty
		req.Force = false
	}

	// Get codebase path from request or use default
	codebasePath := r.URL.Query().Get("codebasePath")
	if codebasePath == "" {
		codebasePath = "." // Default to current directory
	}

	// Sanitize path to prevent directory traversal attacks
	codebasePath = sanitizePath(codebasePath)

	// Validate that the path is safe to use
	if !isValidPath(codebasePath) {
		LogErrorWithContext(ctx, fmt.Errorf("invalid codebase path: %s", codebasePath), "Invalid codebase path provided")
		WriteErrorResponse(w, &ValidationError{
			Field:   "codebasePath",
			Message: "Invalid or unsafe codebase path",
			Code:    "invalid_path",
		}, http.StatusBadRequest)
		return
	}

	result, err := VerifyTask(ctx, taskID, codebasePath, req.Force)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to verify task")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "verify_task",
			Message:       "Failed to verify task",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// scanTasksHandler handles POST /api/v1/tasks/scan
func scanTasksHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	project, err := getProjectFromContext(ctx)
	if err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "authorization",
			Message: "Unauthorized",
			Code:    "unauthorized",
		}, http.StatusUnauthorized)
		return
	}

	var req struct {
		CodebasePath string `json:"codebasePath"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.CodebasePath == "" {
		WriteErrorResponse(w, &ValidationError{
			Field:   "codebasePath",
			Message: "codebasePath is required",
			Code:    "required",
		}, http.StatusBadRequest)
		return
	}

	// Detect tasks
	detector := NewTaskDetector(project.ID, req.CodebasePath)
	tasks, err := detector.DetectTasks(ctx)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to detect tasks")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "detect_tasks",
			Message:       "Failed to detect tasks",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// Deduplicate
	uniqueTasks, err := DeduplicateTasks(ctx, project.ID, tasks)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to deduplicate tasks")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "deduplicate_tasks",
			Message:       "Failed to deduplicate tasks",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// Store tasks
	taskIDs, err := StoreDetectedTasks(ctx, project.ID, uniqueTasks)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to store tasks")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "store_tasks",
			Message:       "Failed to store tasks",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"detected": len(tasks),
		"unique":   len(uniqueTasks),
		"stored":   len(taskIDs),
		"task_ids": taskIDs,
	})
}

// verifyAllTasksHandler handles POST /api/v1/tasks/verify-all
func verifyAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	project, err := getProjectFromContext(ctx)
	if err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "authorization",
			Message: "Unauthorized",
			Code:    "unauthorized",
		}, http.StatusUnauthorized)
		return
	}

	var req struct {
		StatusFilter   string `json:"status,omitempty"`
		PriorityFilter string `json:"priority,omitempty"`
		Force          bool   `json:"force,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use defaults if body is empty
		req.Force = false
	}

	// Get codebase path
	codebasePath := r.URL.Query().Get("codebasePath")
	if codebasePath == "" {
		codebasePath = "."
	}

	// Sanitize path to prevent directory traversal attacks
	codebasePath = sanitizePath(codebasePath)

	// Validate that the path is safe to use
	if !isValidPath(codebasePath) {
		LogErrorWithContext(ctx, fmt.Errorf("invalid codebase path: %s", codebasePath), "Invalid codebase path provided")
		WriteErrorResponse(w, &ValidationError{
			Field:   "codebasePath",
			Message: "Invalid or unsafe codebase path",
			Code:    "invalid_path",
		}, http.StatusBadRequest)
		return
	}

	// Build list request
	listReq := ListTasksRequest{
		StatusFilter:   req.StatusFilter,
		PriorityFilter: req.PriorityFilter,
		Limit:          GetConfig().Limits.MaxTaskListLimit,
		Offset:         0,
	}

	// Get tasks to verify
	response, err := ListTasks(ctx, project.ID, listReq)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to list tasks")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "list_tasks",
			Message:       "Failed to list tasks",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// Verify tasks (sequentially for now, can be parallelized later)
	verifiedCount := 0
	failedCount := 0
	skippedCount := 0

	for _, task := range response.Tasks {
		if task.Status == "completed" {
			skippedCount++
			continue
		}

		_, err := VerifyTask(ctx, task.ID, codebasePath, req.Force)
		if err != nil {
			LogErrorWithContext(ctx, err, fmt.Sprintf("Failed to verify task %s", task.ID))
			failedCount++
		} else {
			verifiedCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":    len(response.Tasks),
		"verified": verifiedCount,
		"failed":   failedCount,
		"skipped":  skippedCount,
	})
}
