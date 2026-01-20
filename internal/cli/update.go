// Package cli provides update-rules command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/divyang-garg/sentinel-hub-api/internal/constants"
	"github.com/divyang-garg/sentinel-hub-api/internal/hub"
)

// runUpdateRules updates rules from external sources
func runUpdateRules(args []string) error {
	fmt.Println("🔄 Checking for rule updates...")

	// Parse flags
	force := false
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		}
	}

	rulesDir := ".cursor/rules"

	// Check if rules directory exists
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		fmt.Println("⚠️  Rules directory not found. Run 'sentinel init' first.")
		return fmt.Errorf("rules directory not found")
	}

	// Try to connect to Hub
	hubURL := getHubURL()
	apiKey := getAPIKey()

	client := hub.NewClient(hubURL, apiKey)
	if !client.IsAvailable() {
		fmt.Println("⚠️  Hub is not available. Using default rule templates.")
		return updateRulesFromDefaults(rulesDir, force)
	}

	fmt.Println("✅ Hub connected")
	fmt.Println("⚠️  Hub-based rule updates not yet implemented")
	fmt.Println("\n💡 Falling back to default templates...")

	return updateRulesFromDefaults(rulesDir, force)
}

// updateRulesFromDefaults updates rules from built-in constants
func updateRulesFromDefaults(rulesDir string, force bool) error {
	// Backup existing rules if they exist
	backupDir := filepath.Join(rulesDir, ".backup-"+time.Now().Format("20060102-150405"))
	if !force {
		fmt.Printf("📦 Creating backup at %s\n", backupDir)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}

		// Copy existing rules to backup
		entries, err := os.ReadDir(rulesDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
					continue
				}
				src := filepath.Join(rulesDir, entry.Name())
				dst := filepath.Join(backupDir, entry.Name())
				data, err := os.ReadFile(src)
				if err == nil {
					os.WriteFile(dst, data, 0644)
				}
			}
		}
	}

	// Update core rules from constants
	// Only update rules that have corresponding constants
	rules := map[string]string{
		"constitution.md": constants.Constitution,
		"security.md":     constants.Firewall,
	}

	updated := 0
	for filename, content := range rules {
		filePath := filepath.Join(rulesDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			fmt.Printf("❌ Failed to update %s: %v\n", filename, err)
		} else {
			fmt.Printf("✅ Updated %s\n", filename)
			updated++
		}
	}

	fmt.Printf("\n🎯 Updated %d rule files\n", updated)

	if !force && updated > 0 {
		fmt.Println("\n💡 Review changes and restore from backup if needed:")
		fmt.Printf("   Backup: %s\n", backupDir)
	}

	return nil
}
