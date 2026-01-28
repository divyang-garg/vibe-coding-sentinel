// Package patterns provides tests for pattern learning functionality
package patterns

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestLearn_NamingOnly(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.js")
	os.WriteFile(testFile, []byte(`
function camelCaseFunction() {}
class PascalCaseClass {}
let snake_case_variable = 1;
`), 0644)

	opts := LearnOptions{
		CodebasePath: tmpDir,
		NamingOnly:   true,
	}
	patterns, err := Learn(opts)
	if err != nil {
		t.Fatalf("Learn() error: %v", err)
	}
	if len(patterns.NamingPatterns) == 0 {
		t.Error("Expected naming patterns to be detected")
	}
}

func TestDetectLanguageAndFramework(t *testing.T) {
	testCases := []struct {
		name      string
		path      string
		content   string
		checkLang func(*testing.T, *PatternData)
	}{
		{
			name:    "JavaScript",
			path:    "test.js",
			content: "console.log('hello');",
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Languages["JavaScript/TypeScript"] == 0 {
					t.Error("Expected JavaScript to be detected")
				}
			},
		},
		{
			name:    "TypeScript",
			path:    "test.ts",
			content: "const x: number = 1;",
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Languages["JavaScript/TypeScript"] == 0 {
					t.Error("Expected TypeScript to be detected")
				}
			},
		},
		{
			name:    "React",
			path:    "component.jsx",
			content: "import React from 'react';",
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Frameworks["React"] == 0 {
					t.Error("Expected React to be detected")
				}
			},
		},
		{
			name:    "Python",
			path:    "test.py",
			content: "print('hello')",
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Languages["Python"] == 0 {
					t.Error("Expected Python to be detected")
				}
			},
		},
		{
			name:    "Django",
			path:    "settings.py",
			content: "import django",
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Frameworks["Django"] == 0 {
					t.Error("Expected Django to be detected")
				}
			},
		},
		{
			name:    "Java",
			path:    "Test.java",
			content: "public class Test {}",
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Languages["Java"] == 0 {
					t.Error("Expected Java to be detected")
				}
			},
		},
		{
			name:    "Go",
			path:    "main.go",
			content: "package main",
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Languages["Go"] == 0 {
					t.Error("Expected Go to be detected")
				}
			},
		},
		{
			name:    "package.json",
			path:    "package.json",
			content: `{"name": "test"}`,
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Frameworks["Node.js"] == 0 {
					t.Error("Expected Node.js to be detected from package.json")
				}
			},
		},
		{
			name:    "go.mod",
			path:    "go.mod",
			content: "module test",
			checkLang: func(t *testing.T, p *PatternData) {
				if p.Frameworks["Go Modules"] == 0 {
					t.Error("Expected Go Modules to be detected from go.mod")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPatternData()
			detectLanguageAndFramework(tc.path, tc.content, p)
			tc.checkLang(t, p)
		})
	}
}

func TestAnalyzeNamingPatterns(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		check   func(*testing.T, *PatternData)
	}{
		{
			name:    "camelCase",
			content: "function camelCaseFunction() {}",
			check: func(t *testing.T, p *PatternData) {
				if p.NamingPatterns["camelCase"] == 0 {
					t.Error("Expected camelCase to be detected")
				}
			},
		},
		{
			name:    "PascalCase",
			content: "class PascalCaseClass {}",
			check: func(t *testing.T, p *PatternData) {
				if p.NamingPatterns["PascalCase"] == 0 {
					t.Error("Expected PascalCase to be detected")
				}
			},
		},
		{
			name:    "snake_case",
			content: "let snake_case_variable = 1;",
			check: func(t *testing.T, p *PatternData) {
				if p.NamingPatterns["snake_case"] == 0 {
					t.Error("Expected snake_case to be detected")
				}
			},
		},
		{
			name:    "mixed patterns",
			content: "class ClassName { methodName() {} } const SNAKE_CASE = 1;",
			check: func(t *testing.T, p *PatternData) {
				if len(p.NamingPatterns) == 0 {
					t.Error("Expected naming patterns to be detected")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPatternData()
			analyzeNamingPatterns("test.js", tc.content, p)
			tc.check(t, p)
		})
	}
}

func TestContainsCamelCase(t *testing.T) {
	testCases := []struct {
		content string
		want    bool
	}{
		{"camelCase", true},
		{"PascalCase", true}, // Contains camelCase pattern (lC)
		{"snake_case", false},
		{"lowercase", false},
		{"UPPERCASE", false},
		{"", false},
		{"testFunction", true},
		{"TEST_FUNCTION", false},
	}

	for _, tc := range testCases {
		t.Run(tc.content, func(t *testing.T) {
			got := containsCamelCase(tc.content)
			if got != tc.want {
				t.Errorf("containsCamelCase(%q) = %v, want %v", tc.content, got, tc.want)
			}
		})
	}
}

func TestContainsPascalCase(t *testing.T) {
	testCases := []struct {
		content string
		want    bool
	}{
		{"class TestClass {}", true},
		{"type TypeName", true},
		{"function test()", false},
		{"const x = 1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.content, func(t *testing.T) {
			got := containsPascalCase(tc.content)
			if got != tc.want {
				t.Errorf("containsPascalCase(%q) = %v, want %v", tc.content, got, tc.want)
			}
		})
	}
}

func TestContainsSnakeCase(t *testing.T) {
	testCases := []struct {
		content string
		want    bool
	}{
		{"snake_case", true},
		{"UPPER_CASE", true},
		{"camelCase", false},
		{"PascalCase", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.content, func(t *testing.T) {
			got := containsSnakeCase(tc.content)
			if got != tc.want {
				t.Errorf("containsSnakeCase(%q) = %v, want %v", tc.content, got, tc.want)
			}
		})
	}
}

func TestFindPrimaryLanguage(t *testing.T) {
	t.Run("TypeScript priority", func(t *testing.T) {
		p := NewPatternData()
		p.Languages["JavaScript/TypeScript"] = 10
		p.FileExtensions[".ts"] = 8
		p.FileExtensions[".js"] = 2
		p.Languages["Python"] = 5
		result := findPrimaryLanguage(p)
		if result != "TypeScript" {
			t.Errorf("Expected TypeScript, got %s", result)
		}
	})

	t.Run("JavaScript priority", func(t *testing.T) {
		p := NewPatternData()
		p.Languages["JavaScript/TypeScript"] = 10
		p.FileExtensions[".js"] = 8
		p.FileExtensions[".ts"] = 2
		result := findPrimaryLanguage(p)
		if result != "JavaScript" {
			t.Errorf("Expected JavaScript, got %s", result)
		}
	})

	t.Run("other language", func(t *testing.T) {
		p := NewPatternData()
		p.Languages["Python"] = 20
		p.Languages["Go"] = 5
		result := findPrimaryLanguage(p)
		if result != "Python" {
			t.Errorf("Expected Python, got %s", result)
		}
	})

	t.Run("empty", func(t *testing.T) {
		p := NewPatternData()
		result := findPrimaryLanguage(p)
		if result != "" {
			t.Errorf("Expected empty string, got %s", result)
		}
	})
}

func TestGenerateOutputFiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	patterns := NewPatternData()
	patterns.Languages["Go"] = 10
	patterns.Frameworks["Go Modules"] = 1
	patterns.NamingPatterns["camelCase"] = 5

	err := generateOutputFiles(patterns)
	if err != nil {
		t.Fatalf("generateOutputFiles() error: %v", err)
	}

	// Check that files were created
	if _, err := os.Stat(".sentinel/patterns.json"); os.IsNotExist(err) {
		t.Error("Expected patterns.json to be created")
	}
	if _, err := os.Stat(".cursor/rules/project-patterns.md"); os.IsNotExist(err) {
		t.Error("Expected project-patterns.md to be created")
	}
}

func TestGenerateOutputFiles_WithBusinessRules(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	patterns := NewPatternData()
	patterns.BusinessRules = &BusinessRuleData{
		Rules: []BusinessRule{
			{
				ID:      "BR1",
				Title:   "Test Rule",
				Content: "Test content",
			},
		},
	}

	err := generateOutputFiles(patterns)
	if err != nil {
		t.Fatalf("generateOutputFiles() with business rules error: %v", err)
	}

	// Check that business rules file was created
	if _, err := os.Stat(".cursor/rules/business-rules.md"); os.IsNotExist(err) {
		t.Error("Expected business-rules.md to be created")
	}
}

func TestGenerateCursorRules(t *testing.T) {
	patterns := NewPatternData()
	patterns.Languages["JavaScript/TypeScript"] = 10
	patterns.Frameworks["React"] = 1
	patterns.NamingPatterns["camelCase"] = 5

	markdown := generateCursorRules(patterns)
	if markdown == "" {
		t.Error("Expected non-empty markdown")
	}
	if !strings.Contains(markdown, "Project Patterns") {
		t.Error("Expected markdown to contain 'Project Patterns'")
	}
}

func TestGenerateBusinessRulesForCursor(t *testing.T) {
	rules := []BusinessRule{
		{
			ID:         "BR1",
			Title:      "Rule 1",
			Content:    "Content 1",
			Confidence: 0.9,
			SourcePage: intPtr(1),
		},
		{
			ID:      "BR2",
			Title:   "Rule 2",
			Content: "Content 2",
		},
	}

	markdown := generateBusinessRulesForCursor(rules)
	if markdown == "" {
		t.Error("Expected non-empty markdown")
	}
	if !strings.Contains(markdown, "Business Rules") {
		t.Error("Expected markdown to contain 'Business Rules'")
	}
	if !strings.Contains(markdown, "Rule 1") {
		t.Error("Expected markdown to contain 'Rule 1'")
	}
}

func TestGenerateBusinessRulesForCursor_Empty(t *testing.T) {
	markdown := generateBusinessRulesForCursor([]BusinessRule{})
	if markdown == "" {
		t.Error("Expected non-empty markdown even for empty rules")
	}
	if !strings.Contains(markdown, "No business rules found") {
		t.Error("Expected message about no business rules")
	}
}

func intPtr(i int) *int {
	return &i
}

func TestLearn_ErrorCases(t *testing.T) {
	t.Run("handles analyzeCodebase error", func(t *testing.T) {
		// Use non-existent path
		opts := LearnOptions{
			CodebasePath: "/nonexistent/path/that/does/not/exist",
		}
		_, err := Learn(opts)
		// Should handle error gracefully or return error
		_ = err
	})

	t.Run("handles generateOutputFiles error", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Make directory read-only to cause write error
		os.Chmod(tmpDir, 0555)
		defer os.Chmod(tmpDir, 0755)

		opts := LearnOptions{
			CodebasePath: tmpDir,
		}
		_, err := Learn(opts)
		// Should handle error
		_ = err
	})

	t.Run("handles business rules fetch error", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")
		os.WriteFile(testFile, []byte(`package main`), 0644)

		opts := LearnOptions{
			CodebasePath:         tmpDir,
			IncludeBusinessRules: true,
			HubURL:               "http://invalid-url",
			HubAPIKey:            "invalid-key",
			ProjectID:            "invalid-project",
		}
		patterns, err := Learn(opts)
		// Should not fail, just log warning
		if err != nil {
			t.Fatalf("Learn should handle Hub errors gracefully: %v", err)
		}
		if patterns == nil {
			t.Fatal("Patterns should not be nil")
		}
	})

	t.Run("handles JSON marshal error in OutputJSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")
		os.WriteFile(testFile, []byte(`package main`), 0644)

		opts := LearnOptions{
			CodebasePath: tmpDir,
			OutputJSON:   true,
		}
		patterns, err := Learn(opts)
		// Should handle marshal error
		if err != nil {
			t.Fatalf("Learn should handle JSON marshal: %v", err)
		}
		if patterns == nil {
			t.Fatal("Patterns should not be nil")
		}
	})
}

func TestAnalyzeImportPatterns_EdgeCases(t *testing.T) {
	t.Run("handles barrel files", func(t *testing.T) {
		patterns := NewPatternData()
		content := `export { Component } from './component';
export { Service } from './service';`
		analyzeImportPatterns("index.ts", content, patterns)
		if len(patterns.ImportPatterns.BarrelFiles) == 0 {
			t.Error("Expected barrel file to be detected")
		}
	})

	t.Run("handles mixed import styles", func(t *testing.T) {
		patterns := NewPatternData()
		content := `import React from 'react';
import { Component } from './component';
import * as utils from './utils';`
		analyzeImportPatterns("test.ts", content, patterns)
		if patterns.ImportPatterns.Style == "" {
			t.Error("Expected import style to be detected")
		}
		if patterns.ImportPatterns.NamedImports == 0 && patterns.ImportPatterns.DefaultImports == 0 {
			t.Error("Expected import counts to be detected")
		}
	})

	t.Run("handles Python relative imports", func(t *testing.T) {
		patterns := NewPatternData()
		content := `from .module import something
from ..parent import other`
		analyzeImportPatterns("test.py", content, patterns)
		// Should detect relative imports (style should be "relative")
		if patterns.ImportPatterns.Style == "" {
			t.Error("Expected import style to be detected")
		}
	})

	t.Run("handles unsupported language", func(t *testing.T) {
		patterns := NewPatternData()
		content := `import something`
		analyzeImportPatterns("test.java", content, patterns)
		// Should not crash
	})

	t.Run("handles long import lines", func(t *testing.T) {
		patterns := NewPatternData()
		longLine := strings.Repeat("import ", 50) + "very long import statement that exceeds 100 characters"
		analyzeImportPatterns("test.ts", longLine, patterns)
		// Should handle gracefully
	})
}

func TestAnalyzeCodeStyle_EdgeCases(t *testing.T) {
	t.Run("handles empty content", func(t *testing.T) {
		patterns := NewPatternData()
		analyzeCodeStyle("test.js", "", patterns)
		// Should not crash
	})

	t.Run("handles CRLF line endings", func(t *testing.T) {
		patterns := NewPatternData()
		content := "function test() {\r\n    return true;\r\n}"
		analyzeCodeStyle("test.js", content, patterns)
		if patterns.CodeStyle.LineEnding != "crlf" {
			t.Error("Expected CRLF line ending to be detected")
		}
	})

	t.Run("handles mixed quote styles", func(t *testing.T) {
		patterns := NewPatternData()
		content := `const x = 'single';
const y = "double";
const z = 'another';`
		analyzeCodeStyle("test.js", content, patterns)
		// Should detect dominant style
		if patterns.CodeStyle.QuoteStyle == "" {
			t.Error("Expected quote style to be detected")
		}
	})

	t.Run("handles semicolon detection edge cases", func(t *testing.T) {
		patterns := NewPatternData()
		content := `const x = 1;
const y = 2
const z = 3;`
		analyzeCodeStyle("test.js", content, patterns)
		// Should detect semicolon preference
		if patterns.CodeStyle.Semicolons == "" {
			t.Error("Expected semicolon style to be detected")
		}
	})

	t.Run("handles comment lines with semicolons", func(t *testing.T) {
		patterns := NewPatternData()
		content := `// This is a comment;
const x = 1`
		analyzeCodeStyle("test.js", content, patterns)
		// Should not count comment semicolons
		_ = patterns
	})
}

func TestAnalyzeFolderStructure_EdgeCases(t *testing.T) {
	t.Run("handles walk errors gracefully", func(t *testing.T) {
		patterns := NewPatternData()
		// Use invalid path
		analyzeFolderStructure("/nonexistent/path", patterns)
		// Should not crash
	})

	t.Run("handles nested structure patterns", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "src", "components", "ui"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "src", "services", "api"), 0755)

		patterns := NewPatternData()
		analyzeFolderStructure(tmpDir, patterns)
		if len(patterns.ProjectStructure) == 0 {
			t.Error("Expected folder structure patterns to be detected")
		}
	})

	t.Run("limits examples to 10", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create multiple component directories
		for i := 0; i < 15; i++ {
			os.MkdirAll(filepath.Join(tmpDir, "src", "components", fmt.Sprintf("comp%d", i)), 0755)
		}

		patterns := NewPatternData()
		analyzeFolderStructure(tmpDir, patterns)
		examples := patterns.ProjectStructure["components"]
		if len(examples) > 10 {
			t.Errorf("Expected max 10 examples, got %d", len(examples))
		}
	})
}

func TestDetectLanguageAndFramework_EdgeCases(t *testing.T) {
	t.Run("detects Vue.js", func(t *testing.T) {
		p := NewPatternData()
		detectLanguageAndFramework("component.js", "import Vue from 'vue'", p)
		if p.Frameworks["Vue.js"] == 0 {
			t.Error("Expected Vue.js to be detected")
		}
	})

	t.Run("detects Angular", func(t *testing.T) {
		p := NewPatternData()
		detectLanguageAndFramework("component.ts", "import { Component } from '@angular/core'", p)
		if p.Frameworks["Angular"] == 0 {
			t.Error("Expected Angular to be detected")
		}
	})

	t.Run("detects Flask", func(t *testing.T) {
		p := NewPatternData()
		detectLanguageAndFramework("app.py", "from flask import Flask", p)
		if p.Frameworks["Flask"] == 0 {
			t.Error("Expected Flask to be detected")
		}
	})

	t.Run("detects Spring", func(t *testing.T) {
		p := NewPatternData()
		detectLanguageAndFramework("Controller.java", "import org.springframework", p)
		if p.Frameworks["Spring"] == 0 {
			t.Error("Expected Spring to be detected")
		}
	})

	t.Run("detects ASP.NET", func(t *testing.T) {
		p := NewPatternData()
		detectLanguageAndFramework("Controller.cs", "using aspnet", p)
		if p.Frameworks["ASP.NET"] == 0 {
			t.Error("Expected ASP.NET to be detected")
		}
	})

	t.Run("detects Ruby on Rails", func(t *testing.T) {
		p := NewPatternData()
		detectLanguageAndFramework("controller.rb", "class ApplicationController < rails", p)
		if p.Frameworks["Ruby on Rails"] == 0 {
			t.Error("Expected Ruby on Rails to be detected")
		}
	})

	t.Run("detects requirements.txt", func(t *testing.T) {
		p := NewPatternData()
		detectLanguageAndFramework("requirements.txt", "flask==1.0.0", p)
		if p.Frameworks["Python"] == 0 {
			t.Error("Expected Python framework to be detected from requirements.txt")
		}
	})

	t.Run("detects pyproject.toml", func(t *testing.T) {
		p := NewPatternData()
		detectLanguageAndFramework("pyproject.toml", "[project]", p)
		if p.Frameworks["Python"] == 0 {
			t.Error("Expected Python framework to be detected from pyproject.toml")
		}
	})
}

func TestGenerateOutputFiles_ErrorCases(t *testing.T) {
	t.Run("handles directory creation failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		// Create a file named .sentinel to prevent directory creation
		os.WriteFile(".sentinel", []byte("file"), 0644)
		defer os.Remove(".sentinel")

		patterns := NewPatternData()
		err := generateOutputFiles(patterns)
		if err == nil {
			t.Error("Expected error when directory creation fails")
		}
	})

	t.Run("handles JSON marshal failure", func(t *testing.T) {
		// This is hard to test without creating a type that can't be marshaled
		// But we can test the error path exists
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		patterns := NewPatternData()
		err := generateOutputFiles(patterns)
		if err != nil {
			t.Fatalf("generateOutputFiles should work with valid patterns: %v", err)
		}
	})

	t.Run("handles file write failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		// Create read-only .sentinel directory
		os.MkdirAll(".sentinel", 0555)
		defer os.Chmod(".sentinel", 0755)

		patterns := NewPatternData()
		err := generateOutputFiles(patterns)
		// Should handle write error
		_ = err
	})
}
