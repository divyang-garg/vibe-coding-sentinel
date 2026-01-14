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
