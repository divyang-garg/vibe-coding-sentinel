// Package repository contains knowledge extraction and document processing implementations.
package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"sentinel-hub-api/models"

	pdf "github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
	"github.com/xuri/excelize/v2"
)

// KnowledgeExtractorImpl implements KnowledgeExtractor
type KnowledgeExtractorImpl struct {
	llmExtractor *LLMExtractor
}

// NewKnowledgeExtractor creates a new knowledge extractor instance
func NewKnowledgeExtractor() *KnowledgeExtractorImpl {
	return &KnowledgeExtractorImpl{
		llmExtractor: NewLLMExtractor(),
	}
}

// ExtractFromText extracts knowledge items from plain text
func (k *KnowledgeExtractorImpl) ExtractFromText(ctx context.Context, text string, docID string) ([]models.KnowledgeItem, error) {
	var items []models.KnowledgeItem

	// Try LLM extraction first
	if k.llmExtractor.enabled {
		llmRules, err := k.llmExtractor.ExtractWithLLM(ctx, text, docID)
		if err == nil && len(llmRules) > 0 {
			// Convert LLM results to knowledge items
			for i, rule := range llmRules {
				item := models.KnowledgeItem{
					ID:         getString(rule, "id", fmt.Sprintf("ki_%s_llm_%d", docID, i+1)),
					Type:       "business_rule",
					Title:      getString(rule, "title", ""),
					Content:    getString(rule, "description", ""),
					Confidence: getFloat64(rule, "confidence", 0.8),
					SourcePage: 0,
					Status:     "extracted",
					DocumentID: docID,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
					StructuredData: map[string]interface{}{
						"extraction_method": "llm",
					},
				}
				items = append(items, item)
			}
			if len(items) > 0 {
				return items, nil
			}
		}
	}

	// Fallback to regex extraction
	return k.extractWithRegex(ctx, text, docID), nil
}

// extractWithRegex performs regex-based extraction (fallback)
func (k *KnowledgeExtractorImpl) extractWithRegex(ctx context.Context, text string, docID string) []models.KnowledgeItem {
	var items []models.KnowledgeItem

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
				if content != "" && len(content) > 10 {
					title := strings.ReplaceAll(pattern.ruleType, "_", " ")
					if len(title) > 0 {
						title = strings.ToUpper(title[:1]) + title[1:]
					}
					items = append(items, models.KnowledgeItem{
						ID:         fmt.Sprintf("ki_%s_%d", docID, len(items)+1),
						Type:       pattern.ruleType,
						Title:      fmt.Sprintf("%s Extracted", title),
						Content:    content,
						Confidence: pattern.confidence,
						SourcePage: 0,
						Status:     "extracted",
						DocumentID: docID,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
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

	return items
}

func getString(m map[string]interface{}, key string, defaultVal string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getFloat64(m map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return defaultVal
}

// ExtractFromFile extracts knowledge items from a file
func (k *KnowledgeExtractorImpl) ExtractFromFile(ctx context.Context, filePath string, mimeType string, docID string) ([]models.KnowledgeItem, error) {
	// Parse file based on MIME type
	text, err := parseFileContent(filePath, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	if text == "" {
		return []models.KnowledgeItem{}, nil
	}

	// Extract using text extraction
	return k.ExtractFromText(ctx, text, docID)
}

// parseFileContent extracts text from various file types
func parseFileContent(filePath string, mimeType string) (string, error) {
	switch mimeType {
	case "text/plain", "text/markdown":
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	case "application/pdf":
		return parsePDF(filePath)
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/msword":
		return parseDOCX(filePath)
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/vnd.ms-excel":
		return parseXLSX(filePath)
	default:
		// Try to determine by file extension
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".pdf":
			return parsePDF(filePath)
		case ".docx":
			return parseDOCX(filePath)
		case ".xlsx", ".xls":
			return parseXLSX(filePath)
		case ".txt", ".md":
			content, err := os.ReadFile(filePath)
			if err != nil {
				return "", err
			}
			return string(content), nil
		default:
			return "", fmt.Errorf("unsupported file type: %s", mimeType)
		}
	}
}

// parsePDF extracts text from PDF files
func parsePDF(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	reader, err := pdf.NewReader(f, info.Size())
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var text strings.Builder
	fontMap := make(map[string]*pdf.Font)
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		pageText, _ := page.GetPlainText(fontMap)
		text.WriteString(pageText)
		if i < reader.NumPage() {
			text.WriteString("\n")
		}
	}

	return text.String(), nil
}

// parseDOCX extracts text from DOCX files
func parseDOCX(filePath string) (string, error) {
	doc, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX file: %w", err)
	}
	defer doc.Close()

	return doc.Editable().GetContent(), nil
}

// parseXLSX extracts text from Excel XLSX files
func parseXLSX(filePath string) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	var text strings.Builder

	// Iterate all sheets
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return "", fmt.Errorf("Excel file has no sheets")
	}

	for _, sheetName := range sheetList {
		text.WriteString(fmt.Sprintf("## Sheet: %s\n\n", sheetName))

		rows, err := f.GetRows(sheetName)
		if err != nil {
			// Skip sheets that can't be read, continue with others
			text.WriteString(fmt.Sprintf("(Error reading sheet: %v)\n\n", err))
			continue
		}

		if len(rows) == 0 {
			text.WriteString("(Sheet is empty)\n\n")
			continue
		}

		// First row as header
		if len(rows) > 0 {
			text.WriteString("| ")
			for _, cell := range rows[0] {
				text.WriteString(cell + " | ")
			}
			text.WriteString("\n|")
			for range rows[0] {
				text.WriteString("---|")
			}
			text.WriteString("\n")

			// Data rows
			for i := 1; i < len(rows); i++ {
				text.WriteString("| ")
				for _, cell := range rows[i] {
					text.WriteString(cell + " | ")
				}
				text.WriteString("\n")
			}
		}
		text.WriteString("\n")
	}

	return text.String(), nil
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
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // XLSX
		"application/vnd.ms-excel", // XLS
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
