// Package cli provides baseline command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"
	"os"
	"time"
)

// runBaseline executes the baseline command
func runBaseline(args []string) error {
	if len(args) == 0 {
		return showBaseline()
	}

	switch args[0] {
	case "add":
		return addToBaseline(args[1:])
	case "remove":
		return removeFromBaseline(args[1:])
	case "clear":
		return clearBaseline()
	case "export":
		return exportBaseline(args[1:])
	case "import":
		return importBaseline(args[1:])
	case "help", "--help", "-h":
		return printBaselineHelp()
	default:
		return fmt.Errorf("unknown baseline command: %s\n\nRun 'sentinel baseline help' for usage", args[0])
	}
}

// showBaseline displays current baseline entries
func showBaseline() error {
	baseline, err := loadBaseline()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("📋 Baseline is empty")
			fmt.Println("\nUse 'sentinel baseline add' to accept findings.")
			return nil
		}
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	if len(baseline.Entries) == 0 {
		fmt.Println("📋 Baseline is empty")
		return nil
	}

	fmt.Printf("📋 Baseline (%d entries)\n", len(baseline.Entries))
	fmt.Println("=" + string(make([]byte, 50)))

	for i, entry := range baseline.Entries {
		fmt.Printf("\n%d. %s:%d - %s\n", i+1, entry.File, entry.Line, entry.Pattern)
		if entry.Reason != "" {
			fmt.Printf("   Reason: %s\n", entry.Reason)
		}
		fmt.Printf("   Added: %s by %s\n", entry.AddedAt.Format("2006-01-02"), entry.AddedBy)
	}

	return nil
}

// addToBaseline adds a finding to the baseline
func addToBaseline(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: sentinel baseline add <file> <line> [reason]")
	}

	file := args[0]
	var line int
	var reason string

	_, err := fmt.Sscanf(args[1], "%d", &line)
	if err != nil {
		return fmt.Errorf("invalid line number: %s", args[1])
	}

	if len(args) > 2 {
		reason = args[2]
	} else {
		reason = "Accepted via baseline command"
	}

	// Load existing baseline
	baseline, err := loadBaseline()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	// Get current user (simplified - in production would use git config)
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	// Create new entry
	entry := BaselineEntry{
		Pattern: "manual_baseline",
		File:    file,
		Line:    line,
		Reason:  reason,
		AddedBy: user,
		AddedAt: time.Now(),
		Hash:    fmt.Sprintf("%s:%d", file, line),
	}

	// Add to baseline
	baseline.Entries = append(baseline.Entries, entry)

	// Save baseline
	if err := saveBaseline(baseline); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	fmt.Printf("✅ Added to baseline: %s:%d\n", file, line)
	if reason != "" {
		fmt.Printf("   Reason: %s\n", reason)
	}

	return nil
}

// removeFromBaseline removes a finding from the baseline
func removeFromBaseline(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sentinel baseline remove <index>")
	}

	var index int
	_, err := fmt.Sscanf(args[0], "%d", &index)
	if err != nil {
		return fmt.Errorf("invalid index: %s", args[0])
	}

	// Load existing baseline
	baseline, err := loadBaseline()
	if err != nil {
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	if index < 1 || index > len(baseline.Entries) {
		return fmt.Errorf("index out of range: %d (baseline has %d entries)", index, len(baseline.Entries))
	}

	// Remove entry (convert to 0-based index)
	removed := baseline.Entries[index-1]
	baseline.Entries = append(baseline.Entries[:index-1], baseline.Entries[index:]...)

	// Save baseline
	if err := saveBaseline(baseline); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	fmt.Printf("✅ Removed from baseline: %s:%d\n", removed.File, removed.Line)

	return nil
}

// clearBaseline removes all baseline entries
func clearBaseline() error {
	baselinePath := getBaselinePath()

	// Check if baseline exists
	if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
		fmt.Println("📋 Baseline is already empty")
		return nil
	}

	// Remove baseline file
	if err := os.Remove(baselinePath); err != nil {
		return fmt.Errorf("failed to remove baseline: %w", err)
	}

	fmt.Println("✅ Baseline cleared")
	return nil
}

// printBaselineHelp displays help for the baseline command
func printBaselineHelp() error {
	fmt.Println("Usage: sentinel baseline <command> [options]")
	fmt.Println("")
	fmt.Println("Manage accepted findings (baseline).")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  (no command)        Show current baseline")
	fmt.Println("  add <file> <line>   Add finding to baseline")
	fmt.Println("  remove <index>      Remove finding from baseline")
	fmt.Println("  clear               Clear all baseline entries")
	fmt.Println("  export <file>       Export baseline to JSON file")
	fmt.Println("  import <file>       Import baseline from JSON file")
	fmt.Println("  help                Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  sentinel baseline                    # Show all baseline entries")
	fmt.Println("  sentinel baseline add app.js 42      # Accept finding at app.js:42")
	fmt.Println("  sentinel baseline remove 1           # Remove first entry")
	fmt.Println("  sentinel baseline export baseline.json")
	return nil
}
