// Package services provides tests for comprehensive code analysis
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sentinel-hub-api/models"
)

// mockASTService provides a mock implementation of ASTService
type mockASTService struct{}

func (m *mockASTService) AnalyzeAST(ctx context.Context, req models.ASTAnalysisRequest) (*models.ASTAnalysisResponse, error) {
	return &models.ASTAnalysisResponse{}, nil
}

func (m *mockASTService) AnalyzeMultiFile(ctx context.Context, req models.MultiFileASTRequest) (*models.MultiFileASTResponse, error) {
	return &models.MultiFileASTResponse{}, nil
}

func (m *mockASTService) AnalyzeSecurity(ctx context.Context, req models.SecurityASTRequest) (*models.SecurityASTResponse, error) {
	return &models.SecurityASTResponse{}, nil
}

func (m *mockASTService) AnalyzeCrossFile(ctx context.Context, req models.CrossFileASTRequest) (*models.CrossFileASTResponse, error) {
	return &models.CrossFileASTResponse{}, nil
}

func (m *mockASTService) GetSupportedAnalyses(ctx context.Context) (*models.SupportedAnalysesResponse, error) {
	return &models.SupportedAnalysesResponse{}, nil
}

// mockKnowledgeService provides a mock implementation of KnowledgeService
type mockKnowledgeService struct{}

func (m *mockKnowledgeService) RunGapAnalysis(ctx context.Context, req GapAnalysisRequest) (*GapAnalysisReport, error) {
	return nil, nil
}

func (m *mockKnowledgeService) ListKnowledgeItems(ctx context.Context, req ListKnowledgeItemsRequest) ([]KnowledgeItem, error) {
	return nil, nil
}

func (m *mockKnowledgeService) CreateKnowledgeItem(ctx context.Context, item KnowledgeItem) (*KnowledgeItem, error) {
	return nil, nil
}

func (m *mockKnowledgeService) GetKnowledgeItem(ctx context.Context, id string) (*KnowledgeItem, error) {
	return nil, nil
}

func (m *mockKnowledgeService) UpdateKnowledgeItem(ctx context.Context, id string, item KnowledgeItem) (*KnowledgeItem, error) {
	return nil, nil
}

func (m *mockKnowledgeService) DeleteKnowledgeItem(ctx context.Context, id string) error {
	return nil
}

func (m *mockKnowledgeService) GetBusinessContext(ctx context.Context, req BusinessContextRequest) (*BusinessContextResponse, error) {
	return &BusinessContextResponse{
		Rules:        []KnowledgeItem{},
		Entities:     []KnowledgeItem{},
		UserJourneys: []KnowledgeItem{},
	}, nil
}

func (m *mockKnowledgeService) SyncKnowledge(ctx context.Context, req SyncKnowledgeRequest) (*SyncKnowledgeResponse, error) {
	return nil, nil
}

// mockLogger provides a mock implementation of Logger
type mockLogger struct{}

func (m *mockLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {}
func (m *mockLogger) Error(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {}
func (m *mockLogger) Info(ctx context.Context, msg string, fields ...map[string]interface{}) {}
func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {}

func TestComprehensiveAnalysisService_ExecuteAnalysis(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	t.Run("success_shallow_mode", func(t *testing.T) {
		// Given
		testDir := createTestDir(t)
		ctx := context.Background()
		req := ComprehensiveAnalysisRequest{
			ProjectID:    "test-project",
			CodebasePath: testDir,
			Mode:         "auto",
			Depth:        "shallow",
		}

		// When
		result, err := service.ExecuteAnalysis(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-project", result.ProjectID)
		assert.Equal(t, "auto", result.Mode)
		assert.Equal(t, "shallow", result.Depth)
		assert.NotNil(t, result.Layers)
		assert.NotEmpty(t, result.AnalyzedAt)
	})

	t.Run("success_deep_mode", func(t *testing.T) {
		// Given
		testDir := createTestDir(t)
		ctx := context.Background()
		req := ComprehensiveAnalysisRequest{
			ProjectID:    "test-project",
			CodebasePath: testDir,
			Mode:         "auto",
			Depth:        "deep",
		}

		// When
		result, err := service.ExecuteAnalysis(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "deep", result.Depth)
	})

	t.Run("missing_project_id", func(t *testing.T) {
		// Given
		ctx := context.Background()
		req := ComprehensiveAnalysisRequest{
			Mode: "auto",
			Depth: "shallow",
		}

		// When
		result, err := service.ExecuteAnalysis(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "project_id is required")
	})

	t.Run("with_business_context", func(t *testing.T) {
		// Given
		testDir := createTestDir(t)
		ctx := context.Background()
		req := ComprehensiveAnalysisRequest{
			ProjectID:              "test-project",
			CodebasePath:            testDir,
			Feature:                 "test-feature",
			Mode:                    "auto",
			Depth:                   "shallow",
			IncludeBusinessContext:  true,
		}

		// When
		result, err := service.ExecuteAnalysis(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.BusinessContext)
	})

	t.Run("manual_mode_with_files", func(t *testing.T) {
		// Given
		ctx := context.Background()
		req := ComprehensiveAnalysisRequest{
			ProjectID: "test-project",
			Mode:      "manual",
			Depth:     "shallow",
			Files:     []string{"handler.go", "service.go"},
		}

		// When
		result, err := service.ExecuteAnalysis(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "manual", result.Mode)
	})
}

func TestComprehensiveAnalysisService_detectLayers(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	t.Run("auto_mode", func(t *testing.T) {
		// Given
		testDir := createTestDir(t)
		ctx := context.Background()

		// When
		layers, err := service.detectLayers(ctx, testDir, "auto", nil)

		// Then
		assert.NoError(t, err)
		// Should return default layers if none detected
		assert.NotEmpty(t, layers)
	})

	t.Run("manual_mode_with_files", func(t *testing.T) {
		// Given
		ctx := context.Background()
		files := []string{"handler.go", "service.go", "test.go"}

		// When
		layers, err := service.detectLayers(ctx, ".", "manual", files)

		// Then
		assert.NoError(t, err)
		assert.NotEmpty(t, layers)
	})

	t.Run("invalid_path", func(t *testing.T) {
		// Given
		ctx := context.Background()
		invalidPath := "/nonexistent/path/that/does/not/exist"

		// When
		layers, err := service.detectLayers(ctx, invalidPath, "auto", nil)

		// Then
		assert.Error(t, err)
		assert.Empty(t, layers)
	})
}

func TestComprehensiveAnalysisService_analyzeLayer(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	t.Run("shallow_analysis", func(t *testing.T) {
		// Given
		testDir := createTestDir(t)
		ctx := context.Background()
		layer := "api"
		depth := "shallow"

		// When
		result, err := service.analyzeLayer(ctx, testDir, layer, depth, nil)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, layer, result.Layer)
		assert.NotEmpty(t, result.AnalyzedAt)
	})

	t.Run("deep_analysis", func(t *testing.T) {
		// Given
		testDir := createTestDir(t)
		ctx := context.Background()
		layer := "logic"
		depth := "deep"

		// When
		result, err := service.analyzeLayer(ctx, testDir, layer, depth, nil)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, layer, result.Layer)
	})
}

func TestComprehensiveAnalysisService_detectLanguage(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	tests := []struct {
		name     string
		file     string
		expected string
	}{
		{"go_file", "test.go", "go"},
		{"javascript_file", "test.js", "javascript"},
		{"typescript_file", "test.ts", "typescript"},
		{"python_file", "test.py", "python"},
		{"unknown_file", "test.xyz", "go"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectLanguage(tt.file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComprehensiveAnalysisService_classifyDependency(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	tests := []struct {
		name     string
		dep      string
		expected string
	}{
		{"internal_relative", "./local/package", "internal"},
		{"internal_parent", "../parent/package", "internal"},
		{"external_github", "github.com/user/repo", "external"},
		{"external_npm", "@npm/package", "external"},
		{"standard_library", "fmt", "standard_library"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.classifyDependency(tt.dep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComprehensiveAnalysisService_aggregateResults(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	t.Run("multiple_layers", func(t *testing.T) {
		// Given
		layers := []LayerAnalysis{
			{
				Layer:        "api",
				QualityScore: 85.0,
				AnalyzedAt:   time.Now().Format(time.RFC3339),
			},
			{
				Layer:        "logic",
				QualityScore: 90.0,
				AnalyzedAt:   time.Now().Format(time.RFC3339),
			},
		}
		req := ComprehensiveAnalysisRequest{
			ProjectID: "test-project",
			Mode:      "auto",
			Depth:     "shallow",
		}

		// When
		result := service.aggregateResults(layers, req)

		// Then
		assert.NotNil(t, result)
		assert.Equal(t, "test-project", result.ProjectID)
		assert.Equal(t, 87.5, result.OverallScore) // Average of 85 and 90
		assert.Equal(t, len(layers), len(result.Layers))
	})

	t.Run("empty_layers", func(t *testing.T) {
		// Given
		layers := []LayerAnalysis{}
		req := ComprehensiveAnalysisRequest{
			ProjectID: "test-project",
		}

		// When
		result := service.aggregateResults(layers, req)

		// Then
		assert.NotNil(t, result)
		assert.Equal(t, 0.0, result.OverallScore)
	})
}

func TestComprehensiveAnalysisService_contextCancellation(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	t.Run("cancelled_context", func(t *testing.T) {
		// Given
		testDir := createTestDir(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := ComprehensiveAnalysisRequest{
			ProjectID:    "test-project",
			CodebasePath: testDir,
			Mode:         "auto",
			Depth:        "shallow",
		}

		// When
		result, err := service.ExecuteAnalysis(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestComprehensiveAnalysisService_getBusinessContext(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	t.Run("success", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project"
		feature := "test-feature"

		// When
		result, err := service.getBusinessContext(ctx, projectID, feature)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("nil_knowledge_service", func(t *testing.T) {
		// Given
		serviceNoKnowledge := NewComprehensiveAnalysisService(
			&mockASTService{},
			nil, // No knowledge service
			&mockLogger{},
		)
		ctx := context.Background()

		// When
		result, err := serviceNoKnowledge.getBusinessContext(ctx, "test-project", "test-feature")

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "knowledge service not available")
	})
}

func TestComprehensiveAnalysisService_fileBelongsToLayer(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	tests := []struct {
		name     string
		file     string
		layer    string
		expected bool
	}{
		{"ui_file", "component.jsx", "ui", true},
		{"api_file", "handler.go", "api", true},
		{"database_file", "migration.sql", "database", true},
		{"logic_file", "service.go", "logic", true},
		{"test_file", "service_test.go", "tests", true},
		{"mismatch", "service.go", "ui", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.fileBelongsToLayer(tt.file, tt.layer)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a temporary test directory
func createTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "comprehensive_test_*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

func TestComprehensiveAnalysisService_findLayerFiles(t *testing.T) {
	service := NewComprehensiveAnalysisService(
		&mockASTService{},
		&mockKnowledgeService{},
		&mockLogger{},
	)

	t.Run("with_specified_files", func(t *testing.T) {
		// Given
		specifiedFiles := []string{"handler.go", "service.go", "component.jsx"}

		// When
		apiFiles := service.findLayerFiles(".", "api", specifiedFiles)

		// Then
		assert.NotEmpty(t, apiFiles)
	})

	t.Run("scan_codebase", func(t *testing.T) {
		// Given
		testDir := createTestDir(t)
		testFile := filepath.Join(testDir, "handler.go")
		err := os.WriteFile(testFile, []byte("package main"), 0644)
		require.NoError(t, err)

		// When
		files := service.findLayerFiles(testDir, "api", nil)

		// Then
		// Should find files matching layer patterns
		assert.NotNil(t, files)
	})
}
