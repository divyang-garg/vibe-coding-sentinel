// Package services - Document integration tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDocumentRepository implements DocumentRepository for testing
type MockDocumentRepository struct {
	mock.Mock
	documents map[string]*models.Document
}

func NewMockDocumentRepository() *MockDocumentRepository {
	return &MockDocumentRepository{
		documents: make(map[string]*models.Document),
	}
}

func (m *MockDocumentRepository) Save(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	if args.Error(0) == nil {
		m.documents[doc.ID] = doc
	}
	return args.Error(0)
}

func (m *MockDocumentRepository) FindByID(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if doc, ok := m.documents[id]; ok {
		return doc, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) FindByProjectID(ctx context.Context, projectID string) ([]models.Document, error) {
	args := m.Called(ctx, projectID)
	var docs []models.Document
	for _, doc := range m.documents {
		if doc.ProjectID == projectID {
			docs = append(docs, *doc)
		}
	}
	return docs, args.Error(1)
}

func (m *MockDocumentRepository) Update(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	if args.Error(0) == nil {
		m.documents[doc.ID] = doc
	}
	return args.Error(0)
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		delete(m.documents, id)
	}
	return args.Error(0)
}

func (m *MockDocumentRepository) UpdateStatus(ctx context.Context, id string, status string, progress int, errorMsg string) error {
	args := m.Called(ctx, id, status, progress, errorMsg)
	if doc, ok := m.documents[id]; ok && args.Error(0) == nil {
		doc.Status = models.DocumentStatus(status)
		doc.Progress = progress
		if errorMsg != "" {
			doc.Error = errorMsg
		}
	}
	return args.Error(0)
}

func (m *MockDocumentRepository) UpdateProcessedAt(ctx context.Context, id string, processedAt *time.Time) error {
	args := m.Called(ctx, id, processedAt)
	if doc, ok := m.documents[id]; ok && args.Error(0) == nil {
		doc.ProcessedAt = processedAt
	}
	return args.Error(0)
}

// MockKnowledgeExtractor implements KnowledgeExtractor for testing
type MockKnowledgeExtractor struct {
	mock.Mock
}

func (m *MockKnowledgeExtractor) ExtractFromText(ctx context.Context, text string, docID string) ([]models.KnowledgeItem, error) {
	args := m.Called(ctx, text, docID)
	return args.Get(0).([]models.KnowledgeItem), args.Error(1)
}

func (m *MockKnowledgeExtractor) ExtractFromFile(ctx context.Context, filePath string, mimeType string, docID string) ([]models.KnowledgeItem, error) {
	args := m.Called(ctx, filePath, mimeType, docID)
	return args.Get(0).([]models.KnowledgeItem), args.Error(1)
}

func (m *MockKnowledgeExtractor) ClassifyKnowledgeItem(ctx context.Context, item *models.KnowledgeItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockKnowledgeExtractor) ValidateKnowledgeItem(ctx context.Context, item *models.KnowledgeItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

// MockDocumentValidator implements DocumentValidator for testing
type MockDocumentValidator struct {
	mock.Mock
}

func (m *MockDocumentValidator) ValidateFile(ctx context.Context, filePath string, mimeType string) error {
	args := m.Called(ctx, filePath, mimeType)
	return args.Error(0)
}

func (m *MockDocumentValidator) ValidateContent(ctx context.Context, content []byte, mimeType string) error {
	args := m.Called(ctx, content, mimeType)
	return args.Error(0)
}

func (m *MockDocumentValidator) CheckSecurity(ctx context.Context, filePath string) error {
	args := m.Called(ctx, filePath)
	return args.Error(0)
}

func (m *MockDocumentValidator) ValidateSize(ctx context.Context, size int64) error {
	args := m.Called(ctx, size)
	return args.Error(0)
}

// MockSearchEngine implements SearchEngine for testing
type MockSearchEngine struct {
	mock.Mock
}

func (m *MockSearchEngine) IndexDocument(ctx context.Context, docID string, content string, knowledgeItems []models.KnowledgeItem) error {
	args := m.Called(ctx, docID, content, knowledgeItems)
	return args.Error(0)
}

func (m *MockSearchEngine) SearchDocuments(ctx context.Context, projectID string, query string) ([]models.KnowledgeItem, error) {
	args := m.Called(ctx, projectID, query)
	return args.Get(0).([]models.KnowledgeItem), args.Error(1)
}

func (m *MockSearchEngine) UpdateIndex(ctx context.Context, docID string, content string, knowledgeItems []models.KnowledgeItem) error {
	args := m.Called(ctx, docID, content, knowledgeItems)
	return args.Error(0)
}

func (m *MockSearchEngine) DeleteFromIndex(ctx context.Context, docID string) error {
	args := m.Called(ctx, docID)
	return args.Error(0)
}

// createTestFile creates a temporary test file with content
func createTestFile(t *testing.T, content string, ext string) string {
	tmpFile, err := os.CreateTemp("", "test-*."+ext)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.WriteString(content)
	tmpFile.Close()
	return tmpFile.Name()
}

// TestDocumentUploadProcessExtract tests the full document workflow
func TestDocumentUploadProcessExtract(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project-123"

	// Setup mocks
	mockRepo := NewMockDocumentRepository()
	mockExtractor := &MockKnowledgeExtractor{}
	mockValidator := &MockDocumentValidator{}
	mockSearch := &MockSearchEngine{}

	service := NewDocumentService(mockRepo, mockExtractor, mockValidator, mockSearch)

	t.Run("text file upload and process", func(t *testing.T) {
		// Create test file
		testContent := "This is a test document with some content."
		filePath := createTestFile(t, testContent, "txt")
		defer os.Remove(filePath)

		// Get file info
		fileInfo, err := os.Stat(filePath)
		assert.NoError(t, err)

		// Setup mock expectations
		mockValidator.On("ValidateFile", ctx, filePath, "text/plain").Return(nil)

		docID := "doc-test-123"
		doc := &models.Document{
			ID:          docID,
			ProjectID:   projectID,
			Name:        "test.txt",
			OriginalName: "test.txt",
			FilePath:    filePath,
			MimeType:    "text/plain",
			Size:        fileInfo.Size(),
			Status:      models.DocumentStatusUploaded,
			Progress:    0,
			CreatedAt:   time.Now(),
		}

		mockRepo.On("Save", ctx, mock.AnythingOfType("*models.Document")).Return(nil).Run(func(args mock.Arguments) {
			savedDoc := args.Get(1).(*models.Document)
			savedDoc.ID = docID
		})

		// Upload document
		uploadReq := models.DocumentUploadRequest{
			ProjectID:    projectID,
			OriginalName: "test.txt",
			Name:         "test.txt",
		}

		uploadResp, err := service.UploadDocument(ctx, uploadReq, filePath, "text/plain")
		assert.NoError(t, err)
		assert.NotNil(t, uploadResp)
		
		// Capture actual docID from upload response
		actualDocID := uploadResp.Document.ID
		assert.NotEmpty(t, actualDocID)

		// Update doc with actual ID for processing
		doc.ID = actualDocID

		// Setup processing mocks using actual docID
		mockRepo.On("FindByID", ctx, actualDocID).Return(doc, nil)
		mockRepo.On("UpdateStatus", ctx, actualDocID, string(models.DocumentStatusProcessing), 10, "").Return(nil)
		mockRepo.On("Save", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

		knowledgeItems := []models.KnowledgeItem{
			{
				ID:         "ki-1",
				DocumentID: actualDocID,
				Type:       "text",
				Title:      "Test Knowledge",
				Content:    testContent,
				Confidence: 0.9,
				Status:     "pending",
			},
		}

		mockExtractor.On("ExtractFromText", ctx, testContent, actualDocID).Return(knowledgeItems, nil)
		mockExtractor.On("ClassifyKnowledgeItem", ctx, mock.AnythingOfType("*models.KnowledgeItem")).Return(nil)
		mockExtractor.On("ValidateKnowledgeItem", ctx, mock.AnythingOfType("*models.KnowledgeItem")).Return(nil)
		mockRepo.On("UpdateStatus", ctx, actualDocID, string(models.DocumentStatusCompleted), 100, "").Return(nil)
		mockRepo.On("UpdateProcessedAt", ctx, actualDocID, mock.AnythingOfType("*time.Time")).Return(nil)
		mockSearch.On("IndexDocument", ctx, actualDocID, testContent, knowledgeItems).Return(nil)

		// Process document
		result, err := service.ProcessDocument(ctx, actualDocID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, actualDocID, result.DocumentID)
		assert.Greater(t, len(result.KnowledgeItems), 0)

		// Verify all mocks were called
		mockRepo.AssertExpectations(t)
		mockExtractor.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
		mockSearch.AssertExpectations(t)
	})

	t.Run("invalid file type", func(t *testing.T) {
		filePath := createTestFile(t, "content", "txt")
		defer os.Remove(filePath)

		mockValidator.On("ValidateFile", ctx, filePath, "application/unknown").Return(nil)

		uploadReq := models.DocumentUploadRequest{
			ProjectID:    projectID,
			OriginalName: "test.unknown",
			Name:         "test.unknown",
		}

		uploadResp, err := service.UploadDocument(ctx, uploadReq, filePath, "application/unknown")
		assert.NoError(t, err) // Upload succeeds
		assert.NotNil(t, uploadResp)

		// Processing should handle unknown type gracefully
		doc := uploadResp.Document
		unknownText := fmt.Sprintf("Document content extraction not implemented for %s", doc.MimeType)
		mockRepo.On("FindByID", ctx, doc.ID).Return(&doc, nil)
		mockRepo.On("UpdateStatus", ctx, doc.ID, string(models.DocumentStatusProcessing), 10, "").Return(nil)
		mockRepo.On("Save", ctx, mock.AnythingOfType("*models.Document")).Return(nil)
		mockExtractor.On("ExtractFromText", ctx, unknownText, doc.ID).Return([]models.KnowledgeItem{}, nil)
		mockRepo.On("UpdateStatus", ctx, doc.ID, string(models.DocumentStatusCompleted), 100, "").Return(nil)
		mockRepo.On("UpdateProcessedAt", ctx, doc.ID, mock.AnythingOfType("*time.Time")).Return(nil)
		mockSearch.On("IndexDocument", ctx, doc.ID, unknownText, []models.KnowledgeItem{}).Return(nil)

		result, err := service.ProcessDocument(ctx, doc.ID)
		// Should complete but with placeholder text for unknown types
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestDocumentErrorHandling tests error cases in document processing
func TestDocumentErrorHandling(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project-123"

	mockRepo := NewMockDocumentRepository()
	mockExtractor := &MockKnowledgeExtractor{}
	mockValidator := &MockDocumentValidator{}
	mockSearch := &MockSearchEngine{}

	service := NewDocumentService(mockRepo, mockExtractor, mockValidator, mockSearch)

	t.Run("document not found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, "nonexistent").Return((*models.Document)(nil), fmt.Errorf("document not found"))

		_, err := service.ProcessDocument(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document not found")
	})

	t.Run("corrupted file", func(t *testing.T) {
		filePath := createTestFile(t, "corrupted content", "txt")
		defer os.Remove(filePath)

		docID := "doc-corrupted-123"
		doc := &models.Document{
			ID:        docID,
			ProjectID: projectID,
			FilePath:  filePath,
			MimeType:  "text/plain",
			Status:    models.DocumentStatusUploaded,
		}

		mockRepo.On("FindByID", ctx, docID).Return(doc, nil)
		mockRepo.On("UpdateStatus", ctx, docID, string(models.DocumentStatusProcessing), 10, "").Return(nil)

		// Simulate file read error by removing file
		os.Remove(filePath)

		mockRepo.On("UpdateStatus", ctx, docID, string(models.DocumentStatusFailed), 0, mock.AnythingOfType("string")).Return(nil)

		_, err := service.ProcessDocument(ctx, docID)
		assert.Error(t, err)
	})
}
