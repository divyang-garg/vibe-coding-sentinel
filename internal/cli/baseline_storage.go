// Package cli provides baseline storage operations
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/divyang-garg/sentinel-hub-api/internal/config"
)

// loadBaseline loads the baseline from disk
func loadBaseline() (*Baseline, error) {
	baselinePath := getBaselinePath()

	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return &Baseline{Version: "1.0", Entries: []BaselineEntry{}}, err
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, err
	}

	return &baseline, nil
}

// saveBaseline saves the baseline to disk
func saveBaseline(baseline *Baseline) error {
	baselinePath := getBaselinePath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(baselinePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}

	return config.WriteFile(baselinePath, string(data))
}

// getBaselinePath returns the path to the baseline file
func getBaselinePath() string {
	return ".sentinel/baseline.json"
}

// exportBaseline exports baseline to a file
func exportBaseline(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sentinel baseline export <file>")
	}

	outputFile := args[0]

	baseline, err := loadBaseline()
	if err != nil {
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Baseline exported to: %s\n", outputFile)
	return nil
}

// importBaseline imports baseline from a file
func importBaseline(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sentinel baseline import <file>")
	}

	inputFile := args[0]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return fmt.Errorf("failed to parse baseline: %w", err)
	}

	if err := saveBaseline(&baseline); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	fmt.Printf("✅ Imported %d baseline entries\n", len(baseline.Entries))
	return nil
}
