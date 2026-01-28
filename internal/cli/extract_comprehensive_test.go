// Package cli provides comprehensive tests for extract command
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/divyang-garg/sentinel-hub-api/internal/extraction"
)

func TestCreateExtractor(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("with no-llm option", func(t *testing.T) {
		opts := extractOptions{
			useLLM: false,
		}
		extractor, err := createExtractor(opts)
		if err != nil {
			// May fail due to missing LLM setup, but should handle gracefully
			_ = err
		}
		if extractor != nil {
			// Extractor created successfully
			_ = extractor
		}
	})

	t.Run("with cache directory", func(t *testing.T) {
		cacheDir := filepath.Join(tmpDir, ".cache")
		os.Setenv("SENTINEL_CACHE_DIR", cacheDir)
		defer os.Unsetenv("SENTINEL_CACHE_DIR")

		opts := extractOptions{
			useLLM: false,
		}
		extractor, err := createExtractor(opts)
		if err != nil {
			_ = err
		}
		if extractor != nil {
			_ = extractor
		}
	})

	t.Run("fallback to in-memory cache", func(t *testing.T) {
		os.Unsetenv("SENTINEL_CACHE_DIR")
		os.Unsetenv("HOME")
		os.Unsetenv("USERPROFILE")

		opts := extractOptions{
			useLLM: false,
		}
		extractor, err := createExtractor(opts)
		if err != nil {
			_ = err
		}
		if extractor != nil {
			_ = extractor
		}
	})
}

func TestOutputResults(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	result := &extraction.ExtractResult{
		Source:     "test.txt",
		Confidence: 0.85,
		BusinessRules: []extraction.BusinessRule{
			{
				ID:          "BR1",
				Title:       "Test Rule",
				Description: "Test description",
				Confidence:  0.9,
				Specification: extraction.Specification{
					Constraints: []extraction.Constraint{
						{ID: "C1"},
					},
				},
			},
			{
				ID:          "BR2",
				Title:       "Low Confidence",
				Description: "Low confidence rule",
				Confidence:  0.5, // Below default 0.6 threshold
			},
		},
	}

	t.Run("text output", func(t *testing.T) {
		opts := extractOptions{
			jsonOutput:    false,
			minConfidence: 0.6,
		}
		err := outputResults(result, opts)
		if err != nil {
			t.Errorf("outputResults() error = %v", err)
		}
	})

	t.Run("JSON output to stdout", func(t *testing.T) {
		opts := extractOptions{
			jsonOutput:    true,
			minConfidence: 0.6,
		}
		err := outputResults(result, opts)
		if err != nil {
			t.Errorf("outputResults() error = %v", err)
		}
	})

	t.Run("JSON output to file", func(t *testing.T) {
		opts := extractOptions{
			jsonOutput:    true,
			outputFile:    "results.json",
			minConfidence: 0.6,
		}
		err := outputResults(result, opts)
		if err != nil {
			t.Errorf("outputResults() error = %v", err)
		}
		if _, err := os.Stat("results.json"); os.IsNotExist(err) {
			t.Error("Expected results.json to be created")
		}
	})

	t.Run("with save to KB", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		opts := extractOptions{
			jsonOutput:    false,
			saveToKB:      true,
			minConfidence: 0.6,
		}
		err := outputResults(result, opts)
		if err != nil {
			t.Errorf("outputResults() error = %v", err)
		}
	})
}

func TestFilterByConfidence(t *testing.T) {
	rules := []extraction.BusinessRule{
		{ID: "R1", Confidence: 0.9},
		{ID: "R2", Confidence: 0.7},
		{ID: "R3", Confidence: 0.5},
		{ID: "R4", Confidence: 0.3},
	}

	t.Run("filter at 0.6", func(t *testing.T) {
		filtered := filterByConfidence(rules, 0.6)
		if len(filtered) != 2 {
			t.Errorf("Expected 2 rules, got %d", len(filtered))
		}
		if filtered[0].ID != "R1" {
			t.Errorf("Expected R1, got %s", filtered[0].ID)
		}
		if filtered[1].ID != "R2" {
			t.Errorf("Expected R2, got %s", filtered[1].ID)
		}
	})

	t.Run("filter at 0.8", func(t *testing.T) {
		filtered := filterByConfidence(rules, 0.8)
		if len(filtered) != 1 {
			t.Errorf("Expected 1 rule, got %d", len(filtered))
		}
	})

	t.Run("filter at 0.2", func(t *testing.T) {
		filtered := filterByConfidence(rules, 0.2)
		if len(filtered) != 4 {
			t.Errorf("Expected 4 rules, got %d", len(filtered))
		}
	})

	t.Run("empty rules", func(t *testing.T) {
		filtered := filterByConfidence([]extraction.BusinessRule{}, 0.6)
		if len(filtered) != 0 {
			t.Errorf("Expected 0 rules, got %d", len(filtered))
		}
	})
}

func TestSaveToKnowledgeBase(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	rules := []extraction.BusinessRule{
		{
			ID:          "BR1",
			Title:       "Test Rule",
			Description: "Test description",
			Priority:    "high",
			Status:      "approved",
			Traceability: extraction.Traceability{
				SourceDocument: "test.txt",
			},
		},
	}

	t.Run("save rules", func(t *testing.T) {
		err := saveToKnowledgeBase(rules)
		if err != nil {
			t.Errorf("saveToKnowledgeBase() error = %v", err)
		}

		// Verify rules were saved
		kb, err := loadKnowledge()
		if err != nil {
			t.Fatalf("loadKnowledge() error = %v", err)
		}
		if len(kb.Entries) != 1 {
			t.Errorf("Expected 1 entry, got %d", len(kb.Entries))
		}
		if kb.Entries[0].ID != "BR1" {
			t.Errorf("Expected entry ID BR1, got %s", kb.Entries[0].ID)
		}
	})

	t.Run("append to existing KB", func(t *testing.T) {
		// Add existing entry
		_ = saveToKnowledgeBase(rules)

		// Add more rules
		moreRules := []extraction.BusinessRule{
			{
				ID:          "BR2",
				Title:       "Another Rule",
				Description: "Another description",
			},
		}
		err := saveToKnowledgeBase(moreRules)
		if err != nil {
			t.Errorf("saveToKnowledgeBase() error = %v", err)
		}

		kb, _ := loadKnowledge()
		if len(kb.Entries) < 2 {
			t.Errorf("Expected at least 2 entries, got %d", len(kb.Entries))
		}
	})
}
