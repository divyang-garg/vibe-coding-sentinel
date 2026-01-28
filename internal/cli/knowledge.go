// Package cli provides knowledge command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/divyang-garg/sentinel-hub-api/internal/config"
)

// KnowledgeEntry represents a knowledge base entry
type KnowledgeEntry struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Source    string    `json:"source"` // File path or URL
	Type      string    `json:"type"`   // requirement, decision, pattern, etc.
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"` // draft, approved, archived
}

// KnowledgeBase represents the knowledge base
type KnowledgeBase struct {
	Version string           `json:"version"`
	Entries []KnowledgeEntry `json:"entries"`
}

// runKnowledge executes the knowledge command
func runKnowledge(args []string) error {
	if len(args) == 0 {
		return listKnowledge()
	}

	switch args[0] {
	case "add":
		return addKnowledge(args[1:])
	case "list":
		return listKnowledge()
	case "search":
		return searchKnowledge(args[1:])
	case "export":
		return exportKnowledge(args[1:])
	case "import":
		return importKnowledge(args[1:])
	case "extract":
		return runExtract(args[1:])
	case "help", "--help", "-h":
		return printKnowledgeHelp()
	default:
		return listKnowledge()
	}
}

// listKnowledge lists all knowledge entries
func listKnowledge() error {
	kb, err := loadKnowledge()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("📚 Knowledge base is empty")
			fmt.Println("\nUse 'sentinel knowledge add' to add entries.")
			return nil
		}
		return fmt.Errorf("unable to load knowledge base: %w", err)
	}

	if len(kb.Entries) == 0 {
		fmt.Println("📚 Knowledge base is empty")
		return nil
	}

	fmt.Printf("📚 Knowledge Base (%d entries)\n", len(kb.Entries))
	fmt.Println(strings.Repeat("=", 60))

	for i, entry := range kb.Entries {
		fmt.Printf("\n%d. [%s] %s\n", i+1, entry.Type, entry.Title)
		if entry.Source != "" {
			fmt.Printf("   Source: %s\n", entry.Source)
		}
		if len(entry.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(entry.Tags, ", "))
		}
		fmt.Printf("   Status: %s\n", entry.Status)
	}

	return nil
}

// addKnowledge adds a new knowledge entry
func addKnowledge(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: sentinel knowledge add <title> <content> [type] [tags...]")
	}

	title := args[0]
	content := args[1]
	entryType := "requirement"
	if len(args) > 2 {
		entryType = args[2]
	}

	tags := []string{}
	if len(args) > 3 {
		tags = args[3:]
	}

	kb, err := loadKnowledge()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to load knowledge base: %w", err)
	}

	entry := KnowledgeEntry{
		ID:        generateID(),
		Title:     title,
		Content:   content,
		Type:      entryType,
		Tags:      tags,
		Status:    "draft",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	kb.Entries = append(kb.Entries, entry)

	if err := saveKnowledge(kb); err != nil {
		return fmt.Errorf("unable to save knowledge base: %w", err)
	}

	fmt.Printf("✅ Added knowledge entry: %s\n", title)
	return nil
}

// searchKnowledge searches knowledge entries
func searchKnowledge(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sentinel knowledge search <query>")
	}

	query := strings.ToLower(strings.Join(args, " "))
	kb, err := loadKnowledge()
	if err != nil {
		return fmt.Errorf("unable to load knowledge base: %w", err)
	}

	matches := []KnowledgeEntry{}
	for _, entry := range kb.Entries {
		if strings.Contains(strings.ToLower(entry.Title), query) ||
			strings.Contains(strings.ToLower(entry.Content), query) ||
			containsTag(entry.Tags, query) {
			matches = append(matches, entry)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No knowledge entries found matching: %s\n", query)
		return nil
	}

	fmt.Printf("📚 Found %d matching entries:\n", len(matches))
	for i, entry := range matches {
		fmt.Printf("\n%d. [%s] %s\n", i+1, entry.Type, entry.Title)
		fmt.Printf("   %s\n", truncate(entry.Content, 100))
	}

	return nil
}

// exportKnowledge exports knowledge base to a file
func exportKnowledge(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sentinel knowledge export <file>")
	}

	outputFile := args[0]
	kb, err := loadKnowledge()
	if err != nil {
		return fmt.Errorf("unable to load knowledge base: %w", err)
	}

	data, err := json.MarshalIndent(kb, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to format knowledge data: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("unable to save file: %w", err)
	}

	fmt.Printf("✅ Knowledge base exported to: %s\n", outputFile)
	return nil
}

// importKnowledge imports knowledge base from a file
func importKnowledge(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sentinel knowledge import <file>")
	}

	inputFile := args[0]
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}

	var kb KnowledgeBase
	if err := json.Unmarshal(data, &kb); err != nil {
		return fmt.Errorf("unable to parse knowledge data: %w", err)
	}

	if err := saveKnowledge(&kb); err != nil {
		return fmt.Errorf("unable to save knowledge base: %w", err)
	}

	fmt.Printf("✅ Imported %d knowledge entries\n", len(kb.Entries))
	return nil
}

// loadKnowledge loads knowledge base from disk
func loadKnowledge() (*KnowledgeBase, error) {
	kbPath := getKnowledgePath()
	data, err := os.ReadFile(kbPath)
	if err != nil {
		return &KnowledgeBase{Version: "1.0", Entries: []KnowledgeEntry{}}, err
	}

	var kb KnowledgeBase
	if err := json.Unmarshal(data, &kb); err != nil {
		return nil, err
	}

	return &kb, nil
}

// saveKnowledge saves knowledge base to disk
func saveKnowledge(kb *KnowledgeBase) error {
	kbPath := getKnowledgePath()
	if err := os.MkdirAll(filepath.Dir(kbPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(kb, "", "  ")
	if err != nil {
		return err
	}

	return config.WriteFile(kbPath, string(data))
}

// getKnowledgePath returns the path to the knowledge base file
func getKnowledgePath() string {
	return ".sentinel/knowledge.json"
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("kb_%d", time.Now().UnixNano())
}

func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// printKnowledgeHelp displays help for the knowledge command
func printKnowledgeHelp() error {
	fmt.Println("Usage: sentinel knowledge <command> [options]")
	fmt.Println("")
	fmt.Println("Manage knowledge base entries.")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  (no command)        List all knowledge entries")
	fmt.Println("  add <title> <content> [type] [tags...]  Add new entry")
	fmt.Println("  list                List all entries")
	fmt.Println("  search <query>      Search entries")
	fmt.Println("  extract <file>      Extract business rules from document")
	fmt.Println("  export <file>       Export to JSON")
	fmt.Println("  import <file>       Import from JSON")
	fmt.Println("  help                Show this help")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  sentinel knowledge add 'Auth Flow' 'User must authenticate' requirement security")
	fmt.Println("  sentinel knowledge search auth")
	return nil
}
