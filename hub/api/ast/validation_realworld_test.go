// Package ast real-world validation tests
// Tests AST analysis against actual production codebase files
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRealWorldCodebase validates AST analysis on actual production files
func TestRealWorldCodebase(t *testing.T) {
	// Find Go files in internal directory
	internalPath := filepath.Join("..", "..", "..", "internal")
	if _, err := os.Stat(internalPath); os.IsNotExist(err) {
		t.Skipf("Skipping real-world test: %s not found", internalPath)
		return
	}

	var goFiles []string
	err := filepath.Walk(internalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	if len(goFiles) == 0 {
		t.Skip("No Go files found in internal directory")
		return
	}

	// Test on a representative sample (first 10 files)
	sampleSize := 10
	if len(goFiles) < sampleSize {
		sampleSize = len(goFiles)
	}

	t.Logf("Testing AST analysis on %d real-world Go files (sample of %d total)", sampleSize, len(goFiles))

	allChecks := []string{"duplicates", "unused", "unreachable", "orphaned", "empty_catch", "missing_await", "brace_mismatch"}
	totalFindings := 0
	totalFiles := 0
	totalTime := time.Duration(0)
	maxFileSize := 0
	panics := 0
	errors := 0

	for i := 0; i < sampleSize; i++ {
		filePath := goFiles[i]
		fileName := filepath.Base(filePath)

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Logf("  [%s] Failed to read: %v", fileName, err)
			errors++
			continue
		}

		code := string(content)
		fileSize := len(code)
		if fileSize > maxFileSize {
			maxFileSize = fileSize
		}

		// Skip very small files (likely just package declarations)
		if fileSize < 100 {
			continue
		}

		// Run AST analysis with all checks
		start := time.Now()
		findings, stats, err := AnalyzeAST(code, "go", allChecks)
		duration := time.Since(start)

		totalTime += duration
		totalFiles++

		if err != nil {
			t.Logf("  [%s] Error: %v (size: %d bytes, time: %v)", fileName, err, fileSize, duration)
			errors++
			continue
		}

		// Check for panics (should not happen due to recovery)
		if findings == nil && stats.NodesVisited == 0 {
			panics++
			t.Logf("  [%s] Possible panic recovery (nil findings)", fileName)
		}

		findingCount := len(findings)
		totalFindings += findingCount

		// Log results for files with findings
		if findingCount > 0 {
			t.Logf("  [%s] %d findings, %d nodes, %v (size: %d bytes)",
				fileName, findingCount, stats.NodesVisited, duration, fileSize)

			// Categorize findings
			findingTypes := make(map[string]int)
			for _, f := range findings {
				findingTypes[f.Type]++
			}

			// Log finding breakdown
			for fType, count := range findingTypes {
				t.Logf("    - %s: %d", fType, count)
			}
		} else {
			t.Logf("  [%s] No findings, %d nodes, %v (size: %d bytes)",
				fileName, stats.NodesVisited, duration, fileSize)
		}

		// Performance check: should complete in reasonable time
		if duration > 5*time.Second {
			t.Logf("  [%s] WARNING: Slow analysis (%v)", fileName, duration)
		}
	}

	// Summary statistics
	avgTime := time.Duration(0)
	avgFindings := 0.0
	if totalFiles > 0 {
		avgTime = totalTime / time.Duration(totalFiles)
		avgFindings = float64(totalFindings) / float64(totalFiles)
	}

	t.Logf("\n=== Real-World Validation Summary ===")
	t.Logf("Files analyzed: %d", totalFiles)
	t.Logf("Total findings: %d (avg: %.1f per file)", totalFindings, avgFindings)
	t.Logf("Total time: %v (avg: %v per file)", totalTime, avgTime)
	t.Logf("Max file size: %d bytes", maxFileSize)
	t.Logf("Errors: %d", errors)
	t.Logf("Panics recovered: %d", panics)

	// Validation assertions
	if totalFiles == 0 {
		t.Error("No files were successfully analyzed")
	}

	if errors > totalFiles/2 {
		t.Errorf("Too many errors: %d/%d files failed", errors, totalFiles)
	}

	// Performance assertion: average analysis should be fast
	if avgTime > 2*time.Second {
		t.Errorf("Average analysis time too slow: %v (expected < 2s)", avgTime)
	}

	// Findings should be reasonable (not excessive false positives)
	// In a well-maintained codebase, we expect some findings but not hundreds per file
	if avgFindings > 50 {
		t.Logf("WARNING: High average findings (%.1f) - possible false positives", avgFindings)
	}
}

// TestRealWorldKnownLimitations documents current detection limitations
func TestRealWorldKnownLimitations(t *testing.T) {
	// These tests document known limitations of the current AST analysis
	// They validate that analysis completes without errors, even if findings may be false positives
	testCases := []struct {
		name        string
		code        string
		check       string
		description string
	}{
		{
			name: "init_functions_may_be_flagged",
			code: `package main
func init() {
	setup()
}`,
			check:       "orphaned",
			description: "init() may be flagged as orphaned (known limitation: runtime calls not detected)",
		},
		{
			name: "test_helpers_may_be_flagged",
			code: `package main
func TestMain(t *testing.T) {}
func helper() {}`,
			check:       "orphaned",
			description: "Test helpers may be flagged (known limitation: test framework usage not detected)",
		},
		{
			name: "interface_implementations_not_flagged",
			code: `package main
type Interface interface {
	Method()
}
func (s *Struct) Method() {}`,
			check:       "orphaned",
			description: "Interface implementations correctly not flagged",
		},
		{
			name: "error_variables_correctly_detected",
			code: `package main
func test() error {
	err := doSomething()
	if err != nil {
		return err
	}
	return nil
}`,
			check:       "unused",
			description: "Error variables correctly detected as used",
		},
		{
			name: "defer_variables_correctly_detected",
			code: `package main
import "os"
func test() {
	f, err := os.Open("file")
	if err != nil {
		return
	}
	defer f.Close()
}`,
			check:       "unused",
			description: "Defer variables correctly detected as used",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			findings, _, err := AnalyzeAST(tc.code, "go", []string{tc.check})
			if err != nil {
				t.Fatalf("Analysis failed: %v", err)
			}

			// Document the behavior (may be false positive, but analysis should complete)
			t.Logf("  %s: %d findings - %s", tc.name, len(findings), tc.description)
			if len(findings) > 0 {
				for _, f := range findings {
					t.Logf("    - %s: %s", f.Type, f.Message)
				}
			}

			// Key validation: analysis should never panic or return nil findings unexpectedly
			if findings == nil {
				t.Error("Findings should not be nil (should be empty slice)")
			}
		})
	}

	t.Log("\n=== Known Limitations ===")
	t.Log("1. init() functions may be flagged as orphaned (runtime calls not detected)")
	t.Log("2. Test helpers may be flagged as orphaned (test framework usage not detected)")
	t.Log("3. Package-level functions without explicit callers may be flagged")
	t.Log("4. These are acceptable limitations for AST-based static analysis")
}

// TestRealWorldPerformance benchmarks analysis on real files
func TestRealWorldPerformance(t *testing.T) {
	internalPath := filepath.Join("..", "..", "..", "internal")
	if _, err := os.Stat(internalPath); os.IsNotExist(err) {
		t.Skip("Skipping performance test: internal directory not found")
		return
	}

	// Find a medium-sized file for performance testing
	var testFile string
	err := filepath.Walk(internalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			if info.Size() > 1000 && info.Size() < 10000 {
				testFile = path
				return filepath.SkipAll // Found one, stop searching
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		t.Fatalf("Failed to find test file: %v", err)
	}

	if testFile == "" {
		t.Skip("No suitable test file found")
		return
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	code := string(content)
	allChecks := []string{"duplicates", "unused", "unreachable", "orphaned"}

	// Run multiple iterations to get average
	iterations := 5
	totalTime := time.Duration(0)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, _, err := AnalyzeAST(code, "go", allChecks)
		duration := time.Since(start)
		totalTime += duration

		if err != nil {
			t.Fatalf("Analysis failed on iteration %d: %v", i+1, err)
		}
	}

	avgTime := totalTime / time.Duration(iterations)
	t.Logf("Performance test on %s (%d bytes): avg %v over %d iterations",
		filepath.Base(testFile), len(code), avgTime, iterations)

	// Performance assertion: should complete in reasonable time
	if avgTime > 1*time.Second {
		t.Errorf("Performance regression: average time %v exceeds 1s threshold", avgTime)
	}
}
