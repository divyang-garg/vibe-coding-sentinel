// Package handlers - Test management HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"

	"github.com/go-chi/chi/v5"
)

// TestHandler handles test-related HTTP requests
type TestHandler struct {
	BaseHandler
	TestService services.TestService
}

// NewTestHandler creates a new test handler with dependencies
func NewTestHandler(testService services.TestService) *TestHandler {
	return &TestHandler{
		TestService: testService,
	}
}

// GenerateTestRequirements handles POST /api/v1/test/requirements/generate
func (h *TestHandler) GenerateTestRequirements(w http.ResponseWriter, r *http.Request) {
	var req services.GenerateTestRequirementsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.ProjectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.TestService.GenerateTestRequirements(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeTestCoverage handles POST /api/v1/test/coverage/analyze
func (h *TestHandler) AnalyzeTestCoverage(w http.ResponseWriter, r *http.Request) {
	var req services.AnalyzeCoverageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.ProjectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		}, http.StatusBadRequest)
		return
	}

	if len(req.TestFiles) == 0 {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "test_files",
			Message: "test_files are required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.TestService.AnalyzeTestCoverage(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// GetTestCoverage handles GET /api/v1/test/coverage/{knowledge_item_id}
func (h *TestHandler) GetTestCoverage(w http.ResponseWriter, r *http.Request) {
	knowledgeItemID := chi.URLParam(r, "knowledge_item_id")
	if knowledgeItemID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "knowledge_item_id",
			Message: "knowledge_item_id is required",
		}, http.StatusBadRequest)
		return
	}

	coverage, err := h.TestService.GetTestCoverage(r.Context(), knowledgeItemID)
	if err != nil {
		if err.Error() == "test coverage not found for knowledge item: "+knowledgeItemID {
			h.WriteErrorResponse(w, err, http.StatusNotFound)
			return
		}
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, coverage)
}

// ValidateTests handles POST /api/v1/test/validations/validate
func (h *TestHandler) ValidateTests(w http.ResponseWriter, r *http.Request) {
	var req services.ValidateTestsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.ProjectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		}, http.StatusBadRequest)
		return
	}

	if req.TestCode == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "test_code",
			Message: "test_code is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.TestService.ValidateTests(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// GetValidationResults handles GET /api/v1/test/validations/{test_requirement_id}
func (h *TestHandler) GetValidationResults(w http.ResponseWriter, r *http.Request) {
	testRequirementID := chi.URLParam(r, "test_requirement_id")
	if testRequirementID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "test_requirement_id",
			Message: "test_requirement_id is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.TestService.GetValidationResults(r.Context(), testRequirementID)
	if err != nil {
		if err.Error() == "validation not found for test requirement: "+testRequirementID {
			h.WriteErrorResponse(w, err, http.StatusNotFound)
			return
		}
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// RunTests handles POST /api/v1/test/execution/run
func (h *TestHandler) RunTests(w http.ResponseWriter, r *http.Request) {
	var req services.TestExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.ProjectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		}, http.StatusBadRequest)
		return
	}

	if len(req.TestFiles) == 0 {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "test_files",
			Message: "test_files are required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.TestService.RunTests(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusAccepted, result)
}

// GetTestExecutionStatus handles GET /api/v1/test/execution/{execution_id}
func (h *TestHandler) GetTestExecutionStatus(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "execution_id")
	if executionID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "execution_id",
			Message: "execution_id is required",
		}, http.StatusBadRequest)
		return
	}

	status, err := h.TestService.GetTestExecutionStatus(r.Context(), executionID)
	if err != nil {
		if err.Error() == "test execution not found: "+executionID {
			h.WriteErrorResponse(w, err, http.StatusNotFound)
			return
		}
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, status)
}
