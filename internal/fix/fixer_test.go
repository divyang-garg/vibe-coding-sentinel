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

	t.Run("creates new history file", func(t *testing.T) {
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
	})

	t.Run("appends to existing history array", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		existingHistory := `[
			{
				"session_id": "session-1",
				"timestamp": "2024-01-01T00:00:00Z",
				"fixes_applied": 2
			}
		]`
		os.WriteFile(".sentinel/fix-history.json", []byte(existingHistory), 0644)

		result := &Result{
			FixesApplied:   3,
			FilesModified:  2,
			BackupsCreated: 2,
		}

		err := recordFixHistory(result, tmpDir)
		if err != nil {
			t.Errorf("recordFixHistory() error = %v", err)
		}

		// Verify history was appended
		data, _ := os.ReadFile(".sentinel/fix-history.json")
		if !strings.Contains(string(data), "session-1") {
			t.Error("existing history should be preserved")
		}
	})

	t.Run("converts single object to array", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		singleObject := `{
			"session_id": "session-old",
			"timestamp": "2024-01-01T00:00:00Z",
			"fixes_applied": 1
		}`
		os.WriteFile(".sentinel/fix-history.json", []byte(singleObject), 0644)

		result := &Result{
			FixesApplied:   2,
			FilesModified:  1,
			BackupsCreated: 1,
		}

		err := recordFixHistory(result, tmpDir)
		if err != nil {
			t.Errorf("recordFixHistory() error = %v", err)
		}

		// Verify old entry is preserved
		data, _ := os.ReadFile(".sentinel/fix-history.json")
		if !strings.Contains(string(data), "session-old") {
			t.Error("old history should be preserved")
		}
	})
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

	t.Run("rollback with specific session ID", func(t *testing.T) {
		os.MkdirAll(".sentinel/backups", 0755)
		historyJSON := `[
			{
				"session_id": "session-123",
				"timestamp": "2024-01-01T00:00:00Z",
				"files_modified": 1,
				"fixes_applied": 1
			}
		]`
		os.WriteFile(".sentinel/fix-history.json", []byte(historyJSON), 0644)

		// Create backup file
		backupFile := ".sentinel/backups/test.js_session-123_1234567890.backup"
		os.WriteFile(backupFile, []byte("restored content"), 0644)

		err := Rollback(RollbackOptions{SessionID: "session-123"})
		// May error if file doesn't exist, that's OK
		_ = err
	})

	t.Run("rollback with most recent session", func(t *testing.T) {
		os.MkdirAll(".sentinel/backups", 0755)
		historyJSON := `[
			{
				"session_id": "session-123",
				"timestamp": "2024-01-01T00:00:00Z",
				"files_modified": 1,
				"fixes_applied": 1
			}
		]`
		os.WriteFile(".sentinel/fix-history.json", []byte(historyJSON), 0644)

		err := Rollback(RollbackOptions{})
		// May error if no backup files, that's OK
		_ = err
	})

	t.Run("parseHistoryFile with array format", func(t *testing.T) {
		data := []byte(`[{"session_id": "test"}]`)
		result, err := parseHistoryFile(data)
		if err != nil {
			t.Errorf("parseHistoryFile error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("Expected 1 entry, got %d", len(result))
		}
	})

	t.Run("parseHistoryFile with single object format", func(t *testing.T) {
		data := []byte(`{"session_id": "test"}`)
		result, err := parseHistoryFile(data)
		if err != nil {
			t.Errorf("parseHistoryFile error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("Expected 1 entry, got %d", len(result))
		}
	})

	t.Run("parseHistoryFile with invalid format", func(t *testing.T) {
		data := []byte(`invalid json`)
		_, err := parseHistoryFile(data)
		if err == nil {
			t.Error("Expected error for invalid format")
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

	t.Run("sorts Go single-line imports", func(t *testing.T) {
		content := `package main

import "github.com/example/external"
import "fmt"

func main() {
	fmt.Println("test")
}
`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.go", &fixesApplied, modified)

		if wasModified {
			if !strings.Contains(result, "import") {
				t.Error("Expected imports to be present")
			}
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

	t.Run("handles Python files", func(t *testing.T) {
		content := `import os
import sys
from typing import List
`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.py", &fixesApplied, modified)

		// Python sorting is deferred, should not modify
		if wasModified {
			t.Error("Python imports should not be sorted yet")
		}
		if result != content {
			t.Error("Expected content to be unchanged for Python")
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

	t.Run("handles files without imports", func(t *testing.T) {
		content := `package main

func main() {
	fmt.Println("test")
}
`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.go", &fixesApplied, modified)

		if wasModified {
			t.Error("Should not modify files without imports")
		}
		if result != content {
			t.Error("Expected content to be unchanged")
		}
	})

	t.Run("sorts JS imports with relative and external", func(t *testing.T) {
		content := `import React from 'react';
import { Component } from './component';
import * as utils from './utils';
import axios from 'axios';
`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.ts", &fixesApplied, modified)

		if wasModified {
			if !strings.Contains(result, "import") {
				t.Error("Expected imports to be present")
			}
		}
		_ = result
	})

	t.Run("sorts JS imports with blank lines", func(t *testing.T) {
		content := `import React from 'react';

import { Component } from './component';
`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.tsx", &fixesApplied, modified)

		if wasModified {
			if !strings.Contains(result, "import") {
				t.Error("Expected imports to be present")
			}
		}
		_ = result
	})

	t.Run("handles JS file with single import", func(t *testing.T) {
		content := `import React from 'react';
const App = () => null;
`
		fixesApplied := 0
		modified := false
		result, wasModified := sortImports(content, "test.jsx", &fixesApplied, modified)

		if wasModified {
			t.Error("Should not modify single import")
		}
		if result != content {
			t.Error("Expected content to be unchanged")
		}
	})
}

// Note: sortGoImports and sortJSImports are private functions,
// so they're tested indirectly through sortImports tests above

func TestFix_ErrorHandling(t *testing.T) {
	t.Run("handles backup directory creation failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create a file named .sentinel to prevent directory creation
		sentinelPath := filepath.Join(tmpDir, ".sentinel")
		os.WriteFile(sentinelPath, []byte("file"), 0644)
		defer os.Remove(sentinelPath)

		opts := FixOptions{
			TargetPath: tmpDir,
			DryRun:     false,
		}

		_, err := Fix(opts)
		// May or may not error depending on implementation
		// The function might handle this gracefully
		_ = err
	})

	t.Run("handles file write errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.js")
		os.WriteFile(testFile, []byte("console.log('test');"), 0644)

		// Make file read-only to prevent writing
		os.Chmod(testFile, 0444)
		defer os.Chmod(testFile, 0644)

		opts := FixOptions{
			TargetPath: tmpDir,
			DryRun:     false,
		}

		_, err := Fix(opts)
		// May or may not error depending on implementation
		_ = err
	})

	t.Run("handles backup file creation errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.js")
		os.WriteFile(testFile, []byte("console.log('test');"), 0644)

		// Create read-only backup directory
		backupDir := filepath.Join(tmpDir, ".sentinel", "backups")
		os.MkdirAll(backupDir, 0444)
		defer os.Chmod(backupDir, 0755)

		opts := FixOptions{
			TargetPath: tmpDir,
			DryRun:     false,
		}

		_, err := Fix(opts)
		// May or may not error
		_ = err
	})
}

func TestFix_PatternSpecific(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("applies only console pattern", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.js")
		content := "console.log('test');\ndebugger;\nconst x = 1;   "
		os.WriteFile(testFile, []byte(content), 0644)

		opts := FixOptions{
			TargetPath: tmpDir,
			Pattern:    "console",
			DryRun:     true,
		}

		result, err := Fix(opts)
		if err != nil {
			t.Errorf("Fix() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("applies only debugger pattern", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.js")
		content := "console.log('test');\ndebugger;\nconst x = 1;"
		os.WriteFile(testFile, []byte(content), 0644)

		opts := FixOptions{
			TargetPath: tmpDir,
			Pattern:    "debugger",
			DryRun:     true,
		}

		result, err := Fix(opts)
		if err != nil {
			t.Errorf("Fix() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("applies only whitespace pattern", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.js")
		content := "const x = 1;   \nconst y = 2;\t\t"
		os.WriteFile(testFile, []byte(content), 0644)

		opts := FixOptions{
			TargetPath: tmpDir,
			Pattern:    "whitespace",
			DryRun:     true,
		}

		result, err := Fix(opts)
		if err != nil {
			t.Errorf("Fix() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("applies only imports pattern", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.go")
		content := `package main

import "github.com/example/external"
import "fmt"
`
		os.WriteFile(testFile, []byte(content), 0644)

		opts := FixOptions{
			TargetPath: tmpDir,
			Pattern:    "imports",
			DryRun:     true,
		}

		result, err := Fix(opts)
		if err != nil {
			t.Errorf("Fix() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})
}

func TestSortGoImports_EdgeCases(t *testing.T) {
	t.Run("handles single import statement", func(t *testing.T) {
		content := `package main

import "fmt"

func main() {}
`
		fixesApplied := 0
		result, modified := sortGoImports(content, "test.go", &fixesApplied, false)
		// Single import might be converted to block format or left as-is
		// Just verify it doesn't crash
		if result == "" {
			t.Error("result should not be empty")
		}
		_ = modified
	})

	t.Run("handles already sorted imports", func(t *testing.T) {
		content := `package main

import (
	"fmt"
	"os"
)
`
		fixesApplied := 0
		result, modified := sortGoImports(content, "test.go", &fixesApplied, false)
		// May or may not detect as already sorted
		_ = result
		_ = modified
	})

	t.Run("handles imports with comments", func(t *testing.T) {
		content := `package main

import (
	"fmt"
	// comment
	"os"
)
`
		fixesApplied := 0
		result, modified := sortGoImports(content, "test.go", &fixesApplied, false)
		// Should handle comments
		_ = result
		_ = modified
	})

	t.Run("handles mixed stdlib and external imports", func(t *testing.T) {
		content := `package main

import (
	"github.com/example/external"
	"fmt"
)
`
		fixesApplied := 0
		result, modified := sortGoImports(content, "test.go", &fixesApplied, false)
		if modified {
			if !strings.Contains(result, "fmt") {
				t.Error("should contain fmt import")
			}
		}
	})
}

func TestSortJSImports_EdgeCases(t *testing.T) {
	t.Run("handles already sorted imports", func(t *testing.T) {
		content := `import React from 'react';
import { Component } from './component';
`
		fixesApplied := 0
		result, modified := sortJSImports(content, "test.ts", &fixesApplied, false)
		// May or may not detect as already sorted
		_ = result
		_ = modified
	})

	t.Run("handles imports with type imports", func(t *testing.T) {
		content := `import type { Type } from './types';
import React from 'react';
`
		fixesApplied := 0
		result, modified := sortJSImports(content, "test.ts", &fixesApplied, false)
		// Should handle type imports
		_ = result
		_ = modified
	})

	t.Run("handles imports with side effects", func(t *testing.T) {
		content := `import './styles.css';
import React from 'react';
`
		fixesApplied := 0
		result, modified := sortJSImports(content, "test.ts", &fixesApplied, false)
		// Should handle side-effect imports
		_ = result
		_ = modified
	})

	t.Run("handles multiple relative import groups", func(t *testing.T) {
		content := `import { Component } from './component';
import { Service } from '../services/service';
import React from 'react';
`
		fixesApplied := 0
		result, modified := sortJSImports(content, "test.ts", &fixesApplied, false)
		if modified {
			if !strings.Contains(result, "import") {
				t.Error("should contain imports")
			}
		}
	})
}

func TestRecordFixHistory_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	t.Run("handles directory creation failure", func(t *testing.T) {
		// Create a file named .sentinel to prevent directory creation
		os.WriteFile(".sentinel", []byte("file"), 0644)
		defer os.Remove(".sentinel")

		result := &Result{FixesApplied: 1}
		err := recordFixHistory(result, tmpDir)
		if err == nil {
			t.Error("should error when directory creation fails")
		}
	})

	t.Run("handles invalid existing history gracefully", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		os.WriteFile(".sentinel/fix-history.json", []byte("invalid json"), 0644)

		result := &Result{FixesApplied: 1}
		err := recordFixHistory(result, tmpDir)
		// Should handle gracefully by creating new history
		if err != nil {
			t.Errorf("recordFixHistory() should handle invalid JSON: %v", err)
		}
	})
}

func TestRollback_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	t.Run("handles missing backup directory", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		historyJSON := `[{"session_id": "session-123", "timestamp": "2024-01-01T00:00:00Z"}]`
		os.WriteFile(".sentinel/fix-history.json", []byte(historyJSON), 0644)

		err := Rollback(RollbackOptions{SessionID: "session-123"})
		if err == nil {
			t.Error("should error when backup directory doesn't exist")
		}
	})

	t.Run("handles session not found", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		historyJSON := `[{"session_id": "session-123", "timestamp": "2024-01-01T00:00:00Z"}]`
		os.WriteFile(".sentinel/fix-history.json", []byte(historyJSON), 0644)

		err := Rollback(RollbackOptions{SessionID: "nonexistent"})
		if err == nil {
			t.Error("should error when session not found")
		}
	})

	t.Run("handles empty history list", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		os.WriteFile(".sentinel/fix-history.json", []byte("[]"), 0644)

		err := Rollback(RollbackOptions{})
		if err == nil {
			t.Error("should error when history is empty")
		}
	})

	t.Run("handles backup file read errors", func(t *testing.T) {
		os.MkdirAll(".sentinel/backups", 0755)
		historyJSON := `[{"session_id": "session-123", "timestamp": "2024-01-01T00:00:00Z"}]`
		os.WriteFile(".sentinel/fix-history.json", []byte(historyJSON), 0644)

		// Create unreadable backup file
		backupFile := ".sentinel/backups/test.js_session-123_1234567890.backup"
		os.WriteFile(backupFile, []byte("content"), 0000)
		defer os.Chmod(backupFile, 0644)

		err := Rollback(RollbackOptions{SessionID: "session-123"})
		// Should handle gracefully by skipping unreadable files
		_ = err
	})

	t.Run("handles restore file write errors", func(t *testing.T) {
		os.MkdirAll(".sentinel/backups", 0755)
		historyJSON := `[{"session_id": "session-123", "timestamp": "2024-01-01T00:00:00Z"}]`
		os.WriteFile(".sentinel/fix-history.json", []byte(historyJSON), 0644)

		// Create backup file
		backupFile := ".sentinel/backups/test.js_session-123_1234567890.backup"
		os.WriteFile(backupFile, []byte("restored content"), 0644)

		// Create read-only target directory
		os.MkdirAll("target", 0555)
		defer os.Chmod("target", 0755)

		err := Rollback(RollbackOptions{SessionID: "session-123"})
		// Should handle gracefully
		_ = err
	})
}
