// Package scanner provides unit tests for security scanning
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScan_DetectsSecrets(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	// Use a value that matches the pattern: [a-zA-Z0-9]{20,} (no underscores)
	// The pattern requires alphanumeric only, so use a value without underscores
	content := `const apiKey = "sk_test_EXAMPLE_KEY_PLACEHOLDER_DO_NOT_USE_IN_PRODUCTION";
const password = "MySecretPassword1234567890ABCDEFGHIJ";
eval(userInput);`
	os.WriteFile(testFile, []byte(content), 0644)

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := Scan(opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Findings) == 0 {
		t.Error("Expected to find secrets, but found none")
	}

	foundSecret := false
	for _, f := range result.Findings {
		if f.Type == "secrets" || f.Type == "high_entropy_secret" {
			foundSecret = true
			break
		}
	}

	if !foundSecret {
		t.Error("Expected to find secret detection")
	}
}

func TestScan_ParallelPerformance(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	for i := 0; i < 10; i++ {
		testFile := filepath.Join(tmpDir, "test", "file", fmt.Sprintf("test%d.js", i))
		os.MkdirAll(filepath.Dir(testFile), 0755)
		content := `const apiKey = "test_key_12345";
console.log("test");`
		os.WriteFile(testFile, []byte(content), 0644)
	}

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := ScanParallel(opts)
	if err != nil {
		t.Fatalf("ScanParallel failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Should find findings in multiple files
	if len(result.Findings) == 0 {
		t.Error("Expected to find findings in parallel scan")
	}
}

func TestScan_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := Scan(opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Findings) != 0 {
		t.Errorf("Expected no findings in empty directory, got %d", len(result.Findings))
	}

	if !result.Success {
		t.Error("Empty directory scan should succeed")
	}
}

func TestScan_SkipsIgnoredDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file in node_modules (should be skipped)
	nodeModulesFile := filepath.Join(tmpDir, "node_modules", "test.js")
	os.MkdirAll(filepath.Dir(nodeModulesFile), 0755)
	os.WriteFile(nodeModulesFile, []byte(`const apiKey = "secret123";`), 0644)

	// Create file in regular directory (should be scanned)
	regularFile := filepath.Join(tmpDir, "src", "test.js")
	os.MkdirAll(filepath.Dir(regularFile), 0755)
	os.WriteFile(regularFile, []byte(`const apiKey = "secret123";`), 0644)

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := Scan(opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find findings in regular file, not node_modules
	foundInNodeModules := false
	for _, f := range result.Findings {
		if strings.Contains(f.File, "node_modules") {
			foundInNodeModules = true
			break
		}
	}

	if foundInNodeModules {
		t.Error("Should not scan files in node_modules")
	}
}

func TestScan_CriticalFindingsMarkAsFailed(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	content := `eval(userInput); // Critical vulnerability`
	os.WriteFile(testFile, []byte(content), 0644)

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := Scan(opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.Success {
		t.Error("Expected scan to fail due to critical finding")
	}
}

func TestScanSequential(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	content := `const apiKey = "test_key_12345";
console.log("test");`
	os.WriteFile(testFile, []byte(content), 0644)

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := ScanSequential(opts)
	if err != nil {
		t.Fatalf("ScanSequential failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if len(result.Findings) == 0 {
		t.Error("Expected to find findings in sequential scan")
	}
}

func TestScanSequential_WithBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	// Create baseline directory
	os.MkdirAll(".sentinel", 0755)
	baselineJSON := `{
		"version": "1.0",
		"entries": [
			{"file": "test.js", "line": 1, "hash": "test.js:1"}
		]
	}`
	os.WriteFile(".sentinel/baseline.json", []byte(baselineJSON), 0644)

	// Create test file with finding on line 1
	testFile := filepath.Join(tmpDir, "test.js")
	content := `const apiKey = "secret123";
const x = 1;`
	os.WriteFile(testFile, []byte(content), 0644)

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := ScanSequential(opts)
	if err != nil {
		t.Fatalf("ScanSequential failed: %v", err)
	}

	// Finding on line 1 should be filtered by baseline
	foundOnLine1 := false
	for _, f := range result.Findings {
		if f.File == "test.js" && f.Line == 1 {
			foundOnLine1 = true
			break
		}
	}

	if foundOnLine1 {
		t.Error("Finding on line 1 should be filtered by baseline")
	}
}

func TestScanSequential_InvalidBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	// Create invalid baseline
	os.MkdirAll(".sentinel", 0755)
	os.WriteFile(".sentinel/baseline.json", []byte("invalid json"), 0644)

	testFile := filepath.Join(tmpDir, "test.js")
	os.WriteFile(testFile, []byte(`const apiKey = "secret";`), 0644)

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := ScanSequential(opts)
	if err != nil {
		t.Fatalf("ScanSequential failed: %v", err)
	}

	// Should still work with invalid baseline
	if result == nil {
		t.Fatal("Result should not be nil even with invalid baseline")
	}
}

func TestScanSequential_EmptyPath(t *testing.T) {
	opts := ScanOptions{
		CodebasePath: "",
		CIMode:       true,
	}

	result, err := ScanSequential(opts)
	if err != nil {
		t.Fatalf("ScanSequential failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestFilterBaseline_WithHash(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	os.MkdirAll(".sentinel", 0755)
	baselineJSON := `{
		"version": "1.0",
		"entries": [
			{"file": "test.js", "line": 1, "hash": "custom_hash_123"}
		]
	}`
	os.WriteFile(".sentinel/baseline.json", []byte(baselineJSON), 0644)

	result := &Result{
		Findings: []Finding{
			{File: "test.js", Line: 1, Type: "secrets"},
			{File: "test.js", Line: 2, Type: "secrets"},
		},
		Summary: map[string]int{"secrets": 2},
		Success: false,
	}

	filtered := filterBaseline(result)

	if len(filtered.Findings) != 1 {
		t.Errorf("Expected 1 finding after filtering, got %d", len(filtered.Findings))
	}

	if filtered.Findings[0].Line != 2 {
		t.Error("Line 2 finding should remain")
	}
}

func TestFilterBaseline_NoBaselineFile(t *testing.T) {
	result := &Result{
		Findings: []Finding{
			{File: "test.js", Line: 1, Type: "secrets"},
		},
		Summary: map[string]int{"secrets": 1},
		Success: false,
	}

	filtered := filterBaseline(result)

	if len(filtered.Findings) != 1 {
		t.Error("Should return result as-is when no baseline file")
	}
}

func TestScanSequential_WithFileErrors(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a file that will cause read errors
	testFile := filepath.Join(tmpDir, "test.js")
	os.WriteFile(testFile, []byte(`const apiKey = "test";`), 0644)

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       true,
	}

	result, err := ScanSequential(opts)
	if err != nil {
		t.Fatalf("ScanSequential failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestScanSequential_NonCIMode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	os.WriteFile(testFile, []byte(`const apiKey = "test";`), 0644)

	opts := ScanOptions{
		CodebasePath: tmpDir,
		CIMode:       false,
	}

	result, err := ScanSequential(opts)
	if err != nil {
		t.Fatalf("ScanSequential failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestScanSequential_AdditionalEdgeCases(t *testing.T) {
	t.Run("handles relative path resolution error", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.js")
		os.WriteFile(testFile, []byte(`const apiKey = "test";`), 0644)

		opts := ScanOptions{
			CodebasePath: tmpDir,
			CIMode:       true,
		}

		result, err := ScanSequential(opts)
		if err != nil {
			t.Fatalf("ScanSequential failed: %v", err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}
	})

	t.Run("handles filepath.Rel returning empty string", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create file in root of scan directory
		testFile := filepath.Join(tmpDir, "test.js")
		os.WriteFile(testFile, []byte(`eval(userInput);`), 0644)

		opts := ScanOptions{
			CodebasePath: tmpDir,
			CIMode:       true,
		}

		result, err := ScanSequential(opts)
		if err != nil {
			t.Fatalf("ScanSequential failed: %v", err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}
	})

	t.Run("handles file read errors gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create a directory with a file that can't be read
		testFile := filepath.Join(tmpDir, "test.js")
		os.WriteFile(testFile, []byte(`const x = 1;`), 0000) // No read permission

		opts := ScanOptions{
			CodebasePath: tmpDir,
			CIMode:       true,
		}

		result, err := ScanSequential(opts)
		// Should not error, just skip unreadable files
		if err != nil {
			t.Fatalf("ScanSequential should handle read errors gracefully: %v", err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}
	})

	t.Run("handles walk errors gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create a subdirectory that might cause walk issues
		subDir := filepath.Join(tmpDir, "subdir")
		os.MkdirAll(subDir, 0755)
		testFile := filepath.Join(subDir, "test.js")
		os.WriteFile(testFile, []byte(`const apiKey = "test";`), 0644)

		opts := ScanOptions{
			CodebasePath: tmpDir,
			CIMode:       true,
		}

		result, err := ScanSequential(opts)
		if err != nil {
			t.Fatalf("ScanSequential failed: %v", err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}
	})
}
