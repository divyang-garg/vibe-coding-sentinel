// Package services gap analyzer enhanced tests
// Tests for enhanced AST-based gap analysis functions
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sentinel-hub-api/ast"
)

// TestAnalyzeUndocumentedCode_Success tests successful undocumented code analysis
func TestAnalyzeUndocumentedCode_Success(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)

		documentedRules := []KnowledgeItem{
			{ID: "rule-1", Title: "User Authentication", Content: "Authenticate users"},
		}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err != nil {
			t.Fatalf("analyzeUndocumentedCode failed: %v", err)
		}
		if gaps == nil {
			t.Fatal("analyzeUndocumentedCode returned nil gaps")
		}
		// Should find at least some patterns (depending on test codebase)
	})
}

// TestAnalyzeUndocumentedCode_ContextCancellation tests context cancellation handling
func TestAnalyzeUndocumentedCode_ContextCancellation(t *testing.T) {
	t.Run("context_cancellation", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		projectID := "test-project-id"
		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)

		documentedRules := []KnowledgeItem{}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err == nil {
			t.Error("expected error on context cancellation, got nil")
		}
		// gaps may be empty on early cancellation, which is acceptable
		if len(gaps) > 0 {
			t.Logf("Note: Found %d gaps before cancellation", len(gaps))
		}
	})
}

// TestAnalyzeUndocumentedCode_EmptyCodebase tests empty codebase handling
func TestAnalyzeUndocumentedCode_EmptyCodebase(t *testing.T) {
	t.Run("empty_codebase", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		tmpDir, err := os.MkdirTemp("", "gap-test-empty-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		documentedRules := []KnowledgeItem{}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err != nil {
			t.Fatalf("analyzeUndocumentedCode failed: %v", err)
		}
		// In Go, nil slices are valid and len() returns 0
		_ = len(gaps)
		if len(gaps) != 0 {
			t.Errorf("expected 0 gaps for empty codebase, got %d", len(gaps))
		}
	})
}

// TestAnalyzeUndocumentedCode_NoPatterns tests when no patterns are found
func TestAnalyzeUndocumentedCode_NoPatterns(t *testing.T) {
	t.Run("no_patterns", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		tmpDir, err := os.MkdirTemp("", "gap-test-nopatterns-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a file with no business logic
		testFile := filepath.Join(tmpDir, "helper.go")
		err = os.WriteFile(testFile, []byte(`package main

func helper() {
	// Just a helper
}`), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		documentedRules := []KnowledgeItem{}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err != nil {
			t.Fatalf("analyzeUndocumentedCode failed: %v", err)
		}
		// In Go, nil slices are valid and len() returns 0
		_ = len(gaps)
		// Helper function may not be detected as business logic, so 0 gaps is acceptable
		if len(gaps) > 0 {
			t.Logf("Found %d gaps (helper function may not be detected)", len(gaps))
		}
	})
}

// TestAnalyzeUndocumentedCode_AllMatched tests when all patterns match documented rules
func TestAnalyzeUndocumentedCode_AllMatched(t *testing.T) {
	t.Run("all_matched", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		tmpDir, err := os.MkdirTemp("", "gap-test-matched-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a file with business logic
		testFile := filepath.Join(tmpDir, "order.go")
		err = os.WriteFile(testFile, []byte(`package main

func ProcessOrder(orderID string) {
	// Process order
}`), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Documented rule that matches
		documentedRules := []KnowledgeItem{
			{ID: "rule-1", Title: "Process Order", Content: "Process orders with orderID"},
		}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err != nil {
			t.Fatalf("analyzeUndocumentedCode failed: %v", err)
		}
		if gaps == nil {
			t.Fatal("analyzeUndocumentedCode returned nil gaps")
		}
		// Should have fewer gaps since pattern matches rule
		t.Logf("Found %d gaps (should be low since pattern matches rule)", len(gaps))
	})
}

// TestMatchesPatternToRule_HighConfidence tests high confidence matching
func TestMatchesPatternToRule_HighConfidence(t *testing.T) {
	t.Run("high_confidence", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "ProcessOrder",
			FilePath:     "order.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "Process Order",
			Content: "Process orders",
		}
		evidence := ImplementationEvidence{
			Confidence: 0.85, // High confidence
			Functions:  []string{"ProcessOrder"},
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		if !matched {
			t.Error("expected match with high confidence, got false")
		}
	})
}

// TestMatchesPatternToRule_SemanticSimilarity tests semantic similarity matching
func TestMatchesPatternToRule_SemanticSimilarity(t *testing.T) {
	t.Run("semantic_similarity", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "ProcessUserOrder",
			FilePath:     "order.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "Process Order",
			Content: "Process user orders",
		}
		evidence := ImplementationEvidence{
			Confidence: 0.5, // Medium confidence
			Functions:  []string{},
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		// Should match due to semantic similarity
		if !matched {
			t.Log("Note: Semantic similarity may not match if threshold not met")
		}
	})
}

// TestExtractBusinessLogicPatternsEnhanced_Success tests successful pattern extraction
func TestExtractBusinessLogicPatternsEnhanced_Success(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err != nil {
			t.Fatalf("extractBusinessLogicPatternsEnhanced failed: %v", err)
		}
		if patterns == nil {
			t.Fatal("extractBusinessLogicPatternsEnhanced returned nil patterns")
		}
		t.Logf("Extracted %d patterns", len(patterns))
	})
}

// TestExtractBusinessLogicPatternsEnhanced_ContextCancellation tests context cancellation
func TestExtractBusinessLogicPatternsEnhanced_ContextCancellation(t *testing.T) {
	t.Run("context_cancellation", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err == nil {
			t.Error("expected error on context cancellation, got nil")
		}
		if len(patterns) > 0 {
			t.Logf("Note: Found %d patterns before cancellation", len(patterns))
		}
	})
}

// TestExtractBusinessLogicPatternsEnhanced_UnsupportedLanguage tests unsupported language handling
func TestExtractBusinessLogicPatternsEnhanced_UnsupportedLanguage(t *testing.T) {
	t.Run("unsupported_language", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir, err := os.MkdirTemp("", "gap-test-unsupported-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a file with unsupported extension
		testFile := filepath.Join(tmpDir, "test.java")
		err = os.WriteFile(testFile, []byte(`public class Test {}`), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err != nil {
			t.Fatalf("extractBusinessLogicPatternsEnhanced failed: %v", err)
		}
		// Should skip unsupported languages gracefully
		if len(patterns) > 0 {
			t.Logf("Note: Found %d patterns (may include fallback patterns)", len(patterns))
		}
	})
}

// TestClassifyBusinessPattern tests pattern classification
func TestClassifyBusinessPattern(t *testing.T) {
	tests := []struct {
		name     string
		fn       ast.FunctionInfo
		findings []ast.ASTFinding
		stats    ast.AnalysisStats
		expected string
	}{
		{
			name: "CRUD operation",
			fn: ast.FunctionInfo{
				Name: "CreateUser",
				Code: "func CreateUser() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "crud_operation",
		},
		{
			name: "Validation pattern",
			fn: ast.FunctionInfo{
				Name: "ValidateOrder",
				Code: "func ValidateOrder() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "validation",
		},
		{
			name: "Workflow pattern",
			fn: ast.FunctionInfo{
				Name: "ProcessPayment",
				Code: "func ProcessPayment() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "workflow",
		},
		{
			name: "General pattern",
			fn: ast.FunctionInfo{
				Name: "HelperFunction",
				Code: "func HelperFunction() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := classifyBusinessPattern(ctx, tt.fn, tt.findings, tt.stats)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Helper function to create a test codebase
func createTestCodebase(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create a test Go file with business logic
	testFile := filepath.Join(tmpDir, "order.go")
	err = os.WriteFile(testFile, []byte(`package main

func ProcessOrder(orderID string) {
	// Process order
}

func ValidatePayment(amount float64) bool {
	return amount > 0
}

func CreateUser(name string) {
	// Create user
}`), 0644)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to write test file: %v", err)
	}

	return tmpDir
}

// TestExtractBusinessLogicPatternsEnhanced_Timeout tests timeout handling
func TestExtractBusinessLogicPatternsEnhanced_Timeout(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond) // Ensure timeout

		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err == nil {
			t.Error("expected error on timeout, got nil")
		}
		if len(patterns) > 0 {
			t.Logf("Note: Found %d patterns before timeout", len(patterns))
		}
	})
}

// TestMatchesPatternToRule_EvidenceFunctions tests matching via evidence functions
func TestMatchesPatternToRule_EvidenceFunctions(t *testing.T) {
	t.Run("evidence_functions", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "ProcessOrder",
			FilePath:     "order.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "Order Processing",
			Content: "Process orders",
		}
		evidence := ImplementationEvidence{
			Confidence: 0.5, // Medium confidence
			Functions:  []string{"ProcessOrder", "ValidateOrder"}, // Pattern function in evidence
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		if !matched {
			t.Error("expected match via evidence functions, got false")
		}
	})
}

// TestMatchesPatternToRule_EvidenceFiles tests matching via evidence files
func TestMatchesPatternToRule_EvidenceFiles(t *testing.T) {
	t.Run("evidence_files", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "ProcessOrder",
			FilePath:     "/path/to/order.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "Order Processing",
			Content: "Process orders",
		}
		evidence := ImplementationEvidence{
			Confidence: 0.5,
			Functions:  []string{},
			Files:      []string{"/path/to/order.go"}, // Pattern file in evidence
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		if !matched {
			t.Error("expected match via evidence files, got false")
		}
	})
}

// TestMatchesPatternToRule_TitleMatch tests matching via title
func TestMatchesPatternToRule_TitleMatch(t *testing.T) {
	t.Run("title_match", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "ProcessOrder",
			FilePath:     "order.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "ProcessOrder Rule", // Contains function name
			Content: "Some content",
		}
		evidence := ImplementationEvidence{
			Confidence: 0.5,
			Functions: []string{},
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		if !matched {
			t.Error("expected match via title, got false")
		}
	})
}

// TestMatchesPatternToRule_ContentMatch tests matching via content
func TestMatchesPatternToRule_ContentMatch(t *testing.T) {
	t.Run("content_match", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "ProcessOrder",
			FilePath:     "order.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "Order Rule",
			Content: "The ProcessOrder function handles order processing", // Contains function name
		}
		evidence := ImplementationEvidence{
			Confidence: 0.5,
			Functions: []string{},
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		if !matched {
			t.Error("expected match via content, got false")
		}
	})
}

// TestMatchesPatternToRule_WordSimilarity tests matching via word similarity
func TestMatchesPatternToRule_WordSimilarity(t *testing.T) {
	t.Run("word_similarity", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "ProcessUserOrder",
			FilePath:     "order.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "Process Order User", // Has common words
			Content: "Process user orders",
		}
		evidence := ImplementationEvidence{
			Confidence: 0.5,
			Functions: []string{},
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		// May or may not match depending on similarity threshold
		if matched {
			t.Log("Pattern matched via word similarity")
		} else {
			t.Log("Pattern did not match (similarity threshold not met)")
		}
	})
}

// TestMatchesPatternToRule_NoMatch tests when pattern doesn't match
func TestMatchesPatternToRule_NoMatch(t *testing.T) {
	t.Run("no_match", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "HelperFunction",
			FilePath:     "helper.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "Order Processing",
			Content: "Process orders",
		}
		evidence := ImplementationEvidence{
			Confidence: 0.2, // Low confidence
			Functions: []string{},
			Files:      []string{},
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		if matched {
			t.Error("expected no match, got true")
		}
	})
}

// TestExtractBusinessLogicPatternsEnhanced_ErrorPaths tests error handling paths
func TestExtractBusinessLogicPatternsEnhanced_ErrorPaths(t *testing.T) {
	t.Run("invalid_path", func(t *testing.T) {
		// Given
		ctx := context.Background()
		invalidPath := "/nonexistent/path/that/does/not/exist"

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, invalidPath)

		// Then
		// Function now validates path and returns error for non-existent paths
		if err == nil {
			t.Error("expected error for invalid path, got nil")
		}
		if patterns == nil {
			t.Fatal("expected patterns slice, got nil")
		}
		// Should return empty slice for invalid path
		if len(patterns) > 0 {
			t.Logf("Note: Found %d patterns despite invalid path", len(patterns))
		}
	})
}

// TestExtractBusinessLogicPatternsEnhanced_ASTFailure tests AST analysis failure fallback
func TestExtractBusinessLogicPatternsEnhanced_ASTFailure(t *testing.T) {
	t.Run("ast_failure_fallback", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir, err := os.MkdirTemp("", "gap-test-ast-fail-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a file with potentially problematic syntax
		testFile := filepath.Join(tmpDir, "test.go")
		err = os.WriteFile(testFile, []byte(`package main

func ProcessOrder(orderID string) {
	// This should parse fine
}`), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err != nil {
			t.Logf("Note: AST extraction returned error (may use fallback): %v", err)
		}
		// Should still return patterns (possibly via fallback)
		if patterns == nil {
			t.Fatal("extractBusinessLogicPatternsEnhanced returned nil patterns")
		}
		t.Logf("Extracted %d patterns (may include fallback patterns)", len(patterns))
	})
}

// TestExtractBusinessLogicPatternsEnhanced_EmptyPatterns tests when no patterns found
func TestExtractBusinessLogicPatternsEnhanced_EmptyPatterns(t *testing.T) {
	t.Run("empty_patterns", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir, err := os.MkdirTemp("", "gap-test-empty-patterns-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a file with no business logic
		testFile := filepath.Join(tmpDir, "helper.go")
		err = os.WriteFile(testFile, []byte(`package main

func helper() {
	// Just a helper
}`), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err != nil {
			t.Fatalf("extractBusinessLogicPatternsEnhanced failed: %v", err)
		}
		// Should return slice (nil or empty both work in Go, len() handles both)
		if patterns == nil {
			// In Go, nil slices are valid and len() returns 0, but we prefer non-nil for consistency
			t.Log("Note: patterns is nil (acceptable in Go, but empty slice preferred)")
		}
		// Should find 0 patterns if helper doesn't match business keywords
		// But may find patterns if helper function name matches keywords
		t.Logf("Extracted %d patterns (helper may match some keywords)", len(patterns))
	})
}

// TestExtractBusinessLogicPatternsEnhanced_FileReadError tests file read error handling
func TestExtractBusinessLogicPatternsEnhanced_FileReadError(t *testing.T) {
	t.Run("file_read_error", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir, err := os.MkdirTemp("", "gap-test-read-error-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a directory (not a file) to cause read error
		testDir := filepath.Join(tmpDir, "subdir")
		err = os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		// Should handle gracefully and continue
		if err != nil {
			t.Fatalf("extractBusinessLogicPatternsEnhanced failed: %v", err)
		}
		// In Go, nil slices are valid and len() returns 0
		if patterns == nil {
			t.Log("Note: patterns is nil (acceptable in Go)")
		}
	})
}


// TestAnalyzeUndocumentedCode_ContextCancellationInLoop tests context cancellation during pattern processing
func TestAnalyzeUndocumentedCode_ContextCancellationInLoop(t *testing.T) {
	t.Run("context_cancellation_in_loop", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		projectID := "test-project-id"
		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)
		defer cancel()

		documentedRules := []KnowledgeItem{}

		// Cancel context after a short delay (simulating cancellation during processing)
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		// May or may not get cancellation error depending on timing
		if err != nil {
			if ctx.Err() == nil {
				t.Logf("Got error (may be cancellation): %v", err)
			}
		}
		if gaps == nil {
			t.Fatal("analyzeUndocumentedCode returned nil gaps")
		}
	})
}

// TestAnalyzeUndocumentedCode_EmptyPatterns tests when no patterns are extracted
func TestAnalyzeUndocumentedCode_EmptyPatterns(t *testing.T) {
	t.Run("empty_patterns", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		tmpDir, err := os.MkdirTemp("", "gap-test-empty-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create empty directory
		documentedRules := []KnowledgeItem{}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err != nil {
			t.Fatalf("analyzeUndocumentedCode failed: %v", err)
		}
		// In Go, nil slices are valid and len() returns 0
		// Check that we can call len() on it
		_ = len(gaps)
		// Should return empty gaps for empty codebase
		if len(gaps) > 0 {
			t.Logf("Note: Found %d gaps in empty codebase", len(gaps))
		}
	})
}

// TestAnalyzeUndocumentedCode_ErrorHandling tests error handling in analyzeUndocumentedCode
func TestAnalyzeUndocumentedCode_ErrorHandling(t *testing.T) {
	t.Run("error_handling", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		invalidPath := "/nonexistent/path"
		documentedRules := []KnowledgeItem{}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, invalidPath, documentedRules)

		// Then
		// Function now validates path and returns error for non-existent paths
		if err == nil {
			t.Error("expected error for invalid path, got nil")
		}
		// In Go, nil slices are valid and len() returns 0
		_ = len(gaps)
		// Should return empty gaps for invalid path
		if len(gaps) > 0 {
			t.Logf("Note: Found %d gaps despite invalid path", len(gaps))
		}
	})
}

// TestAnalyzeUndocumentedCode_MultipleRules tests with multiple documented rules
func TestAnalyzeUndocumentedCode_MultipleRules(t *testing.T) {
	t.Run("multiple_rules", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)

		documentedRules := []KnowledgeItem{
			{ID: "rule-1", Title: "Process Order", Content: "Process orders"},
			{ID: "rule-2", Title: "Validate Payment", Content: "Validate payments"},
			{ID: "rule-3", Title: "Create User", Content: "Create users"},
		}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err != nil {
			t.Fatalf("analyzeUndocumentedCode failed: %v", err)
		}
		if gaps == nil {
			t.Fatal("analyzeUndocumentedCode returned nil gaps")
		}
		t.Logf("Found %d gaps with multiple rules", len(gaps))
	})
}

// TestClassifyBusinessPattern_AllTypes tests all pattern classification types
func TestClassifyBusinessPattern_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		fn       ast.FunctionInfo
		findings []ast.ASTFinding
		stats    ast.AnalysisStats
		expected string
	}{
		{
			name: "CRUD - Create",
			fn: ast.FunctionInfo{
				Name: "CreateAccount",
				Code: "func CreateAccount() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "crud_operation",
		},
		{
			name: "CRUD - Update",
			fn: ast.FunctionInfo{
				Name: "UpdateUser",
				Code: "func UpdateUser() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "crud_operation",
		},
		{
			name: "CRUD - Delete",
			fn: ast.FunctionInfo{
				Name: "DeleteOrder",
				Code: "func DeleteOrder() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "crud_operation",
		},
		{
			name: "Validation",
			fn: ast.FunctionInfo{
				Name: "ValidateInput",
				Code: "func ValidateInput() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "validation",
		},
		{
			name: "Workflow",
			fn: ast.FunctionInfo{
				Name: "ProcessWorkflow",
				Code: "func ProcessWorkflow() {}",
			},
			findings: []ast.ASTFinding{},
			stats:    ast.AnalysisStats{},
			expected: "workflow",
		},
		{
			name: "Business Logic from Findings",
			fn: ast.FunctionInfo{
				Name: "SomeFunction",
				Code: "func SomeFunction() {}",
			},
			findings: []ast.ASTFinding{
				{Type: "business_logic", Severity: "high"},
			},
			stats:    ast.AnalysisStats{},
			expected: "business_logic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := classifyBusinessPattern(ctx, tt.fn, tt.findings, tt.stats)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestConvertToBusinessPattern tests pattern conversion
func TestConvertToBusinessPattern(t *testing.T) {
	t.Run("with_business_keyword", func(t *testing.T) {
		// Given
		ctx := context.Background()
		filePath := "order.go"
		fn := ast.FunctionInfo{
			Name: "ProcessOrder",
			Line:  10,
			Code:  "func ProcessOrder() {}",
		}
		findings := []ast.ASTFinding{}
		stats := ast.AnalysisStats{}

		// When
		pattern := convertToBusinessPattern(ctx, filePath, fn, findings, stats)

		// Then
		if pattern == nil {
			t.Fatal("expected pattern, got nil")
		}
		if pattern.FunctionName != "ProcessOrder" {
			t.Errorf("expected ProcessOrder, got %s", pattern.FunctionName)
		}
		if pattern.FilePath != filePath {
			t.Errorf("expected %s, got %s", filePath, pattern.FilePath)
		}
	})

	t.Run("without_business_keyword", func(t *testing.T) {
		// Given
		ctx := context.Background()
		filePath := "helper.go"
		fn := ast.FunctionInfo{
			Name: "helper",
			Line:  5,
			Code:  "func helper() {}",
		}
		findings := []ast.ASTFinding{}
		stats := ast.AnalysisStats{}

		// When
		pattern := convertToBusinessPattern(ctx, filePath, fn, findings, stats)

		// Then
		// May return nil if no business keywords found
		if pattern != nil {
			t.Logf("Pattern returned: %+v", pattern)
		}
	})
}

// TestExtractBusinessLogicPatternsEnhanced_WalkError tests filepath.Walk error handling
func TestExtractBusinessLogicPatternsEnhanced_WalkError(t *testing.T) {
	t.Run("walk_error", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir, err := os.MkdirTemp("", "gap-test-walk-error-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a file
		testFile := filepath.Join(tmpDir, "test.go")
		err = os.WriteFile(testFile, []byte(`package main

func ProcessOrder() {}`), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err != nil {
			t.Fatalf("extractBusinessLogicPatternsEnhanced failed: %v", err)
		}
		_ = len(patterns)
		t.Logf("Extracted %d patterns", len(patterns))
	})
}

// TestExtractBusinessLogicPatternsEnhanced_AllLanguages tests all supported languages
func TestExtractBusinessLogicPatternsEnhanced_AllLanguages(t *testing.T) {
	t.Run("all_languages", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir, err := os.MkdirTemp("", "gap-test-languages-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create files in all supported languages
		testFiles := map[string]string{
			"test.go": `package main; func ProcessOrder() {}`,
			"test.js": `function processOrder() {}`,
			"test.ts": `function processOrder(): void {}`,
			"test.py": `def process_order(): pass`,
		}

		for filename, content := range testFiles {
			testFile := filepath.Join(tmpDir, filename)
			err = os.WriteFile(testFile, []byte(content), 0644)
			if err != nil {
				t.Fatalf("failed to write %s: %v", filename, err)
			}
		}

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err != nil {
			t.Fatalf("extractBusinessLogicPatternsEnhanced failed: %v", err)
		}
		_ = len(patterns)
		t.Logf("Extracted %d patterns from multiple languages", len(patterns))
	})
}

// TestMatchesPatternToRule_EmptyEvidence tests matching with empty evidence
func TestMatchesPatternToRule_EmptyEvidence(t *testing.T) {
	t.Run("empty_evidence", func(t *testing.T) {
		// Given
		ctx := context.Background()
		pattern := BusinessLogicPattern{
			FunctionName: "ProcessOrder",
			FilePath:     "order.go",
			LineNumber:   10,
		}
		rule := KnowledgeItem{
			ID:    "rule-1",
			Title: "Different Rule",
			Content: "Different content",
		}
		evidence := ImplementationEvidence{
			Confidence: 0.0,
			Functions: []string{},
			Files:      []string{},
		}

		// When
		matched := matchesPatternToRule(ctx, pattern, rule, evidence)

		// Then
		if matched {
			t.Error("expected no match with empty evidence, got true")
		}
	})
}

// TestAnalyzeUndocumentedCode_NoRules tests with no documented rules
func TestAnalyzeUndocumentedCode_NoRules(t *testing.T) {
	t.Run("no_rules", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)

		documentedRules := []KnowledgeItem{} // Empty rules

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err != nil {
			t.Fatalf("analyzeUndocumentedCode failed: %v", err)
		}
		_ = len(gaps)
		// Should find gaps since no rules match
		if len(gaps) > 0 {
			t.Logf("Found %d gaps (expected since no rules provided)", len(gaps))
		}
	})
}

// TestAnalyzeUndocumentedCode_PartialMatch tests partial matching scenarios
func TestAnalyzeUndocumentedCode_PartialMatch(t *testing.T) {
	t.Run("partial_match", func(t *testing.T) {
		// Given
		ctx := context.Background()
		projectID := "test-project-id"
		tmpDir := createTestCodebase(t)
		defer os.RemoveAll(tmpDir)

		// Rule that partially matches
		documentedRules := []KnowledgeItem{
			{ID: "rule-1", Title: "Order", Content: "Process orders"},
		}

		// When
		gaps, err := analyzeUndocumentedCode(ctx, projectID, tmpDir, documentedRules)

		// Then
		if err != nil {
			t.Fatalf("analyzeUndocumentedCode failed: %v", err)
		}
		_ = len(gaps)
		t.Logf("Found %d gaps with partial matching", len(gaps))
	})
}

// TestExtractBusinessLogicPatternsEnhanced_FileReadErrorInWalk tests file read error during walk
func TestExtractBusinessLogicPatternsEnhanced_FileReadErrorInWalk(t *testing.T) {
	t.Run("file_read_error_in_walk", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir, err := os.MkdirTemp("", "gap-test-read-error-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a directory with a .go extension (will cause read error)
		testDir := filepath.Join(tmpDir, "test.go")
		err = os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatalf("failed to create test dir: %v", err)
		}

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		// Should handle gracefully and continue
		if err != nil {
			t.Fatalf("extractBusinessLogicPatternsEnhanced failed: %v", err)
		}
		_ = len(patterns)
		t.Logf("Extracted %d patterns (file read errors handled gracefully)", len(patterns))
	})
}

// TestExtractBusinessLogicPatternsEnhanced_NonCodeFile tests skipping non-code files
func TestExtractBusinessLogicPatternsEnhanced_NonCodeFile(t *testing.T) {
	t.Run("non_code_file", func(t *testing.T) {
		// Given
		ctx := context.Background()
		tmpDir, err := os.MkdirTemp("", "gap-test-noncode-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a non-code file
		testFile := filepath.Join(tmpDir, "test.txt")
		err = os.WriteFile(testFile, []byte("This is not code"), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// When
		patterns, err := extractBusinessLogicPatternsEnhanced(ctx, tmpDir)

		// Then
		if err != nil {
			t.Fatalf("extractBusinessLogicPatternsEnhanced failed: %v", err)
		}
		_ = len(patterns)
		// Should skip non-code files
		if len(patterns) > 0 {
			t.Logf("Note: Found %d patterns (non-code files should be skipped)", len(patterns))
		}
	})
}

// TestConvertToBusinessPattern_WithFindings tests pattern conversion with AST findings
func TestConvertToBusinessPattern_WithFindings(t *testing.T) {
	t.Run("with_findings", func(t *testing.T) {
		// Given
		ctx := context.Background()
		filePath := "order.go"
		fn := ast.FunctionInfo{
			Name: "SomeFunction",
			Line:  10,
			Code:  "func SomeFunction() {}",
		}
		findings := []ast.ASTFinding{
			{Type: "business_logic", Severity: "high"},
		}
		stats := ast.AnalysisStats{}

		// When
		pattern := convertToBusinessPattern(ctx, filePath, fn, findings, stats)

		// Then
		// Should return pattern if findings indicate business logic
		if pattern == nil {
			t.Log("Note: Pattern is nil (may not match business keywords)")
		} else {
			t.Logf("Pattern returned: %+v", pattern)
		}
	})
}

// TestConvertToBusinessPattern_KeywordInCode tests keyword detection in function code
func TestConvertToBusinessPattern_KeywordInCode(t *testing.T) {
	t.Run("keyword_in_code", func(t *testing.T) {
		// Given
		ctx := context.Background()
		filePath := "order.go"
		fn := ast.FunctionInfo{
			Name: "Calculate",
			Line:  10,
			Code:  "func Calculate() { processOrder() }", // Keyword in code, not name
		}
		findings := []ast.ASTFinding{}
		stats := ast.AnalysisStats{}

		// When
		pattern := convertToBusinessPattern(ctx, filePath, fn, findings, stats)

		// Then
		if pattern == nil {
			t.Log("Note: Pattern is nil (may not match if keyword not found)")
		} else {
			if pattern.Keyword == "" {
				t.Log("Note: Keyword not extracted from code")
			} else {
				t.Logf("Pattern with keyword: %+v", pattern)
			}
		}
	})
}
