// Package services - Document service processing operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"os"
	"sentinel-hub-api/models"
	"time"
)

// ProcessDocument processes a document and extracts knowledge
func (s *DocumentServiceImpl) ProcessDocument(ctx context.Context, docID string) (*models.DocumentProcessingResult, error) {
	if docID == "" {
		return nil, fmt.Errorf("document ID is required")
	}

	// Get document
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	// Check if already processed
	if doc.Status == models.DocumentStatusCompleted {
		return nil, fmt.Errorf("document already processed")
	}

	// Update status to processing
	if err := s.docRepo.UpdateStatus(ctx, docID, string(models.DocumentStatusProcessing), 10, ""); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// Extract text content based on MIME type
	var extractedText string
	switch doc.MimeType {
	case "text/plain":
		content, readErr := os.ReadFile(doc.FilePath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read text file: %w", readErr)
		}
		extractedText = string(content)
	case "text/markdown":
		content, readErr := os.ReadFile(doc.FilePath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read markdown file: %w", readErr)
		}
		extractedText = string(content)
	case "application/pdf":
		var extractErr error
		extractedText, extractErr = extractPDFText(doc.FilePath)
		if extractErr != nil {
			return nil, fmt.Errorf("failed to extract PDF text: %w", extractErr)
		}
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/msword":
		var extractErr error
		extractedText, extractErr = extractDOCXText(doc.FilePath)
		if extractErr != nil {
			return nil, fmt.Errorf("failed to extract DOCX text: %w", extractErr)
		}
	case "image/png", "image/jpeg", "image/jpg", "image/gif":
		var extractErr error
		extractedText, extractErr = extractImageText(doc.FilePath)
		if extractErr != nil {
			return nil, fmt.Errorf("failed to extract image text: %w", extractErr)
		}
	default:
		extractedText = fmt.Sprintf("Document content extraction not implemented for %s", doc.MimeType)
	}

	// Update document with extracted text
	doc.ExtractedText = extractedText
	if err := s.docRepo.Save(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to save extracted text: %w", err)
	}

	// Extract knowledge items
	knowledgeItems, err := s.knowledgeExtractor.ExtractFromText(ctx, extractedText, docID)
	if err != nil {
		// Update status with error
		s.docRepo.UpdateStatus(ctx, docID, "failed", 0, err.Error())
		return nil, fmt.Errorf("failed to extract knowledge: %w", err)
	}

	// Classify and validate knowledge items
	for i := range knowledgeItems {
		if err := s.knowledgeExtractor.ClassifyKnowledgeItem(ctx, &knowledgeItems[i]); err != nil {
			continue // Skip on classification error
		}
		if err := s.knowledgeExtractor.ValidateKnowledgeItem(ctx, &knowledgeItems[i]); err != nil {
			continue // Skip invalid items
		}
	}

	// Index document for search
	if err := s.searchEngine.IndexDocument(ctx, docID, extractedText, knowledgeItems); err != nil {
		// Log but don't fail - indexing is not critical for processing
		// TODO: Use structured logging when available
		fmt.Printf("Warning: failed to index document %s: %v\n", docID, err)
	}

	// Mark as completed
	now := time.Now()
	if err := s.docRepo.UpdateStatus(ctx, docID, string(models.DocumentStatusCompleted), 100, ""); err != nil {
		return nil, fmt.Errorf("failed to mark as completed: %w", err)
	}
	if err := s.docRepo.UpdateProcessedAt(ctx, docID, &now); err != nil {
		return nil, fmt.Errorf("failed to update processed timestamp: %w", err)
	}

	result := &models.DocumentProcessingResult{
		DocumentID:     docID,
		Status:         "completed",
		KnowledgeItems: knowledgeItems,
		ProcessedAt:    now,
		Success:        true,
	}

	return result, nil
}

// GetProcessingStatus retrieves current processing status
func (s *DocumentServiceImpl) GetProcessingStatus(ctx context.Context, docID string) (*models.Document, error) {
	return s.GetDocument(ctx, docID)
}

// RetryProcessing retries failed document processing
func (s *DocumentServiceImpl) RetryProcessing(ctx context.Context, docID string) error {
	if docID == "" {
		return fmt.Errorf("document ID is required")
	}

	// Get document
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to find document: %w", err)
	}
	if doc == nil {
		return fmt.Errorf("document not found")
	}

	// Check if it's in a failed state
	if doc.Status != models.DocumentStatusFailed {
		return fmt.Errorf("document is not in failed state")
	}

	// Reset to queued status
	if err := s.docRepo.UpdateStatus(ctx, docID, string(models.DocumentStatusQueued), 0, ""); err != nil {
		return fmt.Errorf("failed to reset processing status: %w", err)
	}

	return nil
}

// AnalyzeDocument provides document analysis results
func (s *DocumentServiceImpl) AnalyzeDocument(ctx context.Context, docID string) (*models.DocumentProcessingResult, error) {
	if docID == "" {
		return nil, fmt.Errorf("document ID is required")
	}

	// Get document
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	// Return analysis result based on current status
	result := &models.DocumentProcessingResult{
		DocumentID: docID,
		Status:     string(doc.Status),
		Success:    doc.Status == models.DocumentStatusCompleted,
	}
	if doc.ProcessedAt != nil {
		result.ProcessedAt = *doc.ProcessedAt
	}

	if doc.Status == models.DocumentStatusCompleted {
		// Get knowledge items for completed documents
		knowledgeItems, err := s.GetKnowledgeItems(ctx, docID)
		if err == nil {
			result.KnowledgeItems = knowledgeItems
		}
	}

	return result, nil
}

// ValidateDocument validates a document's integrity
func (s *DocumentServiceImpl) ValidateDocument(ctx context.Context, docID string) error {
	if docID == "" {
		return fmt.Errorf("document ID is required")
	}

	// Get document
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to find document: %w", err)
	}
	if doc == nil {
		return fmt.Errorf("document not found")
	}

	// Check if file exists
	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("document file not found on disk")
	}

	// Validate content
	if err := s.documentValidator.ValidateFile(ctx, doc.FilePath, doc.MimeType); err != nil {
		return fmt.Errorf("document validation failed: %w", err)
	}

	return nil
}
