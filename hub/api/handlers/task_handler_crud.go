// Package task_handler_crud - Basic CRUD handlers for task management
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// createTaskHandler handles POST /api/v1/tasks
func createTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	// Validate and sanitize input
	if req.Title == "" {
		WriteErrorResponse(w, &ValidationError{
			Field:   "title",
			Message: "Title is required",
			Code:    "required",
		}, http.StatusBadRequest)
		return
	}

	// Sanitize title and description (prevent XSS, SQL injection)
	req.Title = sanitizeString(req.Title, GetConfig().Limits.MaxTaskTitleLength)
	if len(req.Title) > GetConfig().Limits.MaxTaskTitleLength {
		WriteErrorResponse(w, &ValidationError{
			Field:   "title",
			Message: fmt.Sprintf("Title too long (max %d characters)", GetConfig().Limits.MaxTaskTitleLength),
			Code:    "too_long",
		}, http.StatusBadRequest)
		return
	}

	if req.Description != "" {
		req.Description = sanitizeString(req.Description, GetConfig().Limits.MaxTaskDescriptionLength)
		if len(req.Description) > GetConfig().Limits.MaxTaskDescriptionLength {
			WriteErrorResponse(w, &ValidationError{
				Field:   "description",
				Message: fmt.Sprintf("Description too long (max %d characters)", GetConfig().Limits.MaxTaskDescriptionLength),
				Code:    "too_long",
			}, http.StatusBadRequest)
			return
		}
	}

	// Validate source
	validSources := map[string]bool{
		"cursor": true, "manual": true, "change_request": true, "comprehensive_analysis": true,
	}
	if req.Source == "" {
		req.Source = "manual"
	} else if !validSources[req.Source] {
		WriteErrorResponse(w, &ValidationError{
			Field:   "source",
			Message: "Invalid source",
			Code:    "invalid_enum",
		}, http.StatusBadRequest)
		return
	}

	// Validate priority
	if req.Priority != "" {
		validPriorities := map[string]bool{
			"low": true, "medium": true, "high": true, "critical": true,
		}
		if !validPriorities[req.Priority] {
			WriteErrorResponse(w, &ValidationError{
				Field:   "priority",
				Message: "Invalid priority",
				Code:    "invalid_enum",
			}, http.StatusBadRequest)
			return
		}
	}

	// Validate file path (prevent path traversal)
	if req.FilePath != "" {
		req.FilePath = sanitizePath(req.FilePath)
		if !isValidPath(req.FilePath) {
			WriteErrorResponse(w, &ValidationError{
				Field:   "file_path",
				Message: "Invalid file path",
				Code:    "invalid_path",
			}, http.StatusBadRequest)
			return
		}
	}

	task, err := CreateTask(ctx, project.ID, req)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to create task")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "create_task",
			Message:       "Failed to create task",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// getTaskHandler handles GET /api/v1/tasks/{id}
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	// Verify project access
	project, err := getProjectFromContext(ctx)
	if err != nil || task.ProjectID != project.ID {
		WriteErrorResponse(w, &ValidationError{
			Field:   "authorization",
			Message: "Unauthorized",
			Code:    "unauthorized",
		}, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// updateTaskHandler handles PUT /api/v1/tasks/{id}
