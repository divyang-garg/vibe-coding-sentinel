// Package patterns provides pattern learning functionality
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package patterns

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LearnOptions configures pattern learning
type LearnOptions struct {
	NamingOnly           bool
	ImportsOnly          bool
	StructureOnly        bool
	CodebasePath         string
	OutputJSON           bool
	IncludeBusinessRules bool
	HubURL               string
	HubAPIKey            string
	ProjectID            string
}

// Learn analyzes the codebase and learns patterns
func Learn(opts LearnOptions) (*PatternData, error) {
	patterns := NewPatternData()

	codebasePath := opts.CodebasePath
	if codebasePath == "" {
		codebasePath = "."
	}

	// Analyze codebase
	if err := analyzeCodebase(codebasePath, patterns); err != nil {
		return nil, fmt.Errorf("failed to analyze codebase: %w", err)
	}

	// Handle JSON output
	if opts.OutputJSON {
		jsonData, err := json.MarshalIndent(patterns, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal patterns to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
		return patterns, nil
	}

	// Output results based on flags
	if opts.ImportsOnly {
		fmt.Println("Learning import patterns")
		if patterns.ImportPatterns.Style != "" {
			fmt.Printf("Import style: %s\n", patterns.ImportPatterns.Style)
			fmt.Printf("Default imports: %d, Named imports: %d\n", patterns.ImportPatterns.DefaultImports, patterns.ImportPatterns.NamedImports)
		}
		return patterns, nil
	}

	if opts.StructureOnly {
		fmt.Println("Analyzing folder structure")
		analyzeFolderStructure(codebasePath, patterns)
		if len(patterns.ProjectStructure) > 0 {
			for pattern, examples := range patterns.ProjectStructure {
				if len(examples) > 0 {
					fmt.Printf("Pattern: %s (examples: %s)\n", pattern, strings.Join(examples[:min(3, len(examples))], ", "))
				}
			}
		}
		return patterns, nil
	}

	if opts.NamingOnly {
		fmt.Println("Learning naming conventions")
		if len(patterns.NamingPatterns) > 0 {
			fmt.Printf("Functions: %s\n", strings.Join(getKeys(patterns.NamingPatterns), ", "))
		}
		return patterns, nil
	}

	// Full analysis output
	fmt.Println("Learning import patterns")
	fmt.Println("Learning naming conventions")

	// Find primary language
	primaryLang := findPrimaryLanguage(patterns)
	if primaryLang != "" {
		fmt.Printf("Primary language: %s\n", primaryLang)
	}

	// Output frameworks
	for fw := range patterns.Frameworks {
		fmt.Printf("Framework: %s\n", fw)
	}

	// Output naming patterns
	if len(patterns.NamingPatterns) > 0 {
		fmt.Printf("Functions: %s\n", strings.Join(getKeys(patterns.NamingPatterns), ", "))
	}

	// Output import patterns
	if patterns.ImportPatterns.Style != "" {
		fmt.Printf("Import style: %s\n", patterns.ImportPatterns.Style)
	}

	// Output code style
	if patterns.CodeStyle.IndentStyle != "" {
		fmt.Printf("Code style: %s indent (%d), %s quotes, semicolons: %s\n",
			patterns.CodeStyle.IndentStyle, patterns.CodeStyle.IndentSize,
			patterns.CodeStyle.QuoteStyle, patterns.CodeStyle.Semicolons)
	}

	// Detect source root
	if _, err := os.Stat("src"); err == nil {
		fmt.Println("Source root: src/")
	}

	// Analyze folder structure
	analyzeFolderStructure(codebasePath, patterns)

	// Fetch business rules if requested
	if opts.IncludeBusinessRules {
		rules, err := fetchBusinessRulesFromHub(opts.HubURL, opts.HubAPIKey, opts.ProjectID)
		if err != nil {
			// Log warning but don't fail - Hub unavailability is non-critical
			fmt.Printf("Warning: Failed to fetch business rules from Hub: %v\n", err)
		} else if len(rules) > 0 {
			patterns.BusinessRules = &BusinessRuleData{
				Rules: rules,
			}
			fmt.Printf("Fetched %d business rules from Hub\n", len(rules))
		} else {
			fmt.Println("No business rules found in Hub (or Hub unavailable)")
		}
	}

	// Generate output files
	if err := generateOutputFiles(patterns); err != nil {
		return nil, fmt.Errorf("failed to generate output files: %w", err)
	}

	return patterns, nil
}
