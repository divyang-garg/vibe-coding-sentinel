// Package handlers - AST analysis HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// ASTHandler handles AST analysis-related HTTP requests
type ASTHandler struct {
	BaseHandler
	ASTService services.ASTService
}

// NewASTHandler creates a new AST handler with dependencies
func NewASTHandler(astService services.ASTService) *ASTHandler {
	return &ASTHandler{
		ASTService: astService,
	}
}

// AnalyzeAST handles POST /api/v1/ast/analyze
func (h *ASTHandler) AnalyzeAST(w http.ResponseWriter, r *http.Request) {
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

	if req.Language == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "language",
			Message: "Language is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.ASTService.AnalyzeAST(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeMultiFile handles POST /api/v1/ast/analyze/multi
func (h *ASTHandler) AnalyzeMultiFile(w http.ResponseWriter, r *http.Request) {
	var req models.MultiFileASTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "files",
			Message: "At least one file is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.ASTService.AnalyzeMultiFile(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeSecurity handles POST /api/v1/ast/analyze/security
func (h *ASTHandler) AnalyzeSecurity(w http.ResponseWriter, r *http.Request) {
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

	if req.Language == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "language",
			Message: "Language is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.ASTService.AnalyzeSecurity(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// AnalyzeCrossFile handles POST /api/v1/ast/analyze/cross
func (h *ASTHandler) AnalyzeCrossFile(w http.ResponseWriter, r *http.Request) {
	var req models.CrossFileASTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "files",
			Message: "At least one file is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.ASTService.AnalyzeCrossFile(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// GetSupportedAnalyses handles GET /api/v1/ast/supported
func (h *ASTHandler) GetSupportedAnalyses(w http.ResponseWriter, r *http.Request) {
	result, err := h.ASTService.GetSupportedAnalyses(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}
