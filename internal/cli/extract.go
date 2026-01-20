// Package cli provides knowledge extraction command
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/divyang-garg/sentinel-hub-api/internal/extraction"
	cachepkg "github.com/divyang-garg/sentinel-hub-api/internal/extraction/cache"
)

// runExtract handles the 'sentinel knowledge extract' command
func runExtract(args []string) error {
	if len(args) < 1 {
		return printExtractHelp()
	}

	// Check for help flags first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return printExtractHelp()
		}
	}

	// Parse flags
	opts := parseExtractOptions(args)

	// Parse document using document parser
	parser, err := extraction.NewDocumentParser(opts.inputFile)
	if err != nil {
		return fmt.Errorf("document parsing failed: %w", err)
	}

	content, err := parser.Parse(opts.inputFile)
	if err != nil {
		return fmt.Errorf("failed to extract text from document: %w", err)
	}

	// Create extractor with dependencies
	extractor, err := createExtractor(opts)
	if err != nil {
		return fmt.Errorf("failed to create extractor: %w", err)
	}

	// Perform extraction
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var result *extraction.ExtractResult
	if opts.useBatch {
		result, err = extractor.ExtractBatch(ctx, extraction.ExtractRequest{
			Text:       content,
			Source:     opts.inputFile,
			SchemaType: opts.schemaType,
			Options: extraction.ExtractOptions{
				UseLLM:        opts.useLLM,
				UseFallback:   opts.useFallback,
				MinConfidence: opts.minConfidence,
			},
		})
	} else {
		result, err = extractor.Extract(ctx, extraction.ExtractRequest{
			Text:       content,
			Source:     opts.inputFile,
			SchemaType: opts.schemaType,
			Options: extraction.ExtractOptions{
				UseLLM:        opts.useLLM,
				UseFallback:   opts.useFallback,
				MinConfidence: opts.minConfidence,
			},
		})
	}
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Output results
	return outputResults(result, opts)
}

type extractOptions struct {
	inputFile     string
	outputFile    string
	schemaType    string
	useLLM        bool
	useFallback   bool
	useBatch      bool
	minConfidence float64
	jsonOutput    bool
	saveToKB      bool
}

func parseExtractOptions(args []string) extractOptions {
	opts := extractOptions{
		schemaType:    "business_rule", // Default schema type
		useLLM:        true,
		useFallback:   true,
		minConfidence: 0.6,
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--no-llm":
			opts.useLLM = false
		case "--no-fallback":
			opts.useFallback = false
		case "--json":
			opts.jsonOutput = true
		case "--save":
			opts.saveToKB = true
		case "--batch":
			opts.useBatch = true
		case "--schema":
			if i+1 < len(args) {
				opts.schemaType = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				opts.outputFile = args[i+1]
				i++
			}
		case "--min-confidence":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%f", &opts.minConfidence)
				i++
			}
		default:
			if opts.inputFile == "" && !strings.HasPrefix(args[i], "-") {
				opts.inputFile = args[i]
			}
		}
	}

	return opts
}

func createExtractor(opts extractOptions) (extraction.Extractor, error) {
	// Create dependencies
	promptBuilder := extraction.NewPromptBuilder()
	parser := extraction.NewResponseParser()
	scorer := extraction.NewConfidenceScorer()
	fallback := extraction.NewFallbackExtractor()

	// Use file-based cache if available, fallback to in-memory
	cacheDir := os.Getenv("SENTINEL_CACHE_DIR")
	if cacheDir == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE") // Windows
		}
		if homeDir != "" {
			cacheDir = filepath.Join(homeDir, ".sentinel", "cache")
		}
	}

	var cache extraction.Cache
	if cacheDir != "" {
		fileCache, err := cachepkg.NewFileCache(cacheDir)
		if err != nil {
			cache = newSimpleCache() // Fallback to in-memory
		} else {
			cache = fileCache
		}
	} else {
		cache = newSimpleCache()
	}

	logger := newCLILogger()

	// Create LLM client (uses hub/api/llm infrastructure)
	var llmClient extraction.LLMClient
	llmClientAdapter, err := newLLMClientAdapter()
	if err != nil {
		if opts.useLLM {
			return nil, fmt.Errorf("LLM client initialization failed: %w. Use --no-llm for regex-only extraction", err)
		}
		// If --no-llm, create a no-op LLM client
		llmClient = &noOpLLMClient{}
	} else {
		llmClient = llmClientAdapter
	}

	return extraction.NewKnowledgeExtractor(
		llmClient,
		promptBuilder,
		parser,
		scorer,
		fallback,
		cache,
		logger,
	), nil
}

func outputResults(result *extraction.ExtractResult, opts extractOptions) error {
	// Filter by confidence
	filtered := filterByConfidence(result.BusinessRules, opts.minConfidence)

	if opts.jsonOutput {
		data, _ := json.MarshalIndent(result, "", "  ")
		if opts.outputFile != "" {
			if err := os.WriteFile(opts.outputFile, data, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("Results written to %s\n", opts.outputFile)
		} else {
			fmt.Println(string(data))
		}
	} else {
		fmt.Printf("Extracted %d business rules (source: %s, confidence: %.2f)\n\n",
			len(filtered), result.Source, result.Confidence)

		for i, rule := range filtered {
			fmt.Printf("%d. [%s] %s (confidence: %.2f)\n", i+1, rule.ID, rule.Title, rule.Confidence)
			fmt.Printf("   %s\n", rule.Description)
			if len(rule.Specification.Constraints) > 0 {
				fmt.Printf("   Constraints: %d\n", len(rule.Specification.Constraints))
			}
			fmt.Println()
		}
	}

	// Save to knowledge base if requested
	if opts.saveToKB {
		return saveToKnowledgeBase(filtered)
	}

	return nil
}

func filterByConfidence(rules []extraction.BusinessRule, min float64) []extraction.BusinessRule {
	var filtered []extraction.BusinessRule
	for _, r := range rules {
		if r.Confidence >= min {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func saveToKnowledgeBase(rules []extraction.BusinessRule) error {
	kb, err := loadKnowledge()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load knowledge base: %w", err)
	}

	for _, rule := range rules {
		entry := KnowledgeEntry{
			ID:        rule.ID,
			Title:     rule.Title,
			Content:   rule.Description,
			Source:    rule.Traceability.SourceDocument,
			Type:      "business_rule",
			Tags:      []string{"extracted", rule.Priority},
			Status:    rule.Status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		kb.Entries = append(kb.Entries, entry)
	}

	if err := saveKnowledge(kb); err != nil {
		return fmt.Errorf("failed to save to knowledge base: %w", err)
	}

	fmt.Printf("Saved %d rules to knowledge base\n", len(rules))
	return nil
}

func printExtractHelp() error {
	fmt.Println("Usage: sentinel knowledge extract <file> [options]")
	fmt.Println("")
	fmt.Println("Extract business rules from documents using LLM.")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --schema          Schema type: business_rule|entity|api_contract|user_journey|glossary (default: business_rule)")
	fmt.Println("  --batch           Enable batch processing for large documents")
	fmt.Println("  --no-llm          Skip LLM extraction (regex only)")
	fmt.Println("  --no-fallback     Disable regex fallback")
	fmt.Println("  --min-confidence  Minimum confidence threshold (default: 0.6)")
	fmt.Println("  --json            Output as JSON")
	fmt.Println("  --save            Save to knowledge base")
	fmt.Println("  -o, --output      Write results to file")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  sentinel knowledge extract requirements.md")
	fmt.Println("  sentinel knowledge extract spec.txt --json --save")
	return nil
}

// Helper implementations are in extract_helpers.go
