// Package handlers - Document processing HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// DocumentHandler handles document-related HTTP requests
type DocumentHandler struct {
	BaseHandler
	DocumentService services.DocumentService
}

// NewDocumentHandler creates a new document handler with dependencies
func NewDocumentHandler(documentService services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		DocumentService: documentService,
	}
}

// UploadDocument handles POST /api/v1/documents/upload
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	var req models.DocumentUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	// In a real implementation, this would handle file upload
	// For now, return a placeholder response
	response := &models.DocumentUploadResponse{
		Document: &models.Document{
			ID:       "doc-123",
			Name:     req.Name,
			Status:   models.DocumentStatusUploaded,
			Progress: 0,
		},
		Success: true,
		Message: "Document uploaded successfully",
	}

	h.WriteJSONResponse(w, http.StatusCreated, response)
}

// GetDocument handles GET /api/v1/documents/{id}
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Document ID is required",
		}, http.StatusBadRequest)
		return
	}

	// Placeholder implementation
	doc := &models.Document{
		ID:     id,
		Name:   "sample.pdf",
		Status: models.DocumentStatusCompleted,
	}

	h.WriteJSONResponse(w, http.StatusOK, doc)
}

// ListDocuments handles GET /api/v1/documents
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "Project ID is required",
		}, http.StatusBadRequest)
		return
	}

	// Placeholder implementation
	documents := []models.Document{
		{
			ID:        "doc-1",
			ProjectID: projectID,
			Name:      "document1.pdf",
			Status:    models.DocumentStatusCompleted,
		},
		{
			ID:        "doc-2",
			ProjectID: projectID,
			Name:      "document2.pdf",
			Status:    models.DocumentStatusProcessing,
		},
	}

	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"documents": documents,
		"total":     len(documents),
	})
}

// GetDocumentStatus handles GET /api/v1/documents/{id}/status
func (h *DocumentHandler) GetDocumentStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Document ID is required",
		}, http.StatusBadRequest)
		return
	}

	// Placeholder implementation
	status := &models.DocumentProcessingResult{
		DocumentID: id,
		Status:     "completed",
		Success:    true,
	}

	h.WriteJSONResponse(w, http.StatusOK, status)
}
