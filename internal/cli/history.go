// Package cli provides history command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// AuditHistoryEntry represents a single audit run
type AuditHistoryEntry struct {
	Timestamp     time.Time      `json:"timestamp"`
	Success       bool           `json:"success"`
	TotalFindings int            `json:"total_findings"`
	ByType        map[string]int `json:"by_type"`
	BySeverity    map[string]int `json:"by_severity"`
	Duration      time.Duration  `json:"duration"`
}

// AuditHistory contains all audit history
type AuditHistory struct {
	Entries []AuditHistoryEntry `json:"entries"`
}

// runHistory executes the history command
func runHistory(args []string) error {
	last := 10 // Default to last 10
	jsonOutput := false

	// Parse flags
	for i, arg := range args {
		switch arg {
		case "--json":
			jsonOutput = true
		case "--last", "-n":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &last)
			}
		}
	}

	history, err := loadHistory()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("📊 No audit history found")
			fmt.Println("\nRun 'sentinel audit' to start building history.")
			return nil
		}
		return fmt.Errorf("failed to load history: %w", err)
	}

	if len(history.Entries) == 0 {
		fmt.Println("📊 No audit history found")
		return nil
	}

	// Limit to last N entries
	start := 0
	if len(history.Entries) > last {
		start = len(history.Entries) - last
	}
	entries := history.Entries[start:]

	if jsonOutput {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal history: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Display history
	fmt.Printf("📊 Audit History (last %d runs)\n", len(entries))
	fmt.Println("=" + string(make([]byte, 60)))

	for i, entry := range entries {
		status := "✅ PASSED"
		if !entry.Success {
			status = "❌ FAILED"
		}

		fmt.Printf("\n%d. %s - %s\n", i+1, entry.Timestamp.Format("2006-01-02 15:04"), status)
		fmt.Printf("   Findings: %d (Duration: %v)\n", entry.TotalFindings, entry.Duration)

		if len(entry.ByType) > 0 {
			fmt.Print("   By Type: ")
			first := true
			for typ, count := range entry.ByType {
				if !first {
					fmt.Print(", ")
				}
				fmt.Printf("%s: %d", typ, count)
				first = false
			}
			fmt.Println()
		}
	}

	// Show trend if we have multiple entries
	if len(entries) > 1 {
		fmt.Println("\n📈 Trend:")
		first := entries[0]
		last := entries[len(entries)-1]

		delta := last.TotalFindings - first.TotalFindings
		if delta > 0 {
			fmt.Printf("   ⬆️  Findings increased by %d (%.1f%%)\n", delta, float64(delta)/float64(first.TotalFindings)*100)
		} else if delta < 0 {
			fmt.Printf("   ⬇️  Findings decreased by %d (%.1f%%)\n", -delta, float64(-delta)/float64(first.TotalFindings)*100)
		} else {
			fmt.Println("   ➡️  No change in findings")
		}
	}

	return nil
}

// AddToHistory adds a new audit result to history
func AddToHistory(success bool, totalFindings int, byType, bySeverity map[string]int, duration time.Duration) error {
	history, err := loadHistory()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	entry := AuditHistoryEntry{
		Timestamp:     time.Now(),
		Success:       success,
		TotalFindings: totalFindings,
		ByType:        byType,
		BySeverity:    bySeverity,
		Duration:      duration,
	}

	history.Entries = append(history.Entries, entry)

	// Keep only last 100 entries
	if len(history.Entries) > 100 {
		history.Entries = history.Entries[len(history.Entries)-100:]
	}

	return saveHistory(history)
}

// loadHistory loads audit history from disk
func loadHistory() (*AuditHistory, error) {
	data, err := os.ReadFile(".sentinel/audit-history.json")
	if err != nil {
		return &AuditHistory{Entries: []AuditHistoryEntry{}}, err
	}

	var history AuditHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	return &history, nil
}

// saveHistory saves audit history to disk
func saveHistory(history *AuditHistory) error {
	// Ensure directory exists
	if err := os.MkdirAll(".sentinel", 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(".sentinel/audit-history.json", data, 0644)
}
