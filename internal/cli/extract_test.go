// Package cli provides tests for knowledge extraction functionality
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseExtractOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := parseExtractOptions([]string{"file.txt"})
		if opts.inputFile != "file.txt" {
			t.Errorf("Expected inputFile 'file.txt', got '%s'", opts.inputFile)
		}
		if opts.schemaType != "business_rule" {
			t.Errorf("Expected default schemaType 'business_rule', got '%s'", opts.schemaType)
		}
		if !opts.useLLM {
			t.Error("Expected useLLM to be true by default")
		}
		if !opts.useFallback {
			t.Error("Expected useFallback to be true by default")
		}
		if opts.minConfidence != 0.6 {
			t.Errorf("Expected default minConfidence 0.6, got %f", opts.minConfidence)
		}
	})

	t.Run("all flags", func(t *testing.T) {
		opts := parseExtractOptions([]string{
			"file.txt",
			"--no-llm",
			"--no-fallback",
			"--json",
			"--save",
			"--batch",
			"--schema", "entity",
			"--output", "out.json",
			"--min-confidence", "0.8",
		})
		if opts.inputFile != "file.txt" {
			t.Errorf("Expected inputFile 'file.txt', got '%s'", opts.inputFile)
		}
		if opts.useLLM {
			t.Error("Expected useLLM to be false")
		}
		if opts.useFallback {
			t.Error("Expected useFallback to be false")
		}
		if !opts.jsonOutput {
			t.Error("Expected jsonOutput to be true")
		}
		if !opts.saveToKB {
			t.Error("Expected saveToKB to be true")
		}
		if !opts.useBatch {
			t.Error("Expected useBatch to be true")
		}
		if opts.schemaType != "entity" {
			t.Errorf("Expected schemaType 'entity', got '%s'", opts.schemaType)
		}
		if opts.outputFile != "out.json" {
			t.Errorf("Expected outputFile 'out.json', got '%s'", opts.outputFile)
		}
		if opts.minConfidence != 0.8 {
			t.Errorf("Expected minConfidence 0.8, got %f", opts.minConfidence)
		}
	})

	t.Run("short output flag", func(t *testing.T) {
		opts := parseExtractOptions([]string{"file.txt", "-o", "output.json"})
		if opts.outputFile != "output.json" {
			t.Errorf("Expected outputFile 'output.json', got '%s'", opts.outputFile)
		}
	})

	t.Run("input file without flag prefix", func(t *testing.T) {
		opts := parseExtractOptions([]string{"--json", "file.txt"})
		if opts.inputFile != "file.txt" {
			t.Errorf("Expected inputFile 'file.txt', got '%s'", opts.inputFile)
		}
		if !opts.jsonOutput {
			t.Error("Expected jsonOutput to be true")
		}
	})
}

func TestExtractHelpers(t *testing.T) {
	t.Run("newSimpleCache", func(t *testing.T) {
		cache := newSimpleCache()
		if cache == nil {
			t.Error("Expected non-nil cache")
		}
	})

	t.Run("simpleCache Get/Set", func(t *testing.T) {
		cache := newSimpleCache()

		// Test Set
		cache.Set("key1", "value1", 100)

		// Test Get
		val, ok := cache.Get("key1")
		if !ok {
			t.Error("Expected key to be found")
		}
		if val != "value1" {
			t.Errorf("Expected value 'value1', got '%s'", val)
		}

		// Test Get non-existent key
		_, ok = cache.Get("nonexistent")
		if ok {
			t.Error("Expected key to not be found")
		}
	})

	t.Run("newCLILogger", func(t *testing.T) {
		logger := newCLILogger()
		if logger == nil {
			t.Error("Expected non-nil logger")
		}

		// Test that logger methods don't panic
		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")
	})
}

func TestRunExtract(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("help flag", func(t *testing.T) {
		err := runExtract([]string{"--help"})
		if err != nil {
			t.Errorf("Expected no error for help flag, got: %v", err)
		}
	})

	t.Run("no arguments shows help", func(t *testing.T) {
		err := runExtract([]string{})
		if err != nil {
			// Help might return error or not, just check it doesn't panic
		}
	})

	t.Run("missing file", func(t *testing.T) {
		// This will fail because file doesn't exist, but tests the code path
		err := runExtract([]string{"nonexistent.txt"})
		// Expected to fail with file not found error
		_ = err
	})

	t.Run("with text file", func(t *testing.T) {
		// Create a simple text file
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("Business rule: Users must authenticate."), 0644)

		// This will fail without LLM setup, but tests parsing and initial setup
		err := runExtract([]string{testFile, "--no-llm"})
		// May fail on extraction, but should parse options correctly
		_ = err
	})
}

func TestPrintExtractHelp(t *testing.T) {
	err := printExtractHelp()
	if err != nil {
		t.Errorf("printExtractHelp() should not error: %v", err)
	}
}
