// Package models contains document-related data models.
// This file defines all document domain entities following the data-only principle.
package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// DocumentStatus represents the processing status of a document
type DocumentStatus string

const (
	DocumentStatusQueued     DocumentStatus = "queued"
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusCompleted  DocumentStatus = "completed"
	DocumentStatusFailed     DocumentStatus = "failed"
	DocumentStatusUploaded   DocumentStatus = "uploaded"
	DocumentStatusDeleted    DocumentStatus = "deleted"
)

// String returns the string representation of DocumentStatus
func (s DocumentStatus) String() string {
	return string(s)
}

// IsValid checks if the DocumentStatus is valid
func (s DocumentStatus) IsValid() bool {
	switch s {
	case DocumentStatusQueued, DocumentStatusProcessing, DocumentStatusCompleted, DocumentStatusFailed, DocumentStatusUploaded, DocumentStatusDeleted:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler
func (s DocumentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements json.Unmarshaler
func (s *DocumentStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = DocumentStatus(str)
	if !s.IsValid() {
		return fmt.Errorf("invalid document status: %s", str)
	}
	return nil
}

// Document represents a document uploaded for processing
type Document struct {
	ID            string         `json:"id" db:"id" validate:"required,uuid"`
	ProjectID     string         `json:"project_id" db:"project_id" validate:"required,uuid"`
	Name          string         `json:"name" db:"name" validate:"required,min=1,max=255"`
	OriginalName  string         `json:"original_name" db:"original_name" validate:"required,min=1,max=255"`
	Size          int64          `json:"size" db:"size" validate:"required,min=0"`
	MimeType      string         `json:"mime_type" db:"mime_type" validate:"required"`
	Status        DocumentStatus `json:"status" db:"status" validate:"required"`
	Progress      int            `json:"progress" db:"progress" validate:"min=0,max=100"`
	FilePath      string         `json:"-" db:"file_path" validate:"required"`
	ExtractedText string         `json:"extracted_text,omitempty" db:"extracted_text"`
	Error         string         `json:"error,omitempty" db:"error"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	ProcessedAt   *time.Time     `json:"processed_at,omitempty" db:"processed_at"`
}

// KnowledgeItem represents extracted knowledge from documents
type KnowledgeItem struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`  // Maps to item_type
	Title          string                 `json:"title"` // Extracted title
	Content        string                 `json:"content"`
	Confidence     float64                `json:"confidence"`
	SourcePage     int                    `json:"source_page"`
	Status         string                 `json:"status"`
	StructuredData map[string]interface{} `json:"structured_data,omitempty"`
	DocumentID     string                 `json:"document_id"`
	ApprovedBy     *string                `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time             `json:"approved_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// DocumentUploadRequest represents a request to upload a document
type DocumentUploadRequest struct {
	ProjectID    string `json:"project_id"`
	Name         string `json:"name,omitempty"`
	OriginalName string `json:"original_name,omitempty"`
}

// DocumentUploadResponse represents the response after uploading a document
type DocumentUploadResponse struct {
	Document *Document `json:"document"`
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
}

// DocumentProcessingResult represents the result of document processing
type DocumentProcessingResult struct {
	DocumentID     string          `json:"document_id"`
	Status         string          `json:"status"`
	KnowledgeItems []KnowledgeItem `json:"knowledge_items,omitempty"`
	Error          string          `json:"error,omitempty"`
	ProcessedAt    time.Time       `json:"processed_at"`
	Success        bool            `json:"success"`
}
