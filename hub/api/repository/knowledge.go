// Package repository contains knowledge extraction and document processing implementations.
package repository

import (
	"context"
	"fmt"
	"regexp"
	"sentinel-hub-api/models"
	"strings"
)

// KnowledgeExtractorImpl implements KnowledgeExtractor
type KnowledgeExtractorImpl struct{}

// NewKnowledgeExtractor creates a new knowledge extractor instance
func NewKnowledgeExtractor() *KnowledgeExtractorImpl {
	return &KnowledgeExtractorImpl{}
}

// ExtractFromText extracts knowledge items from plain text
func (k *KnowledgeExtractorImpl) ExtractFromText(ctx context.Context, text string, docID string) ([]models.KnowledgeItem, error) {
	var items []models.KnowledgeItem

	// Extract business rules using regex patterns
	rulePatterns := []struct {
		pattern    *regexp.Regexp
		ruleType   string
		confidence float64
	}{
		{regexp.MustCompile(`(?i)(shall|should|must) ([^\n]+)`), "functional_requirement", 0.8},
		{regexp.MustCompile(`(?i)(performance|latency|throughput)[:\s]+([^\n]+)`), "performance_requirement", 0.7},
		{regexp.MustCompile(`(?i)(security|auth|encrypt)[:\s]+([^\n]+)`), "security_requirement", 0.9},
		{regexp.MustCompile(`(?i)(api|endpoint|interface)[:\s]+([^\n]+)`), "api_definition", 0.6},
		{regexp.MustCompile(`(?i)(table|entity|model)[:\s]+([^\n]+)`), "data_table", 0.7},
	}

	for _, pattern := range rulePatterns {
		matches := pattern.pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				content := strings.TrimSpace(match[len(match)-1])
				if content != "" && len(content) > 10 { // Filter out very short matches
					items = append(items, models.KnowledgeItem{
						ID:         fmt.Sprintf("ki_%s_%d", docID, len(items)+1),
						Type:       pattern.ruleType,
						Title:      fmt.Sprintf("%s Extracted", strings.Title(strings.ReplaceAll(pattern.ruleType, "_", " "))),
						Content:    content,
						Confidence: pattern.confidence,
						SourcePage: 0,
						Status:     "extracted",
						DocumentID: docID,
						StructuredData: map[string]interface{}{
							"extraction_method": "regex_pattern",
							"pattern_type":      pattern.ruleType,
							"source":            "text_analysis",
						},
					})
				}
			}
		}
	}

	return items, nil
}

// ExtractFromFile extracts knowledge items from a file (placeholder implementation)
func (k *KnowledgeExtractorImpl) ExtractFromFile(ctx context.Context, filePath string, mimeType string, docID string) ([]models.KnowledgeItem, error) {
	// For now, return empty slice - in production this would parse the file
	// based on mimeType (PDF, DOCX, etc.)
	return []models.KnowledgeItem{}, nil
}

// ClassifyKnowledgeItem classifies a knowledge item (placeholder implementation)
func (k *KnowledgeExtractorImpl) ClassifyKnowledgeItem(ctx context.Context, item *models.KnowledgeItem) error {
	// Basic classification based on content patterns
	content := strings.ToLower(item.Content)

	if strings.Contains(content, "must") || strings.Contains(content, "shall") || strings.Contains(content, "required") {
		item.Type = "functional_requirement"
		item.Confidence = 0.9
	} else if strings.Contains(content, "performance") || strings.Contains(content, "latency") {
		item.Type = "performance_requirement"
		item.Confidence = 0.8
	} else if strings.Contains(content, "security") || strings.Contains(content, "auth") {
		item.Type = "security_requirement"
		item.Confidence = 0.9
	}

	return nil
}

// ValidateKnowledgeItem validates a knowledge item
func (k *KnowledgeExtractorImpl) ValidateKnowledgeItem(ctx context.Context, item *models.KnowledgeItem) error {
	if item.Content == "" {
		return fmt.Errorf("knowledge item content cannot be empty")
	}
	if len(item.Content) < 5 {
		return fmt.Errorf("knowledge item content too short")
	}
	if item.Confidence < 0 || item.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	return nil
}

// DocumentValidatorImpl implements DocumentValidator
type DocumentValidatorImpl struct{}

// NewDocumentValidator creates a new document validator instance
func NewDocumentValidator() *DocumentValidatorImpl {
	return &DocumentValidatorImpl{}
}

// ValidateFile validates a file's basic properties
func (v *DocumentValidatorImpl) ValidateFile(ctx context.Context, filePath string, mimeType string) error {
	// Basic MIME type validation
	allowedTypes := []string{
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
		"text/markdown",
		"application/msword",
	}

	for _, allowed := range allowedTypes {
		if mimeType == allowed {
			return nil
		}
	}

	return fmt.Errorf("unsupported file type: %s", mimeType)
}

// ValidateContent validates file content (placeholder implementation)
func (v *DocumentValidatorImpl) ValidateContent(ctx context.Context, content []byte, mimeType string) error {
	if len(content) == 0 {
		return fmt.Errorf("file content is empty")
	}

	// Basic content validation based on MIME type
	switch mimeType {
	case "application/pdf":
		if len(content) < 100 {
			return fmt.Errorf("PDF file too small to be valid")
		}
		// Check for PDF header
		if len(content) >= 4 && string(content[:4]) != "%PDF" {
			return fmt.Errorf("invalid PDF file format")
		}
	case "text/plain", "text/markdown":
		// Convert to string and check for basic text content
		text := string(content)
		if strings.TrimSpace(text) == "" {
			return fmt.Errorf("text file contains no readable content")
		}
	}

	return nil
}

// CheckSecurity performs basic security checks on the file
func (v *DocumentValidatorImpl) CheckSecurity(ctx context.Context, filePath string) error {
	// Basic security checks - in production this would be more comprehensive
	// Check file size, content patterns, etc.

	// For now, just return success - in production:
	// - Check for embedded scripts
	// - Scan for malware signatures
	// - Validate against security policies
	return nil
}

// ValidateSize validates file size limits
func (v *DocumentValidatorImpl) ValidateSize(ctx context.Context, size int64) error {
	maxSize := int64(100 * 1024 * 1024) // 100MB
	if size > maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes", size, maxSize)
	}

	minSize := int64(10) // 10 bytes minimum
	if size < minSize {
		return fmt.Errorf("file size %d bytes is below minimum allowed size %d bytes", size, minSize)
	}

	return nil
}

// SearchEngineImpl implements SearchEngine (basic in-memory implementation)
type SearchEngineImpl struct {
	index map[string][]models.KnowledgeItem
}

// NewSearchEngine creates a new search engine instance
func NewSearchEngine() *SearchEngineImpl {
	return &SearchEngineImpl{
		index: make(map[string][]models.KnowledgeItem),
	}
}

// IndexDocument indexes a document's knowledge items
func (s *SearchEngineImpl) IndexDocument(ctx context.Context, docID string, content string, knowledgeItems []models.KnowledgeItem) error {
	s.index[docID] = knowledgeItems
	return nil
}

// SearchDocuments searches for knowledge items across documents
func (s *SearchEngineImpl) SearchDocuments(ctx context.Context, projectID string, query string) ([]models.KnowledgeItem, error) {
	var results []models.KnowledgeItem
	query = strings.ToLower(query)

	// Simple text search across all indexed documents
	for _, items := range s.index {
		for _, item := range items {
			content := strings.ToLower(item.Content)
			if strings.Contains(content, query) {
				results = append(results, item)
			}
		}
	}

	return results, nil
}

// UpdateIndex updates the search index for a document
func (s *SearchEngineImpl) UpdateIndex(ctx context.Context, docID string, content string, knowledgeItems []models.KnowledgeItem) error {
	s.index[docID] = knowledgeItems
	return nil
}

// DeleteFromIndex removes a document from the search index
func (s *SearchEngineImpl) DeleteFromIndex(ctx context.Context, docID string) error {
	delete(s.index, docID)
	return nil
}
