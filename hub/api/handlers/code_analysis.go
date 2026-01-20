// Package handlers - Code analysis HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// CodeAnalysisHandler handles code analysis-related HTTP requests
type CodeAnalysisHandler struct {
	BaseHandler
	CodeAnalysisService services.CodeAnalysisService
}

// NewCodeAnalysisHandler creates a new code analysis handler with dependencies
func NewCodeAnalysisHandler(codeAnalysisService services.CodeAnalysisService) *CodeAnalysisHandler {
	return &CodeAnalysisHandler{
		CodeAnalysisService: codeAnalysisService,
	}
}

// AnalyzeCode handles POST /api/v1/analyze/code
func (h *CodeAnalysisHandler) AnalyzeCode(w http.ResponseWriter, r *http.Request) {
	var req models.ASTAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "code",
			Message: "Code is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.AnalyzeCode(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// LintCode handles POST /api/v1/lint/code
func (h *CodeAnalysisHandler) LintCode(w http.ResponseWriter, r *http.Request) {
	var req models.CodeLintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "code",
			Message: "Code is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.LintCode(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// RefactorCode handles POST /api/v1/refactor/code
func (h *CodeAnalysisHandler) RefactorCode(w http.ResponseWriter, r *http.Request) {
	var req models.CodeRefactorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "code",
			Message: "Code is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.RefactorCode(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// GenerateDocs handles POST /api/v1/generate/docs
func (h *CodeAnalysisHandler) GenerateDocs(w http.ResponseWriter, r *http.Request) {
	var req models.DocumentationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "code",
			Message: "Code is required",
		}, http.StatusBadRequest)
		return
	}

	docs, err := h.CodeAnalysisService.GenerateDocumentation(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, docs)
}

// ValidateCode handles POST /api/v1/validate/code
func (h *CodeAnalysisHandler) ValidateCode(w http.ResponseWriter, r *http.Request) {
	var req models.CodeValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "code",
			Message: "Code is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.ValidateCode(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeSecurity handles POST /api/v1/analyze/security
func (h *CodeAnalysisHandler) AnalyzeSecurity(w http.ResponseWriter, r *http.Request) {
	var req models.SecurityASTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "code",
			Message: "Code is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.AnalyzeSecurity(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeVibe handles POST /api/v1/analyze/vibe
func (h *CodeAnalysisHandler) AnalyzeVibe(w http.ResponseWriter, r *http.Request) {
	var req models.ASTAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "code",
			Message: "Code is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.AnalyzeVibe(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeComprehensive handles POST /api/v1/analyze/comprehensive
func (h *CodeAnalysisHandler) AnalyzeComprehensive(w http.ResponseWriter, r *http.Request) {
	var req services.ComprehensiveAnalysisRequest
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

	result, err := h.CodeAnalysisService.AnalyzeComprehensive(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeIntent handles POST /api/v1/analyze/intent
func (h *CodeAnalysisHandler) AnalyzeIntent(w http.ResponseWriter, r *http.Request) {
	var req services.IntentAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "prompt",
			Message: "prompt is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.AnalyzeIntent(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeDocSync handles POST /api/v1/analyze/doc-sync
func (h *CodeAnalysisHandler) AnalyzeDocSync(w http.ResponseWriter, r *http.Request) {
	var req services.DocSyncRequest
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

	if req.CodebasePath == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "codebase_path",
			Message: "codebase_path is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.AnalyzeDocSync(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeBusinessRules handles POST /api/v1/analyze/business-rules
func (h *CodeAnalysisHandler) AnalyzeBusinessRules(w http.ResponseWriter, r *http.Request) {
	var req services.BusinessRulesAnalysisRequest
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

	if req.CodebasePath == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "codebase_path",
			Message: "codebase_path is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.CodeAnalysisService.AnalyzeBusinessRules(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}
