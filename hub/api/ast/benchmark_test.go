// Package ast benchmark tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// generateLargeCode generates a large code string for benchmarking
func generateLargeCode(lang string, lines int) string {
	linesSlice := make([]string, lines)
	switch lang {
	case "go":
		for i := 0; i < lines; i++ {
			linesSlice[i] = "func test" + strings.Repeat(string(rune('A'+i%26)), 1) + "() { fmt.Println(\"test\"); }"
		}
	case "javascript":
		for i := 0; i < lines; i++ {
			linesSlice[i] = "function test" + strings.Repeat(string(rune('A'+i%26)), 1) + "() { console.log('test'); }"
		}
	case "python":
		for i := 0; i < lines; i++ {
			linesSlice[i] = "def test" + strings.Repeat(string(rune('A'+i%26)), 1) + "(): print('test')"
		}
	default:
		for i := 0; i < lines; i++ {
			linesSlice[i] = "// Line " + string(rune('A'+i%26))
		}
	}
	return strings.Join(linesSlice, "\n")
}

// BenchmarkAnalyzeAST_SmallFile benchmarks AST analysis on small files
func BenchmarkAnalyzeAST_SmallFile(b *testing.B) {
	code := `
package main

func hello() {
    fmt.Println("hello")
}

func world() {
    fmt.Println("world")
}
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := AnalyzeAST(code, "go", []string{})
		if err != nil {
			b.Fatalf("AnalyzeAST failed: %v", err)
		}
	}
}

// BenchmarkAnalyzeAST_MediumFile benchmarks AST analysis on medium files
func BenchmarkAnalyzeAST_MediumFile(b *testing.B) {
	code := generateLargeCode("go", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := AnalyzeAST(code, "go", []string{})
		if err != nil {
			b.Fatalf("AnalyzeAST failed: %v", err)
		}
	}
}

// BenchmarkAnalyzeAST_LargeFile benchmarks AST analysis on large files
func BenchmarkAnalyzeAST_LargeFile(b *testing.B) {
	code := generateLargeCode("go", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := AnalyzeAST(code, "go", []string{})
		if err != nil {
			b.Fatalf("AnalyzeAST failed: %v", err)
		}
	}
}

// BenchmarkAnalyzeAST_VeryLargeFile benchmarks AST analysis on very large files
func BenchmarkAnalyzeAST_VeryLargeFile(b *testing.B) {
	code := generateLargeCode("go", 5000)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := AnalyzeAST(code, "go", []string{})
		if err != nil {
			b.Fatalf("AnalyzeAST failed: %v", err)
		}
	}
}

// BenchmarkParserCache benchmarks parser cache performance
func BenchmarkParserCache(b *testing.B) {
	code := "func test() {}"

	// First call to populate cache
	_, _, _ = AnalyzeAST(code, "go", []string{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := AnalyzeAST(code, "go", []string{})
		if err != nil {
			b.Fatalf("AnalyzeAST failed: %v", err)
		}
	}
}

// BenchmarkAnalyzeAST_AllLanguages benchmarks AST analysis across all supported languages
func BenchmarkAnalyzeAST_AllLanguages(b *testing.B) {
	langs := []string{"go", "javascript", "typescript", "python"}

	for _, lang := range langs {
		code := generateLargeCode(lang, 50)
		b.Run(lang, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _, err := AnalyzeAST(code, lang, []string{})
				if err != nil {
					b.Fatalf("AnalyzeAST failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkAnalyzeAST_WithAllAnalyses benchmarks AST analysis with all analyses enabled
func BenchmarkAnalyzeAST_WithAllAnalyses(b *testing.B) {
	code := generateLargeCode("go", 100)
	analyses := []string{"duplicates", "unused", "unreachable", "empty_catch", "missing_await", "brace_mismatch"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := AnalyzeAST(code, "go", analyses)
		if err != nil {
			b.Fatalf("AnalyzeAST failed: %v", err)
		}
	}
}

// BenchmarkSafeSlice benchmarks the safeSlice function
func BenchmarkSafeSlice(b *testing.B) {
	code := strings.Repeat("hello world ", 1000)
	start := uint32(0)
	end := uint32(len(code) / 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = safeSlice(code, start, end)
	}
}

// BenchmarkGetLineColumn benchmarks the getLineColumn function
func BenchmarkGetLineColumn(b *testing.B) {
	code := strings.Repeat("line\n", 1000)
	offset := len(code) / 2

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = getLineColumn(code, offset)
	}
}

// BenchmarkValidationLargeCodebase benchmarks validation on large codebase
func BenchmarkValidationLargeCodebase(b *testing.B) {
	code := generateLargeCode("go", 100)
	projectRoot := b.TempDir()
	filePath := "test.go"

	// Create a simple project structure
	os.WriteFile(filepath.Join(projectRoot, filePath), []byte(code), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := AnalyzeASTWithValidation(code, "go", filePath, projectRoot, []string{"orphaned", "unused"})
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkSearchWithTimeout benchmarks search with timeout
func BenchmarkSearchWithTimeout(b *testing.B) {
	projectRoot := b.TempDir()
	pattern := `\btestFunc\s*\(`

	// Create test files
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(projectRoot, fmt.Sprintf("file%d.go", i)),
			[]byte("func testFunc() {}"), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SearchWithTimeout(pattern, projectRoot, 5*time.Second)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}

// BenchmarkConcurrentValidation benchmarks concurrent validation
func BenchmarkConcurrentValidation(b *testing.B) {
	code := generateLargeCode("go", 50)
	projectRoot := b.TempDir()
	filePath := "test.go"

	// Create test file
	os.WriteFile(filepath.Join(projectRoot, filePath), []byte(code), 0644)

	// Generate findings
	findings, _, err := AnalyzeAST(code, "go", []string{"orphaned", "unused"})
	if err != nil {
		b.Fatalf("Analysis failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidateFindingsConcurrent(findings, filePath, projectRoot, "go", 4)
		if err != nil {
			b.Fatalf("Concurrent validation failed: %v", err)
		}
	}
}
