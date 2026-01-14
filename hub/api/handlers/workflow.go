// Package handlers - Workflow orchestration HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// WorkflowHandler handles workflow-related HTTP requests
type WorkflowHandler struct {
	BaseHandler
	WorkflowService services.WorkflowService
}

// NewWorkflowHandler creates a new workflow handler with dependencies
func NewWorkflowHandler(workflowService services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		WorkflowService: workflowService,
	}
}

// CreateWorkflow handles POST /api/v1/workflows
func (h *WorkflowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var req models.WorkflowDefinition
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	workflow, err := h.WorkflowService.CreateWorkflow(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusCreated, workflow)
}

// GetWorkflow handles GET /api/v1/workflows/{id}
func (h *WorkflowHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Workflow ID is required",
		}, http.StatusBadRequest)
		return
	}

	workflow, err := h.WorkflowService.GetWorkflow(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if workflow == nil {
		h.WriteErrorResponse(w, &models.NotFoundError{
			Resource: "workflow",
			Message:  "Workflow not found",
		}, http.StatusNotFound)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, workflow)
}

// ListWorkflows handles GET /api/v1/workflows
func (h *WorkflowHandler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit := 50 // default
	offset := 0 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	workflows, total, err := h.WorkflowService.ListWorkflows(r.Context(), limit, offset)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"workflows": workflows,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// ExecuteWorkflow handles POST /api/v1/workflows/{id}/execute
func (h *WorkflowHandler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Workflow ID is required",
		}, http.StatusBadRequest)
		return
	}

	execution, err := h.WorkflowService.ExecuteWorkflow(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusAccepted, execution)
}

// GetWorkflowExecution handles GET /api/v1/workflows/executions/{id}
func (h *WorkflowHandler) GetWorkflowExecution(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Execution ID is required",
		}, http.StatusBadRequest)
		return
	}

	execution, err := h.WorkflowService.GetWorkflowExecution(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if execution == nil {
		h.WriteErrorResponse(w, &models.NotFoundError{
			Resource: "workflow_execution",
			Message:  "Workflow execution not found",
		}, http.StatusNotFound)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, execution)
}
