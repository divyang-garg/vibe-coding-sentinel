// Package models - Document model tests
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDocumentStatus tests DocumentStatus enum functionality
func TestDocumentStatus_IsValid(t *testing.T) {
	tests := []struct {
		status DocumentStatus
		want   bool
	}{
		{DocumentStatusQueued, true},
		{DocumentStatusProcessing, true},
		{DocumentStatusCompleted, true},
		{DocumentStatusFailed, true},
		{DocumentStatusUploaded, true},
		{DocumentStatusDeleted, true},
		{DocumentStatus("invalid"), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.want {
			t.Errorf("DocumentStatus.IsValid() = %v, want %v", got, tt.want)
		}
	}
}

func TestDocumentStatus_MarshalJSON(t *testing.T) {
	status := DocumentStatusCompleted
	data, err := status.MarshalJSON()
	if err != nil {
		t.Fatalf("DocumentStatus.MarshalJSON() error = %v", err)
	}
	if string(data) != `"completed"` {
		t.Errorf("DocumentStatus.MarshalJSON() = %v, want %v", string(data), `"completed"`)
	}
}

func TestDocumentStatus_UnmarshalJSON(t *testing.T) {
	var status DocumentStatus
	data := []byte(`"processing"`)
	err := status.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("DocumentStatus.UnmarshalJSON() error = %v", err)
	}
	if status != DocumentStatusProcessing {
		t.Errorf("DocumentStatus.UnmarshalJSON() = %v, want %v", status, DocumentStatusProcessing)
	}
}

// TestDocument_JSONSerialization tests Document model JSON marshaling
func TestDocument_JSONSerialization(t *testing.T) {
	doc := Document{
		ID:           "doc-456",
		ProjectID:    "project-abc",
		Name:         "design_spec.pdf",
		OriginalName: "Design Spec v1.0.pdf",
		MimeType:     "application/pdf",
		Status:       DocumentStatusQueued,
		Progress:     0,
	}

	data, err := json.Marshal(doc)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"status":"queued"`)

	var unmarshaledDoc Document
	err = json.Unmarshal(data, &unmarshaledDoc)
	assert.NoError(t, err)
	assert.Equal(t, DocumentStatusQueued, unmarshaledDoc.Status)
}

// TestValidateDocument tests Document model validation
func TestValidateDocument(t *testing.T) {
	tests := []struct {
		name     string
		document Document
		wantErr  bool
	}{
		{
			name: "valid document",
			document: Document{
				ID:        "doc-123",
				ProjectID: "proj-1",
				Name:      "test.pdf",
				MimeType:  "application/pdf",
				Status:    DocumentStatusQueued,
				Progress:  0,
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			document: Document{
				ID:        "doc-123",
				ProjectID: "proj-1",
				Name:      "test.pdf",
				MimeType:  "application/pdf",
				Status:    "invalid",
				Progress:  0,
			},
			wantErr: true,
		},
		{
			name: "invalid progress",
			document: Document{
				ID:        "doc-123",
				ProjectID: "proj-1",
				Name:      "test.pdf",
				MimeType:  "application/pdf",
				Status:    DocumentStatusQueued,
				Progress:  101,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDocument(&tt.document)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
