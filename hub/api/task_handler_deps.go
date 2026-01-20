// Phase 14E: Task Management API Handlers - Dependency Operations
// HTTP handlers for task dependency endpoints
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sentinel-hub-api/services"

	"github.com/go-chi/chi/v5"
)

// getTaskDependenciesHandler handles GET /api/v1/tasks/{id}/dependencies
func getTaskDependenciesHandler(w http.ResponseWriter, r *http.Request) {
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

	response, err := services.GetTaskDependencies(ctx, taskID)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to get dependencies")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "get_dependencies",
			Message:       "Failed to get dependencies",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// detectDependenciesHandler handles POST /api/v1/tasks/{id}/detect-dependencies
func detectDependenciesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Detect dependencies
	dependencies, err := services.DetectDependencies(ctx, taskID, codebasePath)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to detect dependencies")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "detect_dependencies",
			Message:       "Failed to detect dependencies",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":      taskID,
		"dependencies": dependencies,
		"count":        len(dependencies),
	})
}
