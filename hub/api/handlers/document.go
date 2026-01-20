// Package handlers - Document processing HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"fmt"
	"net/http"
	"os"

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
	// Parse multipart form (100MB max)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		h.WriteErrorResponse(w, fmt.Errorf("failed to parse multipart form: %w", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.WriteErrorResponse(w, fmt.Errorf("file not found in request: %w", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get project ID from header or context
	projectID := r.Header.Get("X-Project-ID")
	if projectID == "" {
		project, err := h.GetProjectFromContext(r.Context())
		if err != nil {
			h.WriteErrorResponse(w, fmt.Errorf("project ID required: %w", err), http.StatusUnauthorized)
			return
		}
		projectID = project.ID
	}

	// Save file to temp location
	tempPath, err := saveUploadedFile(file, header.Filename)
	if err != nil {
		h.WriteErrorResponse(w, fmt.Errorf("failed to save uploaded file: %w", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		// Clean up temp file if handler returns error
		if tempPath != "" {
			os.Remove(tempPath)
		}
	}()

	// Create upload request
	req := models.DocumentUploadRequest{
		ProjectID:    projectID,
		OriginalName: header.Filename,
		Name:         header.Filename,
	}

	// Detect MIME type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = detectMimeType(header.Filename)
	}

	// Call service
	doc, err := h.DocumentService.UploadDocument(r.Context(), req, tempPath, mimeType)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	// Don't clean up temp file on success - service will handle it
	tempPath = ""

	h.WriteJSONResponse(w, http.StatusCreated, doc)
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

	doc, err := h.DocumentService.GetDocument(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusNotFound)
		return
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

	documents, err := h.DocumentService.ListDocuments(r.Context(), projectID)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
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

	doc, err := h.DocumentService.GetProcessingStatus(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusNotFound)
		return
	}

	status := &models.DocumentProcessingResult{
		DocumentID: doc.ID,
		Status:     string(doc.Status),
		Success:    doc.Status == models.DocumentStatusCompleted,
	}

	h.WriteJSONResponse(w, http.StatusOK, status)
}
