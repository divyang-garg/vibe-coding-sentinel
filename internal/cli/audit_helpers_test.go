// Package cli provides tests for audit helper functions
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/divyang-garg/sentinel-hub-api/internal/scanner"
)

func TestSaveResults(t *testing.T) {
	tmpDir := t.TempDir()

	result := &scanner.Result{
		Success:  true,
		Findings: []scanner.Finding{},
		Summary:  make(map[string]int),
	}

	t.Run("save as json", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "results.json")
		err := saveResults(result, filename, "json")
		if err != nil {
			t.Fatalf("Failed to save JSON: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Error("JSON file was not created")
		}
	})

	t.Run("save as html", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "results.html")
		err := saveResults(result, filename, "html")
		if err != nil {
			t.Fatalf("Failed to save HTML: %v", err)
		}

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Error("HTML file was not created")
		}
	})

	t.Run("save as markdown", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "results.md")
		err := saveResults(result, filename, "md")
		if err != nil {
			t.Fatalf("Failed to save Markdown: %v", err)
		}

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Error("Markdown file was not created")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "results.invalid")
		err := saveResults(result, filename, "invalid")
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})
}

func TestGetHubURL(t *testing.T) {
	t.Run("from environment", func(t *testing.T) {
		os.Setenv("SENTINEL_HUB_URL", "http://test-hub:9000")
		defer os.Unsetenv("SENTINEL_HUB_URL")

		url := getHubURL()
		if url != "http://test-hub:9000" {
			t.Errorf("Expected http://test-hub:9000, got %s", url)
		}
	})

	t.Run("default", func(t *testing.T) {
		os.Unsetenv("SENTINEL_HUB_URL")
		url := getHubURL()
		if url != "http://localhost:8080" {
			t.Errorf("Expected default URL, got %s", url)
		}
	})
}

func TestGetAPIKey(t *testing.T) {
	t.Run("from environment", func(t *testing.T) {
		os.Setenv("SENTINEL_HUB_API_KEY", "test-key-12345")
		defer os.Unsetenv("SENTINEL_HUB_API_KEY")

		key := getAPIKey()
		if key != "test-key-12345" {
			t.Errorf("Expected test-key-12345, got %s", key)
		}
	})

	t.Run("not set", func(t *testing.T) {
		os.Unsetenv("SENTINEL_HUB_API_KEY")
		key := getAPIKey()
		if key != "" {
			t.Errorf("Expected empty key, got %s", key)
		}
	})
}

func TestCollectCodeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte("console.log('test')"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# README"), 0644)

	files, err := collectCodeFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to collect files: %v", err)
	}

	// Should find .go and .js, but not .md
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filepath string
		expected string
	}{
		{"test.go", "go"},
		{"test.js", "javascript"},
		{"test.ts", "typescript"},
		{"test.py", "python"},
		{"test.java", "java"},
		{"test.rb", "ruby"},
		{"test.php", "php"},
		{"test.txt", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			lang := detectLanguage(tt.filepath)
			if lang != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, lang)
			}
		})
	}
}

func TestIsCodeFile(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".go", true},
		{".js", true},
		{".ts", true},
		{".py", true},
		{".md", false},
		{".txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := isCodeFile(tt.ext)
			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.ext, result)
			}
		})
	}
}

func TestConvertHubSeverity(t *testing.T) {
	tests := []struct {
		hubSeverity string
		expected    scanner.Severity
	}{
		{"critical", scanner.SeverityCritical},
		{"high", scanner.SeverityHigh},
		{"medium", scanner.SeverityMedium},
		{"low", scanner.SeverityLow},
		{"unknown", scanner.SeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.hubSeverity, func(t *testing.T) {
			result := convertHubSeverity(tt.hubSeverity)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMergeHubFindings(t *testing.T) {
	result := &scanner.Result{
		Success:  true,
		Findings: []scanner.Finding{},
		Summary:  make(map[string]int),
	}

	hubFindings := []scanner.Finding{
		{
			Type:     "ast_issue",
			Severity: scanner.SeverityHigh,
			File:     "test.go",
			Line:     10,
			Message:  "Test finding",
		},
	}

	merged := mergeHubFindings(result, hubFindings)

	if len(merged.Findings) != 1 {
		t.Errorf("Expected 1 finding, got %d", len(merged.Findings))
	}

	if merged.Summary["ast_issue"] != 1 {
		t.Error("Summary not updated correctly")
	}
}
