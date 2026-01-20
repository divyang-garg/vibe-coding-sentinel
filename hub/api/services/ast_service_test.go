// Package services - AST service tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"testing"

	"sentinel-hub-api/models"
)

func TestASTServiceImpl_AnalyzeAST(t *testing.T) {
	service := NewASTService()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := models.ASTAnalysisRequest{
			Code:     "package main\nfunc test() {}\n",
			Language: "go",
			Analyses: []string{"duplicates"},
		}

		result, err := service.AnalyzeAST(ctx, req)
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		if result == nil {
			t.Fatal("AnalyzeAST returned nil result")
		}
		if result.Language != "go" {
			t.Errorf("Expected language 'go', got '%s'", result.Language)
		}
	})

	t.Run("missing_code", func(t *testing.T) {
		req := models.ASTAnalysisRequest{
			Language: "go",
		}

		_, err := service.AnalyzeAST(ctx, req)
		if err == nil {
			t.Error("Expected error for missing code")
		}
	})

	t.Run("missing_language", func(t *testing.T) {
		req := models.ASTAnalysisRequest{
			Code: "package main\n",
		}

		_, err := service.AnalyzeAST(ctx, req)
		if err == nil {
			t.Error("Expected error for missing language")
		}
	})
}

func TestASTServiceImpl_AnalyzeMultiFile(t *testing.T) {
	service := NewASTService()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := models.MultiFileASTRequest{
			Files: []models.FileInput{
				{
					Path:     "file1.go",
					Content:  "package main\nfunc test() {}\n",
					Language: "go",
				},
				{
					Path:     "file2.go",
					Content:  "package main\nfunc test2() {}\n",
					Language: "go",
				},
			},
		}

		result, err := service.AnalyzeMultiFile(ctx, req)
		if err != nil {
			t.Fatalf("AnalyzeMultiFile failed: %v", err)
		}
		if result == nil {
			t.Fatal("AnalyzeMultiFile returned nil result")
		}
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(result.Files))
		}
	})

	t.Run("empty_files", func(t *testing.T) {
		req := models.MultiFileASTRequest{
			Files: []models.FileInput{},
		}

		_, err := service.AnalyzeMultiFile(ctx, req)
		if err == nil {
			t.Error("Expected error for empty files")
		}
	})
}

func TestASTServiceImpl_AnalyzeSecurity(t *testing.T) {
	service := NewASTService()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := models.SecurityASTRequest{
			Code:     "package main\nfunc test() { db.Query(\"SELECT * FROM users WHERE id = \" + id) }\n",
			Language: "go",
		}

		result, err := service.AnalyzeSecurity(ctx, req)
		if err != nil {
			t.Fatalf("AnalyzeSecurity failed: %v", err)
		}
		if result == nil {
			t.Fatal("AnalyzeSecurity returned nil result")
		}
	})

	t.Run("missing_code", func(t *testing.T) {
		req := models.SecurityASTRequest{
			Language: "go",
		}

		_, err := service.AnalyzeSecurity(ctx, req)
		if err == nil {
			t.Error("Expected error for missing code")
		}
	})
}

func TestASTServiceImpl_AnalyzeCrossFile(t *testing.T) {
	service := NewASTService()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := models.CrossFileASTRequest{
			Files: []models.FileInput{
				{
					Path:     "file1.go",
					Content:  "package main\nfunc Exported() {}\n",
					Language: "go",
				},
				{
					Path:     "file2.go",
					Content:  "package main\nimport \"fmt\"\n",
					Language: "go",
				},
			},
		}

		result, err := service.AnalyzeCrossFile(ctx, req)
		if err != nil {
			t.Fatalf("AnalyzeCrossFile failed: %v", err)
		}
		if result == nil {
			t.Fatal("AnalyzeCrossFile returned nil result")
		}
	})

	t.Run("empty_files", func(t *testing.T) {
		req := models.CrossFileASTRequest{
			Files: []models.FileInput{},
		}

		_, err := service.AnalyzeCrossFile(ctx, req)
		if err == nil {
			t.Error("Expected error for empty files")
		}
	})
}

func TestASTServiceImpl_GetSupportedAnalyses(t *testing.T) {
	service := NewASTService()
	ctx := context.Background()

	result, err := service.GetSupportedAnalyses(ctx)
	if err != nil {
		t.Fatalf("GetSupportedAnalyses failed: %v", err)
	}
	if result == nil {
		t.Fatal("GetSupportedAnalyses returned nil result")
	}
	if len(result.Languages) == 0 {
		t.Error("Expected at least one supported language")
	}
	if len(result.Analyses) == 0 {
		t.Error("Expected at least one supported analysis")
	}
}
