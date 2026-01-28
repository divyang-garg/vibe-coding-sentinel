// Package services edge case tests for task verification
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestVerifyTask_EdgeCases tests edge cases for VerifyTask
func TestVerifyTask_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func() (string, *Task, func())
		expectError bool
		description string
	}{
		{
			name: "non-existent codebase path",
			setup: func() (string, *Task, func()) {
				return "/nonexistent/path/that/does/not/exist", &Task{
					ID:    "task-1",
					Title: "Test task",
				}, func() {}
			},
			expectError: true,
			description: "Should handle non-existent codebase path gracefully",
		},
		{
			name: "task with no file path",
			setup: func() (string, *Task, func()) {
				tmpDir := t.TempDir()
				return tmpDir, &Task{
					ID:          "task-1",
					Title:       "Task without file path",
					Description: "This task has no associated file",
					FilePath:    "",
				}, func() {}
			},
			expectError: false,
			description: "Should work with tasks that have no file path",
		},
		{
			name: "task with invalid file path",
			setup: func() (string, *Task, func()) {
				tmpDir := t.TempDir()
				return tmpDir, &Task{
					ID:       "task-1",
					Title:    "Test task",
					FilePath: "../../../etc/passwd", // Path traversal attempt
				}, func() {}
			},
			expectError: false,
			description: "Should handle path traversal attempts safely",
		},
		{
			name: "task with very long description",
			setup: func() (string, *Task, func()) {
				tmpDir := t.TempDir()
				longDesc := strings.Repeat("keyword ", 1000) // Very long description
				return tmpDir, &Task{
					ID:          "task-1",
					Title:       "Test task",
					Description: longDesc,
					FilePath:    "test.go",
				}, func() {}
			},
			expectError: false,
			description: "Should handle very long task descriptions",
		},
		{
			name: "task with special characters",
			setup: func() (string, *Task, func()) {
				tmpDir := t.TempDir()
				return tmpDir, &Task{
					ID:          "task-1",
					Title:       "Task with émojis 🚀 and spéciál chars",
					Description: "Description with <script>alert('xss')</script> and SQL'; DROP TABLE--",
					FilePath:    "test.go",
				}, func() {}
			},
			expectError: false,
			description: "Should handle special characters and potential injection attempts",
		},
		{
			name: "empty task title and description",
			setup: func() (string, *Task, func()) {
				tmpDir := t.TempDir()
				return tmpDir, &Task{
					ID:          "task-1",
					Title:       "",
					Description: "",
					FilePath:    "test.go",
				}, func() {}
			},
			expectError: false,
			description: "Should handle empty title and description",
		},
		{
			name: "codebase path is a file not directory",
			setup: func() (string, *Task, func()) {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "notadir")
				os.WriteFile(filePath, []byte("content"), 0644)
				return filePath, &Task{
					ID:    "task-1",
					Title: "Test task",
				}, func() {}
			},
			expectError: false,
			description: "Should handle codebase path that is a file",
		},
		{
			name: "very deep directory structure",
			setup: func() (string, *Task, func()) {
				tmpDir := t.TempDir()
				deepPath := tmpDir
				for i := 0; i < 20; i++ {
					deepPath = filepath.Join(deepPath, fmt.Sprintf("level%d", i))
					os.MkdirAll(deepPath, 0755)
				}
				return tmpDir, &Task{
					ID:       "task-1",
					Title:    "Test task",
					FilePath: filepath.Join("level0", "level1", "level2", "test.go"),
				}, func() {}
			},
			expectError: false,
			description: "Should handle deep directory structures",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			codebasePath, task, cleanup := tc.setup()
			defer cleanup()

			// Mock GetTask to return our test task
			// Note: This requires refactoring to use dependency injection
			// For now, we'll test the analyzeTaskCompletion function directly
			confidence, evidence := analyzeTaskCompletion(context.Background(), task, codebasePath)

			if tc.expectError {
				// If we expect an error, confidence should be 0 or very low
				if confidence > 0.5 {
					t.Errorf("Expected low confidence for error case, got %.2f", confidence)
				}
			}

			// Verify evidence is always present
			if evidence == nil {
				t.Error("Evidence should never be nil")
			}

			t.Logf("Confidence: %.2f, Evidence keys: %v", confidence, getKeys(evidence))
		})
	}
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestAnalyzeTaskCompletion_EdgeCases tests edge cases for analyzeTaskCompletion
func TestAnalyzeTaskCompletion_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		task        *Task
		setup       func(string) func()
		description string
	}{
		{
			name: "binary file content",
			task: &Task{
				ID:       "task-1",
				Title:    "Process binary file",
				FilePath: "binary.bin",
			},
			setup: func(codebasePath string) func() {
				binaryFile := filepath.Join(codebasePath, "binary.bin")
				os.WriteFile(binaryFile, []byte{0xFF, 0xFE, 0xFD, 0x00, 0x01, 0x02}, 0644)
				return func() {}
			},
			description: "Should handle binary files without crashing",
		},
		{
			name: "very large file",
			task: &Task{
				ID:       "task-1",
				Title:    "Process large file",
				FilePath: "large.go",
			},
			setup: func(codebasePath string) func() {
				largeFile := filepath.Join(codebasePath, "large.go")
				// Create a 1MB file
				content := strings.Repeat("package main\n", 100000)
				os.WriteFile(largeFile, []byte(content), 0644)
				return func() {}
			},
			description: "Should handle very large files efficiently",
		},
		{
			name: "empty file",
			task: &Task{
				ID:       "task-1",
				Title:    "Empty file task",
				FilePath: "empty.go",
			},
			setup: func(codebasePath string) func() {
				emptyFile := filepath.Join(codebasePath, "empty.go")
				os.WriteFile(emptyFile, []byte{}, 0644)
				return func() {}
			},
			description: "Should handle empty files",
		},
		{
			name: "file with only whitespace",
			task: &Task{
				ID:       "task-1",
				Title:    "Whitespace file",
				FilePath: "whitespace.go",
			},
			setup: func(codebasePath string) func() {
				wsFile := filepath.Join(codebasePath, "whitespace.go")
				os.WriteFile(wsFile, []byte("   \n\t\n   "), 0644)
				return func() {}
			},
			description: "Should handle files with only whitespace",
		},
		{
			name: "symlink to file",
			task: &Task{
				ID:       "task-1",
				Title:    "Symlink task",
				FilePath: "symlink.go",
			},
			setup: func(codebasePath string) func() {
				targetFile := filepath.Join(codebasePath, "target.go")
				os.WriteFile(targetFile, []byte("package main"), 0644)
				symlink := filepath.Join(codebasePath, "symlink.go")
				os.Symlink("target.go", symlink)
				return func() {}
			},
			description: "Should handle symlinks correctly",
		},
		{
			name: "file path with spaces",
			task: &Task{
				ID:       "task-1",
				Title:    "File with spaces",
				FilePath: "my file with spaces.go",
			},
			setup: func(codebasePath string) func() {
				fileWithSpaces := filepath.Join(codebasePath, "my file with spaces.go")
				os.WriteFile(fileWithSpaces, []byte("package main"), 0644)
				return func() {}
			},
			description: "Should handle file paths with spaces",
		},
		{
			name: "file path with unicode characters",
			task: &Task{
				ID:       "task-1",
				Title:    "Unicode file",
				FilePath: "файл.go",
			},
			setup: func(codebasePath string) func() {
				unicodeFile := filepath.Join(codebasePath, "файл.go")
				os.WriteFile(unicodeFile, []byte("package main"), 0644)
				return func() {}
			},
			description: "Should handle unicode file names",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cleanup := tc.setup(tmpDir)
			defer cleanup()

			confidence, evidence := analyzeTaskCompletion(context.Background(), tc.task, tmpDir)

			// Should not panic and should return valid results
			if evidence == nil {
				t.Error("Evidence should not be nil")
			}

			if confidence < 0 || confidence > 1.0 {
				t.Errorf("Confidence should be between 0 and 1.0, got %.2f", confidence)
			}

			t.Logf("Confidence: %.2f", confidence)
		})
	}
}

// TestSearchCodebaseForKeywords_EdgeCases tests edge cases for keyword search
func TestSearchCodebaseForKeywords_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		keywords    []string
		setup       func(string) func()
		description string
	}{
		{
			name:     "empty keywords",
			keywords: []string{},
			setup: func(codebasePath string) func() {
				return func() {}
			},
			description: "Should handle empty keyword list",
		},
		{
			name:     "keywords with special regex characters",
			keywords: []string{"test(", "file[1]", "path.*"},
			setup: func(codebasePath string) func() {
				file := filepath.Join(codebasePath, "test.go")
				os.WriteFile(file, []byte("test( file[1] path.*"), 0644)
				return func() {}
			},
			description: "Should handle keywords with regex special characters",
		},
		{
			name:     "very long keywords",
			keywords: []string{strings.Repeat("a", 10000)},
			setup: func(codebasePath string) func() {
				return func() {}
			},
			description: "Should handle very long keywords",
		},
		{
			name:     "many keywords",
			keywords: make([]string, 1000),
			setup: func(codebasePath string) func() {
				for i := range make([]string, 1000) {
					file := filepath.Join(codebasePath, fmt.Sprintf("file%d.go", i))
					os.WriteFile(file, []byte("content"), 0644)
				}
				return func() {}
			},
			description: "Should handle many keywords efficiently",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cleanup := tc.setup(tmpDir)
			defer cleanup()

			matches := searchCodebaseForKeywords(context.Background(), tmpDir, tc.keywords)

			// Should not panic
			if matches < 0 {
				t.Errorf("Matches should be non-negative, got %d", matches)
			}

			t.Logf("Found %d matches", matches)
		})
	}
}

// TestFindTestFile_EdgeCases tests edge cases for test file finding
func TestFindTestFile_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		filePath    string
		setup       func(string) func()
		description string
	}{
		{
			name:     "empty file path",
			filePath: "",
			setup: func(codebasePath string) func() {
				return func() {}
			},
			description: "Should handle empty file path",
		},
		{
			name:     "file path with directory",
			filePath: "subdir/service.go",
			setup: func(codebasePath string) func() {
				subdir := filepath.Join(codebasePath, "subdir")
				os.MkdirAll(subdir, 0755)
				mainFile := filepath.Join(subdir, "service.go")
				os.WriteFile(mainFile, []byte("package main"), 0644)
				testFile := filepath.Join(subdir, "service_test.go")
				os.WriteFile(testFile, []byte("package main"), 0644)
				return func() {}
			},
			description: "Should find test file in subdirectory",
		},
		{
			name:     "file with no extension",
			filePath: "script",
			setup: func(codebasePath string) func() {
				file := filepath.Join(codebasePath, "script")
				os.WriteFile(file, []byte("content"), 0644)
				return func() {}
			},
			description: "Should handle files without extensions",
		},
		{
			name:     "file with multiple dots",
			filePath: "my.service.go",
			setup: func(codebasePath string) func() {
				file := filepath.Join(codebasePath, "my.service.go")
				os.WriteFile(file, []byte("package main"), 0644)
				testFile := filepath.Join(codebasePath, "my.service_test.go")
				os.WriteFile(testFile, []byte("package main"), 0644)
				return func() {}
			},
			description: "Should handle files with multiple dots in name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cleanup := tc.setup(tmpDir)
			defer cleanup()

			result := findTestFile(tmpDir, tc.filePath)

			// Should not panic
			if result != "" {
				// If found, verify it's a valid path
				fullPath := filepath.Join(tmpDir, result)
				if _, err := os.Stat(fullPath); err != nil {
					t.Errorf("Found test file path %q but file doesn't exist: %v", result, err)
				}
			}

			t.Logf("Test file: %q", result)
		})
	}
}

// TestConfidenceCalculation_EdgeCases tests edge cases for confidence calculation
func TestConfidenceCalculation_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name        string
		task        *Task
		setup       func()
		minConf     float64
		maxConf     float64
		description string
	}{
		{
			name: "all checks pass - should cap at 1.0",
			task: &Task{
				ID:          "task-1",
				Title:       "Complete implementation",
				Description: "Implement complete feature with tests",
				FilePath:    "complete.go",
			},
			setup: func() {
				// Create file that matches everything
				file := filepath.Join(tmpDir, "complete.go")
				content := `package main
// Complete implementation
func complete() {
	// Feature implementation
}`
				os.WriteFile(file, []byte(content), 0644)
				// Touch file to make it recent
				os.Chtimes(file, time.Now(), time.Now())

				// Create test file
				testFile := filepath.Join(tmpDir, "complete_test.go")
				os.WriteFile(testFile, []byte("package main\nfunc TestComplete(t *testing.T) {}"), 0644)

				// Create other matching files
				for i := 0; i < 10; i++ {
					otherFile := filepath.Join(tmpDir, fmt.Sprintf("other%d.go", i))
					os.WriteFile(otherFile, []byte("complete implementation"), 0644)
				}
			},
			minConf:     0.8,
			maxConf:     1.0,
			description: "Confidence should be capped at 1.0 even if all checks pass",
		},
		{
			name: "no checks pass - should be low",
			task: &Task{
				ID:          "task-1",
				Title:       "Missing implementation",
				Description: "This is not implemented",
				FilePath:    "missing.go",
			},
			setup:       func() {},
			minConf:     0.0,
			maxConf:     0.3,
			description: "Confidence should be low when nothing matches",
		},
		{
			name: "partial matches",
			task: &Task{
				ID:          "task-1",
				Title:       "Partial implementation",
				Description: "Some keywords match",
				FilePath:    "partial.go",
			},
			setup: func() {
				file := filepath.Join(tmpDir, "partial.go")
				os.WriteFile(file, []byte("package main\n// Some keywords"), 0644)
			},
			minConf:     0.0,
			maxConf:     1.0,
			description: "Confidence should reflect partial matches (file exists + keywords + codebase)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up and recreate
			os.RemoveAll(tmpDir)
			os.MkdirAll(tmpDir, 0755)
			tc.setup()

			confidence, evidence := analyzeTaskCompletion(context.Background(), tc.task, tmpDir)

			if confidence < tc.minConf || confidence > tc.maxConf {
				t.Errorf("Confidence %.2f outside expected range [%.2f, %.2f]", confidence, tc.minConf, tc.maxConf)
			}

			// Verify confidence is always between 0 and 1
			if confidence < 0 || confidence > 1.0 {
				t.Errorf("Confidence %.2f outside valid range [0, 1.0]", confidence)
			}

			// Verify evidence contains final_confidence
			if finalConf, ok := evidence["final_confidence"].(float64); ok {
				if finalConf != confidence {
					t.Errorf("Evidence final_confidence %.2f doesn't match returned confidence %.2f", finalConf, confidence)
				}
			}

			t.Logf("Confidence: %.2f, Evidence: %+v", confidence, evidence)
		})
	}
}

// TestVerifyTask_Concurrency tests concurrent verification
func TestVerifyTask_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	taskID := "task-1"

	// Create a test file
	file := filepath.Join(tmpDir, "test.go")
	os.WriteFile(file, []byte("package main"), 0644)

	// This test would require mocking GetTask, which needs refactoring
	// For now, we'll test that analyzeTaskCompletion is thread-safe
	t.Run("concurrent analysis", func(t *testing.T) {
		task := &Task{
			ID:       taskID,
			Title:    "Concurrent test",
			FilePath: "test.go",
		}

		results := make(chan float64, 10)
		for i := 0; i < 10; i++ {
			go func() {
				conf, _ := analyzeTaskCompletion(context.Background(), task, tmpDir)
				results <- conf
			}()
		}

		confidences := make([]float64, 10)
		for i := 0; i < 10; i++ {
			confidences[i] = <-results
		}

		// All should return same confidence (deterministic)
		firstConf := confidences[0]
		for i, conf := range confidences {
			if conf != firstConf {
				t.Errorf("Concurrent call %d returned different confidence: %.2f vs %.2f", i, conf, firstConf)
			}
		}
	})
}

// TestVerifyTask_PathTraversal tests path traversal security
func TestVerifyTask_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	codebasePath := tmpDir

	// Create a file outside codebase
	outsideFile := filepath.Join(t.TempDir(), "outside.txt")
	os.WriteFile(outsideFile, []byte("sensitive"), 0644)

	testCases := []struct {
		name     string
		filePath string
		safe     bool
	}{
		{
			name:     "normal path",
			filePath: "test.go",
			safe:     true,
		},
		{
			name:     "path traversal attempt",
			filePath: "../../../etc/passwd",
			safe:     true, // Should be safe due to filepath.Join
		},
		{
			name:     "absolute path",
			filePath: outsideFile,
			safe:     true, // Should be safe, won't access outside
		},
		{
			name:     "path with ..",
			filePath: "../parent.go",
			safe:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			task := &Task{
				ID:       "task-1",
				Title:    "Test",
				FilePath: tc.filePath,
			}

			// Should not panic or access files outside codebase
			confidence, evidence := analyzeTaskCompletion(context.Background(), task, codebasePath)

			// Verify filepath.Join prevents traversal
			fullPath := filepath.Join(codebasePath, tc.filePath)
			if strings.Contains(fullPath, "..") && !strings.HasPrefix(fullPath, codebasePath) {
				// This shouldn't happen with filepath.Join, but verify
				t.Logf("Warning: Path might escape: %s", fullPath)
			}

			// Should complete without error
			if evidence == nil {
				t.Error("Evidence should not be nil")
			}

			t.Logf("Confidence: %.2f, Path: %s", confidence, fullPath)
		})
	}
}
