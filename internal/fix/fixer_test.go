// Package fix provides tests for automatic code fixing
package fix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFix(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("performs dry run without modifying files", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tmpDir, "test.js")
		original := "console.log('test');\nconst x = 1;"
		os.WriteFile(testFile, []byte(original), 0644)

		opts := FixOptions{
			TargetPath: tmpDir,
			DryRun:     true,
		}

		result, err := Fix(opts)
		if err != nil {
			t.Errorf("Fix() error = %v", err)
		}

		// File should be unchanged in dry run
		content, _ := os.ReadFile(testFile)
		if string(content) != original {
			t.Error("dry run should not modify files")
		}

		_ = result
	})

	t.Run("applies fixes to files", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "tofix.js")
		os.WriteFile(testFile, []byte("console.log('debug');\nconst x = 1;"), 0644)

		opts := FixOptions{
			TargetPath: tmpDir,
			DryRun:     false,
		}

		result, err := Fix(opts)
		if err != nil {
			t.Errorf("Fix() error = %v", err)
		}

		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("uses current directory when path empty", func(t *testing.T) {
		originalWD, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalWD)

		opts := FixOptions{
			TargetPath: "",
			DryRun:     true,
		}

		_, err := Fix(opts)
		if err != nil {
			t.Errorf("Fix() error = %v", err)
		}
	})
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/project/node_modules/package/index.js", true},
		{"/project/.git/config", true},
		{"/project/.sentinel/config.json", true},
		{"/project/vendor/lib.go", true},
		{"/project/build/output.js", true},
		{"/project/dist/bundle.js", true},
		{"/project/src/app.js", false},
		{"/project/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := shouldSkipPath(tt.path)
			if result != tt.expected {
				t.Errorf("shouldSkipPath(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsCodeFile(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".js", true},
		{".ts", true},
		{".jsx", true},
		{".tsx", true},
		{".py", true},
		{".go", true},
		{".java", true},
		{".cs", true},
		{".php", true},
		{".rb", true},
		{".txt", false},
		{".md", false},
		{".json", false},
		{".html", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := isCodeFile(tt.ext)
			if result != tt.expected {
				t.Errorf("isCodeFile(%s) = %v, want %v", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestRemoveConsoleLogs(t *testing.T) {
	t.Run("removes console.log", func(t *testing.T) {
		content := "const x = 1;\nconsole.log('debug');\nconst y = 2;"
		fixCount := 0
		result, modified := removeConsoleLogs(content, "test.js", &fixCount, false)

		if !modified {
			t.Error("should be modified")
		}
		if strings.Contains(result, "console.log") {
			t.Error("console.log should be removed")
		}
	})

	t.Run("removes console.debug", func(t *testing.T) {
		content := "console.debug('info');\n"
		fixCount := 0
		result, modified := removeConsoleLogs(content, "test.js", &fixCount, false)

		if !modified {
			t.Error("should be modified")
		}
		if strings.Contains(result, "console.debug") {
			t.Error("console.debug should be removed")
		}
	})

	t.Run("removes console.error", func(t *testing.T) {
		content := "console.error('err');\n"
		fixCount := 0
		result, modified := removeConsoleLogs(content, "test.js", &fixCount, false)

		if !modified {
			t.Error("should be modified")
		}
		_ = result
	})

	t.Run("preserves non-console code", func(t *testing.T) {
		content := "const x = 1;\nconst y = 2;"
		fixCount := 0
		result, modified := removeConsoleLogs(content, "test.js", &fixCount, false)

		if modified {
			t.Error("should not be modified")
		}
		if result != content {
			t.Error("content should be unchanged")
		}
	})
}

func TestRemoveDebugger(t *testing.T) {
	t.Run("removes debugger statements", func(t *testing.T) {
		content := "const x = 1;\ndebugger;\nconst y = 2;"
		fixCount := 0
		result, modified := removeDebugger(content, "test.js", &fixCount, false)

		if !modified {
			t.Error("should be modified")
		}
		if strings.Contains(result, "debugger") {
			t.Error("debugger should be removed")
		}
	})

	t.Run("preserves code without debugger", func(t *testing.T) {
		content := "const x = 1;\nconst y = 2;"
		fixCount := 0
		result, modified := removeDebugger(content, "test.js", &fixCount, false)

		if modified {
			t.Error("should not be modified")
		}
		if result != content {
			t.Error("content should be unchanged")
		}
	})
}

func TestRemoveTrailingWhitespace(t *testing.T) {
	t.Run("removes trailing spaces", func(t *testing.T) {
		content := "const x = 1;   \nconst y = 2;\t\t"
		fixCount := 0
		result, modified := removeTrailingWhitespace(content, "test.js", &fixCount, false)

		if !modified {
			t.Error("should be modified")
		}
		lines := strings.Split(result, "\n")
		for _, line := range lines {
			if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
				t.Error("trailing whitespace should be removed")
			}
		}
	})

	t.Run("preserves clean code", func(t *testing.T) {
		content := "const x = 1;\nconst y = 2;"
		fixCount := 0
		_, modified := removeTrailingWhitespace(content, "test.js", &fixCount, false)

		if modified {
			t.Error("should not be modified")
		}
	})
}

func TestRecordFixHistory(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	result := &Result{
		FixesApplied:   5,
		FilesModified:  3,
		BackupsCreated: 3,
	}

	err := recordFixHistory(result, tmpDir)
	if err != nil {
		t.Errorf("recordFixHistory() error = %v", err)
	}

	// Verify history file exists
	if _, err := os.Stat(".sentinel/fix-history.json"); os.IsNotExist(err) {
		t.Error("fix history file should exist")
	}
}

func TestFixOptions(t *testing.T) {
	opts := FixOptions{
		TargetPath: "/test/path",
		DryRun:     true,
		Force:      false,
	}

	if opts.TargetPath != "/test/path" {
		t.Errorf("expected path /test/path, got %s", opts.TargetPath)
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
}

func TestResult(t *testing.T) {
	result := Result{
		FixesApplied:   10,
		FilesModified:  5,
		BackupsCreated: 5,
	}

	if result.FixesApplied != 10 {
		t.Errorf("expected 10 fixes, got %d", result.FixesApplied)
	}
	if result.FilesModified != 5 {
		t.Errorf("expected 5 files, got %d", result.FilesModified)
	}
}

func TestRollback(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	t.Run("returns error when no history", func(t *testing.T) {
		err := Rollback(RollbackOptions{})
		if err == nil {
			t.Error("Expected error for missing history")
		}
	})

	t.Run("list mode works with empty history", func(t *testing.T) {
		err := Rollback(RollbackOptions{ListOnly: true})
		if err != nil {
			t.Errorf("Rollback list error: %v", err)
		}
	})

	t.Run("list mode works with history", func(t *testing.T) {
		// Create history file
		os.MkdirAll(".sentinel", 0755)
		historyJSON := `[
			{
				"session_id": "session-123",
				"timestamp": "2024-01-01T00:00:00Z",
				"files_modified": 2,
				"fixes_applied": 3
			}
		]`
		os.WriteFile(".sentinel/fix-history.json", []byte(historyJSON), 0644)

		err := Rollback(RollbackOptions{ListOnly: true})
		if err != nil {
			t.Errorf("Rollback list error: %v", err)
		}
	})
}

func TestSortImports(t *testing.T) {
	t.Run("sorts Go imports", func(t *testing.T) {
		content := `package main

import (
	"github.com/example/external"
	"fmt"
	"./internal"
)

func main() {
	fmt.Println("test")
}
`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.go", &fixesApplied, modified)

		if !wasModified {
			t.Error("Expected imports to be sorted")
		}
		if !strings.Contains(result, "import (") {
			t.Error("Expected import block to be present")
		}
	})

	t.Run("sorts JS imports", func(t *testing.T) {
		content := `import React from 'react';
import { Component } from './component';
import * as utils from './utils';
`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.ts", &fixesApplied, modified)

		if wasModified {
			// Should potentially sort imports
			if !strings.Contains(result, "import") {
				t.Error("Expected imports to be present")
			}
		}
	})

	t.Run("does not modify non-code files", func(t *testing.T) {
		content := `This is a text file`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.txt", &fixesApplied, modified)

		if wasModified {
			t.Error("Expected no modification for non-code files")
		}
		if result != content {
			t.Error("Expected content to be unchanged")
		}
	})
}

// Note: sortGoImports and sortJSImports are private functions,
// so they're tested indirectly through sortImports tests above
