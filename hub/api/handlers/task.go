// Package handlers - Task management HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"

	"github.com/go-chi/chi/v5"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	BaseHandler
	TaskService services.TaskService
}

// NewTaskHandler creates a new task handler with dependencies
func NewTaskHandler(taskService services.TaskService) *TaskHandler {
	return &TaskHandler{
		TaskService: taskService,
	}
}

// CreateTask handles POST /api/v1/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	task, err := h.TaskService.CreateTask(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusCreated, task)
}

// GetTask handles GET /api/v1/tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Task ID is required",
		}, http.StatusBadRequest)
		return
	}

	task, err := h.TaskService.GetTaskByID(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if task == nil {
		h.WriteErrorResponse(w, &models.NotFoundError{
			Resource: "task",
			Message:  "Task not found",
		}, http.StatusNotFound)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, task)
}

// ListTasks handles GET /api/v1/tasks
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "Project ID is required",
		}, http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	req := models.ListTasksRequest{
		ProjectID: projectID,
		Limit:     50, // default
		Offset:    0,  // default
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	// Extract and validate status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.TaskStatus(statusStr)
		if !status.IsValid() {
			h.WriteErrorResponse(w, fmt.Errorf("invalid status: %s. Valid values are: pending, in_progress, completed, blocked, cancelled", statusStr), http.StatusBadRequest)
			return
		}
		req.Status = statusStr
	}

	response, err := h.TaskService.ListTasks(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdateTask handles PUT /api/v1/tasks/{id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Task ID is required",
		}, http.StatusBadRequest)
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	task, err := h.TaskService.UpdateTask(r.Context(), id, req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, task)
}

// DeleteTask handles DELETE /api/v1/tasks/{id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Task ID is required",
		}, http.StatusBadRequest)
		return
	}

	err := h.TaskService.DeleteTask(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusNoContent, map[string]string{"status": "deleted"})
}

// VerifyTask handles POST /api/v1/tasks/{id}/verify
func (h *TaskHandler) VerifyTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Task ID is required",
		}, http.StatusBadRequest)
		return
	}

	var req models.VerifyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.TaskService.VerifyTask(r.Context(), id, req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// GetTaskDependencies handles GET /api/v1/tasks/{id}/dependencies
func (h *TaskHandler) GetTaskDependencies(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Task ID is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.TaskService.GetDependencies(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AddTaskDependency handles POST /api/v1/tasks/{id}/dependencies
func (h *TaskHandler) AddTaskDependency(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Task ID is required",
		}, http.StatusBadRequest)
		return
	}

	var req models.AddDependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.TaskService.AddDependency(r.Context(), id, req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusCreated, result)
}
