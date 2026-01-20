// Package cli provides review command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// runReview executes the review command
func runReview(args []string) error {
	kb, err := loadKnowledge()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("📚 No knowledge base found")
			return nil
		}
		return fmt.Errorf("failed to load knowledge: %w", err)
	}

	// Find pending entries (draft status)
	pending := []KnowledgeEntry{}
	for _, entry := range kb.Entries {
		if entry.Status == "draft" {
			pending = append(pending, entry)
		}
	}

	if len(pending) == 0 {
		fmt.Println("✅ No pending knowledge entries to review")
		return nil
	}

	fmt.Printf("📋 Reviewing %d pending knowledge entries\n", len(pending))
	fmt.Println(strings.Repeat("=", 60))

	reader := bufio.NewReader(os.Stdin)

	for i, entry := range pending {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(pending), entry.Title)
		fmt.Printf("Type: %s\n", entry.Type)
		fmt.Printf("Content: %s\n", truncate(entry.Content, 200))
		if len(entry.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(entry.Tags, ", "))
		}
		fmt.Print("\n[a]pprove, [r]eject, [s]kip, [q]uit: ")

		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		switch response {
		case "a", "approve":
			entry.Status = "approved"
			entry.UpdatedAt = time.Now()
			updateKnowledgeEntry(kb, entry)
			fmt.Println("✅ Approved")
		case "r", "reject":
			entry.Status = "archived"
			entry.UpdatedAt = time.Now()
			updateKnowledgeEntry(kb, entry)
			fmt.Println("❌ Rejected")
		case "s", "skip":
			fmt.Println("⏭️  Skipped")
		case "q", "quit":
			fmt.Println("👋 Review cancelled")
			return saveKnowledge(kb)
		default:
			fmt.Println("⚠️  Invalid choice, skipping")
		}
	}

	if err := saveKnowledge(kb); err != nil {
		return fmt.Errorf("failed to save knowledge: %w", err)
	}

	fmt.Println("\n✅ Review complete")
	return nil
}

// updateKnowledgeEntry updates an entry in the knowledge base
func updateKnowledgeEntry(kb *KnowledgeBase, updated KnowledgeEntry) {
	for i, entry := range kb.Entries {
		if entry.ID == updated.ID {
			kb.Entries[i] = updated
			return
		}
	}
}
