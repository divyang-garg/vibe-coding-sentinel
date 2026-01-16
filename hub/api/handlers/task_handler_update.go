// Package task_handler_update - Task update and management handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	// Validate and sanitize input
	if req.Title != nil {
		*req.Title = sanitizeString(*req.Title, GetConfig().Limits.MaxTaskTitleLength)
		if len(*req.Title) > GetConfig().Limits.MaxTaskTitleLength {
			WriteErrorResponse(w, &ValidationError{
				Field:   "title",
				Message: fmt.Sprintf("Title too long (max %d characters)", GetConfig().Limits.MaxTaskTitleLength),
				Code:    "too_long",
			}, http.StatusBadRequest)
			return
		}
	}

	if req.Description != nil {
		*req.Description = sanitizeString(*req.Description, GetConfig().Limits.MaxTaskDescriptionLength)
		if len(*req.Description) > GetConfig().Limits.MaxTaskDescriptionLength {
			WriteErrorResponse(w, &ValidationError{
				Field:   "description",
				Message: fmt.Sprintf("Description too long (max %d characters)", GetConfig().Limits.MaxTaskDescriptionLength),
				Code:    "too_long",
			}, http.StatusBadRequest)
			return
		}
	}

	if req.Status != nil {
		validStatuses := map[string]bool{
			"pending": true, "in_progress": true, "completed": true, "blocked": true,
		}
		if !validStatuses[*req.Status] {
			WriteErrorResponse(w, &ValidationError{
				Field:   "status",
				Message: "Invalid status",
				Code:    "invalid_enum",
			}, http.StatusBadRequest)
			return
		}
	}

	if req.Priority != nil {
		validPriorities := map[string]bool{
			"low": true, "medium": true, "high": true, "critical": true,
		}
		if !validPriorities[*req.Priority] {
			WriteErrorResponse(w, &ValidationError{
				Field:   "priority",
				Message: "Invalid priority",
				Code:    "invalid_enum",
			}, http.StatusBadRequest)
			return
		}
	}

	updatedTask, err := UpdateTask(ctx, taskID, req)
	if err != nil {
		if err.Error() == "task was modified by another operation" {
			WriteErrorResponse(w, &ValidationError{
				Field:   "version",
				Message: err.Error(),
				Code:    "conflict",
			}, http.StatusConflict)
			return
		}
		LogErrorWithContext(ctx, err, "Failed to update task")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "update_task",
			Message:       "Failed to update task",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

// listTasksHandler handles GET /api/v1/tasks
func listTasksHandler(w http.ResponseWriter, r *http.Request) {
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

	req := ListTasksRequest{
		StatusFilter:   r.URL.Query().Get("status"),
		PriorityFilter: r.URL.Query().Get("priority"),
		SourceFilter:   r.URL.Query().Get("source"),
		Limit:          GetConfig().Limits.DefaultTaskListLimit,
		Offset:         0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}
	if assignedTo := r.URL.Query().Get("assigned_to"); assignedTo != "" {
		req.AssignedTo = &assignedTo
	}
	if tagsStr := r.URL.Query().Get("tags"); tagsStr != "" {
		req.Tags = strings.Split(tagsStr, ",")
	}
	if includeArchivedStr := r.URL.Query().Get("include_archived"); includeArchivedStr == "true" {
		includeArchived := true
		req.IncludeArchived = &includeArchived
	}

	response, err := ListTasks(ctx, project.ID, req)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to list tasks")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "list_tasks",
			Message:       "Failed to list tasks",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// deleteTaskHandler handles DELETE /api/v1/tasks/{id}
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := DeleteTask(ctx, taskID); err != nil {
		LogErrorWithContext(ctx, err, "Failed to delete task")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "delete_task",
			Message:       "Failed to delete task",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// verifyTaskHandler handles POST /api/v1/tasks/{id}/verify
