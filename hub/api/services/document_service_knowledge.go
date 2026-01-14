// Package services - Document service knowledge operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"sentinel-hub-api/models"
)

// ExtractKnowledge extracts knowledge items from a processed document
func (s *DocumentServiceImpl) ExtractKnowledge(ctx context.Context, docID string) ([]models.KnowledgeItem, error) {
	if docID == "" {
		return nil, fmt.Errorf("document ID is required")
	}

	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	if doc.Status != models.DocumentStatusCompleted {
		return nil, fmt.Errorf("document not yet processed")
	}

	// Extract knowledge from text
	knowledgeItems, err := s.knowledgeExtractor.ExtractFromText(ctx, doc.ExtractedText, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract knowledge: %w", err)
	}

	return knowledgeItems, nil
}

// GetKnowledgeItems retrieves knowledge items for a document
func (s *DocumentServiceImpl) GetKnowledgeItems(ctx context.Context, docID string) ([]models.KnowledgeItem, error) {
	// For now, extract fresh - in production this would be cached
	return s.ExtractKnowledge(ctx, docID)
}

// SearchKnowledge performs knowledge search across project documents
func (s *DocumentServiceImpl) SearchKnowledge(ctx context.Context, projectID string, query string) ([]models.KnowledgeItem, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	results, err := s.searchEngine.SearchDocuments(ctx, projectID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	return results, nil
}
