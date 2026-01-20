// Package services - Document service core operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"os"
	"sentinel-hub-api/models"
	"time"
)

// DocumentService defines the interface for document-related business operations
type DocumentService interface {
	// Document lifecycle operations
	UploadDocument(ctx context.Context, req models.DocumentUploadRequest, filePath string, mimeType string) (*models.DocumentUploadResponse, error)
	GetDocument(ctx context.Context, id string) (*models.Document, error)
	ListDocuments(ctx context.Context, projectID string) ([]models.Document, error)
	DeleteDocument(ctx context.Context, id string) error

	// Document processing operations
	ProcessDocument(ctx context.Context, docID string) (*models.DocumentProcessingResult, error)
	GetProcessingStatus(ctx context.Context, docID string) (*models.Document, error)
	RetryProcessing(ctx context.Context, docID string) error

	// Knowledge extraction operations
	ExtractKnowledge(ctx context.Context, docID string) ([]models.KnowledgeItem, error)
	GetKnowledgeItems(ctx context.Context, docID string) ([]models.KnowledgeItem, error)
	SearchKnowledge(ctx context.Context, projectID string, query string) ([]models.KnowledgeItem, error)

	// Document analysis operations
	AnalyzeDocument(ctx context.Context, docID string) (*models.DocumentProcessingResult, error)
	ValidateDocument(ctx context.Context, docID string) error
}

// DocumentServiceImpl implements DocumentService
type DocumentServiceImpl struct {
	docRepo            DocumentRepository
	knowledgeExtractor KnowledgeExtractor
	documentValidator  DocumentValidator
	searchEngine       SearchEngine
	logger             Logger
}

// Logger defines the interface for structured logging
type Logger interface {
	Warn(ctx context.Context, msg string, fields map[string]interface{})
	Error(ctx context.Context, msg string, err error, fields ...map[string]interface{})
	Info(ctx context.Context, msg string, fields ...map[string]interface{})
	Debug(ctx context.Context, msg string, fields ...map[string]interface{})
}

// NewDocumentService creates a new document service instance
func NewDocumentService(docRepo DocumentRepository, knowledgeExtractor KnowledgeExtractor, validator DocumentValidator, searchEngine SearchEngine, logger Logger) DocumentService {
	return &DocumentServiceImpl{
		docRepo:            docRepo,
		knowledgeExtractor: knowledgeExtractor,
		documentValidator:  validator,
		searchEngine:       searchEngine,
		logger:             logger,
	}
}

// DocumentRepository defines the interface for document data access
type DocumentRepository interface {
	Save(ctx context.Context, doc *models.Document) error
	FindByID(ctx context.Context, id string) (*models.Document, error)
	FindByProjectID(ctx context.Context, projectID string) ([]models.Document, error)
	Update(ctx context.Context, doc *models.Document) error
	Delete(ctx context.Context, id string) error

	// Processing status operations
	UpdateStatus(ctx context.Context, id string, status string, progress int, errorMsg string) error
	UpdateProcessedAt(ctx context.Context, id string, processedAt *time.Time) error
}

// KnowledgeExtractor defines the interface for knowledge extraction operations
type KnowledgeExtractor interface {
	ExtractFromText(ctx context.Context, text string, docID string) ([]models.KnowledgeItem, error)
	ExtractFromFile(ctx context.Context, filePath string, mimeType string, docID string) ([]models.KnowledgeItem, error)
	ClassifyKnowledgeItem(ctx context.Context, item *models.KnowledgeItem) error
	ValidateKnowledgeItem(ctx context.Context, item *models.KnowledgeItem) error
}

// DocumentValidator defines the interface for document validation operations
type DocumentValidator interface {
	ValidateFile(ctx context.Context, filePath string, mimeType string) error
	ValidateContent(ctx context.Context, content []byte, mimeType string) error
	CheckSecurity(ctx context.Context, filePath string) error
	ValidateSize(ctx context.Context, size int64) error
}

// SearchEngine defines the interface for document search operations
type SearchEngine interface {
	IndexDocument(ctx context.Context, docID string, content string, knowledgeItems []models.KnowledgeItem) error
	SearchDocuments(ctx context.Context, projectID string, query string) ([]models.KnowledgeItem, error)
	UpdateIndex(ctx context.Context, docID string, content string, knowledgeItems []models.KnowledgeItem) error
	DeleteFromIndex(ctx context.Context, docID string) error
}

// UploadDocument handles document upload with validation and initial processing
func (s *DocumentServiceImpl) UploadDocument(ctx context.Context, req models.DocumentUploadRequest, filePath string, mimeType string) (*models.DocumentUploadResponse, error) {
	// Validate request
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.OriginalName == "" {
		return nil, fmt.Errorf("document name is required")
	}

	// Validate file
	if err := s.documentValidator.ValidateFile(ctx, filePath, mimeType); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create document record
	doc := &models.Document{
		ID:           generateDocumentID(),
		ProjectID:    req.ProjectID,
		Name:         req.OriginalName,
		OriginalName: req.OriginalName,
		FilePath:     filePath,
		MimeType:     mimeType,
		Size:         fileInfo.Size(),
		Status:       models.DocumentStatusQueued,
		Progress:     0,
		CreatedAt:    time.Now(),
	}

	// Save document record
	if err := s.docRepo.Save(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	return &models.DocumentUploadResponse{
		Document: doc,
		Success:  true,
		Message:  "Document uploaded successfully, processing will start shortly",
	}, nil
}

// GetDocument retrieves a document by ID
func (s *DocumentServiceImpl) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	if id == "" {
		return nil, fmt.Errorf("document ID is required")
	}

	doc, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	return doc, nil
}

// ListDocuments retrieves documents for a project
func (s *DocumentServiceImpl) ListDocuments(ctx context.Context, projectID string) ([]models.Document, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	docs, err := s.docRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	return docs, nil
}

// DeleteDocument deletes a document and its associated data
func (s *DocumentServiceImpl) DeleteDocument(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("document ID is required")
	}

	// Get document to check if it exists and get file path
	doc, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find document: %w", err)
	}
	if doc == nil {
		return fmt.Errorf("document not found")
	}

	// Remove from search index
	if err := s.searchEngine.DeleteFromIndex(ctx, id); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to remove document from search index: %v\n", err)
	}

	// Delete the physical file if it exists
	if doc.FilePath != "" {
		if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to delete file %s: %v\n", doc.FilePath, err)
		}
	}

	// Delete document record
	if err := s.docRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// generateDocumentID generates a unique document ID
func generateDocumentID() string {
	return time.Now().Format("20060102150405") + "_doc"
}
