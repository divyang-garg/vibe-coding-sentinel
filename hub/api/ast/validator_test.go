// Package ast provides tests for codebase validation
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateOrphanedFunction_Found tests validation when function is found in codebase
func TestValidateOrphanedFunction_Found(t *testing.T) {
	// Create a temporary test directory structure
	tmpDir := t.TempDir()
	projectRoot := tmpDir

	// Create a file with a function
	file1 := filepath.Join(tmpDir, "file1.go")
	err := os.WriteFile(file1, []byte(`
package main

func helperFunction() {
	// This function is used
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create another file that calls the function
	file2 := filepath.Join(tmpDir, "file2.go")
	err = os.WriteFile(file2, []byte(`
package main

func main() {
	helperFunction()
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create finding for helperFunction
	finding := &ASTFinding{
		Type:      "orphaned_code",
		Message:   "Potentially orphaned function: 'helperFunction' is defined but never called",
		Code:      "func helperFunction() {\n\t// This function is used\n}",
		Line:      4,
		Validated: false,
	}

	// Validate
	err = ValidateFinding(finding, "file1.go", projectRoot, "go")
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Should have found references, so confidence should be 0
	if finding.Confidence > 0.1 {
		t.Errorf("Expected confidence ~0, got %.2f", finding.Confidence)
	}
	if finding.AutoFixSafe {
		t.Error("Expected AutoFixSafe=false when function is found in codebase")
	}
	if !finding.Validated {
		t.Error("Expected Validated=true after validation")
	}
}

// TestValidateOrphanedFunction_TrulyOrphaned tests validation when function is truly orphaned
func TestValidateOrphanedFunction_TrulyOrphaned(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := tmpDir

	// Create a file with an orphaned function
	file1 := filepath.Join(tmpDir, "file1.go")
	err := os.WriteFile(file1, []byte(`
package main

func trulyOrphaned() {
	// This function is never called
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	finding := &ASTFinding{
		Type:      "orphaned_code",
		Message:   "Potentially orphaned function: 'trulyOrphaned' is defined but never called",
		Code:      "func trulyOrphaned() {\n\t// This function is never called\n}",
		Line:      4,
		Validated: false,
	}

	err = ValidateFinding(finding, "file1.go", projectRoot, "go")
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Should have high confidence (95%) for truly orphaned function
	if finding.Confidence < 0.90 {
		t.Errorf("Expected confidence >= 0.90, got %.2f", finding.Confidence)
	}
	if !finding.AutoFixSafe {
		t.Error("Expected AutoFixSafe=true for truly orphaned function")
	}
	if finding.FixType != "delete" {
		t.Errorf("Expected FixType=delete, got %s", finding.FixType)
	}
}

// TestValidateUnusedVariable_LocalScope tests validation for unused local variable
func TestValidateUnusedVariable_LocalScope(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := tmpDir

	file1 := filepath.Join(tmpDir, "file1.go")
	err := os.WriteFile(file1, []byte(`
package main

func test() {
	var unusedVar int
	// unusedVar is never used
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	finding := &ASTFinding{
		Type:      "unused_variable",
		Message:   "Unused variable: 'unusedVar' is declared but never used",
		Code:      "var unusedVar int",
		Line:      5,
		Validated: false,
	}

	err = ValidateFinding(finding, "file1.go", projectRoot, "go")
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Should have high confidence for local unused variable
	if finding.Confidence < 0.85 {
		t.Errorf("Expected confidence >= 0.85, got %.2f", finding.Confidence)
	}
	if !finding.AutoFixSafe {
		t.Error("Expected AutoFixSafe=true for local unused variable")
	}
}

// TestValidateUnusedVariable_Exported tests that exported variables are never auto-fixed
func TestValidateUnusedVariable_Exported(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := tmpDir

	file1 := filepath.Join(tmpDir, "file1.go")
	err := os.WriteFile(file1, []byte(`
package main

var ExportedVar int
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	finding := &ASTFinding{
		Type:      "unused_variable",
		Message:   "Unused variable: 'ExportedVar' is declared but never used",
		Code:      "var ExportedVar int",
		Line:      4,
		Validated: false,
	}

	err = ValidateFinding(finding, "file1.go", projectRoot, "go")
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Exported variables should never be auto-fixed
	if finding.Confidence > 0.1 {
		t.Errorf("Expected confidence ~0 for exported variable, got %.2f", finding.Confidence)
	}
	if finding.AutoFixSafe {
		t.Error("Expected AutoFixSafe=false for exported variable")
	}
}

// TestValidateEmptyCatch_WithTODO tests validation when empty catch has intent comment
func TestValidateEmptyCatch_WithTODO(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := tmpDir

	file1 := filepath.Join(tmpDir, "file1.js")
	err := os.WriteFile(file1, []byte(`
try {
	doSomething();
} catch (e) {
	// TODO: Add error handling
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	finding := &ASTFinding{
		Type:      "empty_catch",
		Message:   "Empty catch/except block detected - errors are silently ignored",
		Code:      "} catch (e) {\n\t// TODO: Add error handling\n}",
		Line:      4,
		Validated: false,
	}

	err = ValidateFinding(finding, "file1.js", projectRoot, "javascript")
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Should have low confidence when intent comment found
	if finding.Confidence > 0.1 {
		t.Errorf("Expected confidence ~0 when intent comment found, got %.2f", finding.Confidence)
	}
	if finding.AutoFixSafe {
		t.Error("Expected AutoFixSafe=false when intent comment found")
	}
}

// TestValidateEmptyCatch_NoIntent tests validation when empty catch has no intent
func TestValidateEmptyCatch_NoIntent(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := tmpDir

	file1 := filepath.Join(tmpDir, "file1.js")
	err := os.WriteFile(file1, []byte(`
try {
	doSomething();
} catch (e) {
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	finding := &ASTFinding{
		Type:      "empty_catch",
		Message:   "Empty catch/except block detected - errors are silently ignored",
		Code:      "} catch (e) {\n}",
		Line:      4,
		Validated: false,
	}

	err = ValidateFinding(finding, "file1.js", projectRoot, "javascript")
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Should have high confidence when no intent comment
	if finding.Confidence < 0.80 {
		t.Errorf("Expected confidence >= 0.80, got %.2f", finding.Confidence)
	}
	if finding.FixType != "refactor" {
		t.Errorf("Expected FixType=refactor, got %s", finding.FixType)
	}
}

// TestConfidenceScoring tests the confidence scoring matrix
func TestConfidenceScoring(t *testing.T) {
	tests := []struct {
		name             string
		findingType      string
		validationResult ValidationResult
		expectedMin      float64
		expectedMax      float64
	}{
		{
			name:        "orphaned_code_found",
			findingType: "orphaned_code",
			validationResult: ValidationResult{
				FoundInCodebase: true,
				ReferenceCount:  5,
			},
			expectedMin: 0.0,
			expectedMax: 0.1,
		},
		{
			name:        "orphaned_code_not_found",
			findingType: "orphaned_code",
			validationResult: ValidationResult{
				FoundInCodebase: false,
				IsExported:      false,
			},
			expectedMin: 0.90,
			expectedMax: 1.0,
		},
		{
			name:        "unused_variable_local",
			findingType: "unused_variable",
			validationResult: ValidationResult{
				FoundInCodebase: false,
				IsExported:      false,
			},
			expectedMin: 0.85,
			expectedMax: 1.0,
		},
		{
			name:        "empty_catch_no_intent",
			findingType: "empty_catch",
			validationResult: ValidationResult{
				HasIntent: false,
			},
			expectedMin: 0.80,
			expectedMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finding := &ASTFinding{Type: tt.findingType}
			confidence := CalculateConfidence(finding, tt.validationResult)
			if confidence < tt.expectedMin || confidence > tt.expectedMax {
				t.Errorf("Expected confidence between %.2f and %.2f, got %.2f",
					tt.expectedMin, tt.expectedMax, confidence)
			}
		})
	}
}

// TestDetermineAutoFixSafe tests the auto-fix safety determination
func TestDetermineAutoFixSafe(t *testing.T) {
	tests := []struct {
		name         string
		confidence   float64
		findingType  string
		expectedSafe bool
	}{
		{
			name:         "high_confidence_orphaned",
			confidence:   0.95,
			findingType:  "orphaned_code",
			expectedSafe: true,
		},
		{
			name:         "low_confidence",
			confidence:   0.50,
			findingType:  "orphaned_code",
			expectedSafe: false,
		},
		{
			name:         "duplicate_function_never_safe",
			confidence:   0.99,
			findingType:  "duplicate_function",
			expectedSafe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe := DetermineAutoFixSafe(tt.confidence, tt.findingType)
			if safe != tt.expectedSafe {
				t.Errorf("Expected AutoFixSafe=%v, got %v", tt.expectedSafe, safe)
			}
		})
	}
}

// TestSearchCodebase tests the codebase search functionality
func TestSearchCodebase(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := tmpDir

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.go")
	err := os.WriteFile(file1, []byte(`
package main

func testFunction() {
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file2 := filepath.Join(tmpDir, "file2.go")
	err = os.WriteFile(file2, []byte(`
package main

func main() {
	testFunction()
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Search for function calls
	results, err := SearchCodebase(`\btestFunction\s*\(`, projectRoot, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should find at least one reference
	if len(results) == 0 {
		t.Error("Expected to find at least one reference to testFunction")
	}
}
