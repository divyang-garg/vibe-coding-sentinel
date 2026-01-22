// Package cli provides audit command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/divyang-garg/sentinel-hub-api/internal/hub"
	"github.com/divyang-garg/sentinel-hub-api/internal/scanner"
)

// runAudit executes the audit command
func runAudit(args []string) error {
	opts := scanner.ScanOptions{
		CodebasePath: ".",
		CIMode:       false,
		Verbose:      false,
	}

	outputFormat := "text"
	outputFile := ""

	// Parse flags
	for i, arg := range args {
		switch arg {
		case "--ci":
			opts.CIMode = true
		case "--offline":
			opts.Offline = true
		case "--verbose", "-v":
			opts.Verbose = true
		case "--vibe-check":
			opts.VibeCheck = true
		case "--vibe-only":
			opts.VibeOnly = true
			opts.VibeCheck = true // Implied
		case "--deep":
			opts.Deep = true
		case "--analyze-structure":
			opts.AnalyzeStructure = true
		case "--output":
			if i+1 < len(args) {
				outputFormat = args[i+1]
			}
		case "--output-file":
			if i+1 < len(args) {
				outputFile = args[i+1]
			}
		default:
			// Treat as codebase path if it doesn't start with --
			if len(arg) > 0 && arg[0] != '-' && opts.CodebasePath == "." {
				opts.CodebasePath = arg
			}
		}
	}

	// Perform scan
	fmt.Println("🔍 Running security audit...")
	result, err := scanner.Scan(opts)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// If --deep flag is set and not offline, try Hub integration
	if opts.Deep && !opts.Offline {
		hubClient := hub.NewClient(getHubURL(), getAPIKey())
		if hubClient.IsAvailable() {
			if !opts.CIMode {
				fmt.Println("🔗 Hub available, performing deep AST analysis...")
			}

			// Perform Hub-based AST analysis and merge results
			astFindings, err := performHubAnalysis(hubClient, opts.CodebasePath, !opts.CIMode)
			if err != nil {
				if !opts.CIMode {
					fmt.Printf("⚠️  Hub analysis failed: %v\n", err)
				}
			} else {
				// Merge AST findings into results
				result = mergeHubFindings(result, astFindings)
				if !opts.CIMode && len(astFindings) > 0 {
					fmt.Printf("✅ Hub analysis complete: %d AST findings added\n", len(astFindings))
				}
			}
		} else if !opts.CIMode {
			fmt.Println("ℹ️  Hub not available, using local scanning only")
		}
	}

	// Display results
	displayResults(result, opts.CIMode)

	// Save to file if requested
	if outputFile != "" {
		if saveErr := saveResults(result, outputFile, outputFormat); saveErr != nil {
			return fmt.Errorf("failed to save results: %w", saveErr)
		}
		fmt.Printf("📄 Results saved to: %s\n", outputFile)
	}

	// Exit with appropriate code
	if !result.Success {
		if opts.CIMode {
			fmt.Println("⛔ Audit FAILED. Build rejected.")
		} else {
			fmt.Println("⛔ Audit FAILED. Issues found.")
		}
		// Don't exit if we're in a test environment
		// Automatically detect test mode without requiring env var setup:
		// 1. Check SENTINEL_TEST_MODE env var (set by test init() functions)
		// 2. Check if we're running as a test binary by examining executable path
		// This allows tests to run without manual env var setup
		if os.Getenv("SENTINEL_TEST_MODE") == "" {
			// Check if we're running under go test by examining the executable path
			if exe, err := os.Executable(); err == nil {
				if strings.Contains(exe, ".test") || strings.Contains(exe, "/_test/") || strings.Contains(exe, "go-build") {
					// We're in test mode, return error instead of exiting
					return fmt.Errorf("audit failed: %d findings", len(result.Findings))
				}
			}
			// Not in test mode, exit normally
			os.Exit(1)
		}
		// In test mode (env var set), return error instead of exiting
		return fmt.Errorf("audit failed: %d findings", len(result.Findings))
	}

	if !opts.CIMode {
		fmt.Println("✅ Audit PASSED.")
	}
	return nil
}

// displayResults displays scan results
func displayResults(result *scanner.Result, ciMode bool) {
	if ciMode {
		// CI mode: minimal output
		fmt.Printf("Findings: %d\n", len(result.Findings))
		if len(result.Summary) > 0 {
			fmt.Println("Summary:")
			for pattern, count := range result.Summary {
				fmt.Printf("  %s: %d\n", pattern, count)
			}
		}
		return
	}

	// Interactive mode: detailed output
	fmt.Println("\n📊 Audit Results")
	fmt.Println("================")
	fmt.Printf("Timestamp: %s\n", result.Timestamp)
	fmt.Printf("Status: ")
	if result.Success {
		fmt.Println("✅ PASSED")
	} else {
		fmt.Println("❌ FAILED")
	}
	fmt.Printf("Total Findings: %d\n\n", len(result.Findings))

	if len(result.Summary) > 0 {
		fmt.Println("Summary by Type:")
		for pattern, count := range result.Summary {
			fmt.Printf("  %s: %d\n", pattern, count)
		}
		fmt.Println()
	}

	if len(result.Findings) > 0 {
		fmt.Println("Findings:")
		for i, finding := range result.Findings {
			if i >= 10 {
				fmt.Printf("  ... and %d more findings\n", len(result.Findings)-i)
				break
			}
			fmt.Printf("  [%s] %s:%d - %s\n", finding.Severity, finding.File, finding.Line, finding.Message)
			if finding.Pattern != "" {
				fmt.Printf("    Pattern: %s\n", finding.Pattern)
			}
		}
	}
}
