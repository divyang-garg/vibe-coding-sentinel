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
