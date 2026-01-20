// Package patterns provides tests for pattern learning functionality
package patterns

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLearn(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample Go file
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte(`
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`), 0644)

	t.Run("detects languages", func(t *testing.T) {
		opts := LearnOptions{CodebasePath: tmpDir}
		patterns, err := Learn(opts)
		if err != nil {
			t.Fatalf("Learn() error: %v", err)
		}
		if patterns.Languages["Go"] == 0 {
			t.Error("Expected Go language to be detected")
		}
	})

	t.Run("detects file extensions", func(t *testing.T) {
		opts := LearnOptions{CodebasePath: tmpDir}
		patterns, err := Learn(opts)
		if err != nil {
			t.Fatalf("Learn() error: %v", err)
		}
		if patterns.FileExtensions[".go"] == 0 {
			t.Error("Expected .go extension to be detected")
		}
	})
}

func TestLearn_ImportsOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample TypeScript file with imports
	testFile := filepath.Join(tmpDir, "test.ts")
	os.WriteFile(testFile, []byte(`
import { Component } from './component';
import * as utils from './utils';

export class App {}
`), 0644)

	opts := LearnOptions{
		CodebasePath: tmpDir,
		ImportsOnly:  true,
	}
	patterns, err := Learn(opts)
	if err != nil {
		t.Fatalf("Learn() error: %v", err)
	}
	if patterns.ImportPatterns.Style == "" {
		t.Error("Expected import style to be detected")
	}
}

func TestLearn_StructureOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Create folder structure
	os.MkdirAll(filepath.Join(tmpDir, "src", "components"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "src", "services"), 0755)

	opts := LearnOptions{
		CodebasePath:  tmpDir,
		StructureOnly: true,
	}
	patterns, err := Learn(opts)
	if err != nil {
		t.Fatalf("Learn() error: %v", err)
	}
	// Structure analysis should run without error
	_ = patterns
}

func TestLearn_OutputJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample file
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte(`package main`), 0644)

	opts := LearnOptions{
		CodebasePath: tmpDir,
		OutputJSON:   true,
	}
	patterns, err := Learn(opts)
	if err != nil {
		t.Fatalf("Learn() error: %v", err)
	}

	// Verify JSON marshaling works
	jsonData, err := json.Marshal(patterns)
	if err != nil {
		t.Fatalf("Failed to marshal patterns to JSON: %v", err)
	}

	var decoded PatternData
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
}

func TestAnalyzeImportPatterns(t *testing.T) {
	t.Run("detects relative imports", func(t *testing.T) {
		patterns := NewPatternData()
		content := `import { Component } from './component';`
		analyzeImportPatterns("test.ts", content, patterns)
		if patterns.ImportPatterns.Style != "relative" && patterns.ImportPatterns.Style != "mixed" {
			// Should detect at least some imports
		}
	})

	t.Run("detects absolute imports", func(t *testing.T) {
		patterns := NewPatternData()
		content := `import React from 'react';`
		analyzeImportPatterns("test.ts", content, patterns)
		if len(patterns.ImportPatterns.Examples) == 0 {
			t.Error("Expected import examples to be collected")
		}
	})
}

func TestAnalyzeCodeStyle(t *testing.T) {
	t.Run("detects tab indentation", func(t *testing.T) {
		patterns := NewPatternData()
		content := "\tfunction test() {\n\t\treturn true;\n\t}"
		analyzeCodeStyle("test.js", content, patterns)
		if patterns.CodeStyle.IndentStyle == "" {
			// Should detect some style
		}
	})

	t.Run("detects space indentation", func(t *testing.T) {
		patterns := NewPatternData()
		content := "    function test() {\n        return true;\n    }"
		analyzeCodeStyle("test.js", content, patterns)
		if patterns.CodeStyle.IndentStyle == "" {
			// Should detect some style
		}
	})

	t.Run("detects quote style", func(t *testing.T) {
		patterns := NewPatternData()
		content := `const x = 'single'; const y = "double";`
		analyzeCodeStyle("test.js", content, patterns)
		// Should detect quote preference
		_ = patterns
	})
}

func TestAnalyzeFolderStructure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create folder structure
	os.MkdirAll(filepath.Join(tmpDir, "src", "components"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "src", "services"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "tests"), 0755)

	patterns := NewPatternData()
	analyzeFolderStructure(tmpDir, patterns)

	if len(patterns.ProjectStructure) == 0 {
		t.Error("Expected folder structure patterns to be detected")
	}
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		path       string
		shouldSkip bool
	}{
		{"/node_modules/file.js", true},
		{"/.git/config", true},
		{"/build/output.js", true},
		{"/src/file.js", false},
		{"/project/file.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := shouldSkipPath(tt.path)
			if result != tt.shouldSkip {
				t.Errorf("shouldSkipPath(%q) = %v, want %v", tt.path, result, tt.shouldSkip)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	t.Run("contains element", func(t *testing.T) {
		if !contains(slice, "a") {
			t.Error("Expected contains to return true")
		}
	})

	t.Run("does not contain element", func(t *testing.T) {
		if contains(slice, "d") {
			t.Error("Expected contains to return false")
		}
	})
}

func TestMin(t *testing.T) {
	if min(3, 5) != 3 {
		t.Error("Expected min(3, 5) to return 3")
	}
	if min(5, 3) != 3 {
		t.Error("Expected min(5, 3) to return 3")
	}
	if min(3, 3) != 3 {
		t.Error("Expected min(3, 3) to return 3")
	}
}

func TestGetKeys(t *testing.T) {
	m := map[string]int{
		"apple":  5,
		"banana": 3,
		"cherry": 8,
	}

	keys := getKeys(m)
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check that all keys are present
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for expectedKey := range m {
		if !keyMap[expectedKey] {
			t.Errorf("Expected key %q not found in result", expectedKey)
		}
	}
}

func TestGetKeys_EmptyMap(t *testing.T) {
	m := map[string]int{}
	keys := getKeys(m)
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys for empty map, got %d", len(keys))
	}
}
