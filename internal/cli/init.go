// Package cli provides init command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/divyang-garg/sentinel-hub-api/internal/config"
	"github.com/divyang-garg/sentinel-hub-api/internal/constants"
)

// runInit initializes Sentinel in the current project
func runInit(args []string) error {
	fmt.Println("🏗️  Sentinel: Initializing Factory...")

	// 1. BROWNFIELD CHECK - Backup existing rules BEFORE creating directories
	if err := backupExistingRules(".cursor/rules"); err != nil {
		fmt.Printf("❌ Failed to backup existing rules: %v\n", err)
		fmt.Println("   Aborting initialization to prevent data loss.")
		return err
	}

	// 2. SCAFFOLDING - Create directories
	dirs := []string{".cursor/rules", ".github/workflows", "docs/knowledge", "docs/external", "scripts"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, constants.DefaultDirPerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// 3. CONSTITUTION
	if err := config.WriteFile(".cursor/rules/00-constitution.md", constants.Constitution); err != nil {
		return fmt.Errorf("failed to write constitution: %w", err)
	}
	if err := config.WriteFile(".cursor/rules/01-firewall.md", constants.Firewall); err != nil {
		return fmt.Errorf("failed to write firewall: %w", err)
	}
	if err := config.WriteFile("docs/knowledge/client-brief.md", "# Requirements\n"); err != nil {
		return fmt.Errorf("failed to write client brief: %w", err)
	}

	// 4. INTERACTIVE MATRIX
	reader := bufio.NewReader(os.Stdin)

	// -- STACK --
	fmt.Println("\n--- Service Line ---")
	fmt.Println("1) 🌐 Web App")
	fmt.Println("2) 📱 Mobile (Cross-Platform)")
	fmt.Println("3) 🍏 Mobile (Native)")
	fmt.Println("4) 🛍️  Commerce")
	fmt.Println("5) 🧠 AI & Data")
	fmt.Print("Selection: ")
	stack, _ := reader.ReadString('\n')
	stack = strings.TrimSpace(stack)

	switch stack {
	case "1":
		if err := config.WriteFile(".cursor/rules/web.md", constants.WebRules); err != nil {
			return fmt.Errorf("failed to write web rules: %w", err)
		}
	case "2":
		if err := config.WriteFile(".cursor/rules/mobile.md", constants.MobileCrossRules); err != nil {
			return fmt.Errorf("failed to write mobile rules: %w", err)
		}
	case "3":
		if err := config.WriteFile(".cursor/rules/mobile.md", constants.MobileNativeRules); err != nil {
			return fmt.Errorf("failed to write mobile rules: %w", err)
		}
	case "4":
		if err := config.WriteFile(".cursor/rules/commerce.md", constants.CommerceRules); err != nil {
			return fmt.Errorf("failed to write commerce rules: %w", err)
		}
	case "5":
		if err := config.WriteFile(".cursor/rules/ai.md", constants.AIRules); err != nil {
			return fmt.Errorf("failed to write AI rules: %w", err)
		}
	}

	// -- DATABASE --
	fmt.Println("\n--- Database ---")
	fmt.Println("1) SQL")
	fmt.Println("2) NoSQL")
	fmt.Println("3) None")
	fmt.Print("Selection: ")
	db, _ := reader.ReadString('\n')
	db = strings.TrimSpace(db)

	switch db {
	case "1":
		if err := config.WriteFile(".cursor/rules/db-sql.md", constants.SQLRules); err != nil {
			return fmt.Errorf("failed to write SQL rules: %w", err)
		}
	case "2":
		if err := config.WriteFile(".cursor/rules/db-nosql.md", constants.NoSQLRules); err != nil {
			return fmt.Errorf("failed to write NoSQL rules: %w", err)
		}
	}

	// -- PROTOCOL --
	fmt.Println("\n--- Protocol ---")
	fmt.Print("Support SOAP/Legacy? [y/N]: ")
	soap, _ := reader.ReadString('\n')
	if strings.Contains(strings.ToLower(soap), "y") {
		if err := config.WriteFile(".cursor/rules/proto-soap.md", constants.SOAPRules); err != nil {
			return fmt.Errorf("failed to write SOAP rules: %w", err)
		}
	}

	// 5. SECURE GIT
	if err := config.SecureGitIgnore(); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}
	if err := config.CreateCI(); err != nil {
		return fmt.Errorf("failed to create CI workflow: %w", err)
	}

	fmt.Println("✅ Environment Secured. Rules Injected (Hidden).")
	return nil
}

// backupExistingRules safely backs up existing rules if they exist
func backupExistingRules(rulesPath string) error {
	// Check if directory exists
	info, err := os.Stat(rulesPath)
	if os.IsNotExist(err) {
		return nil // No existing rules, nothing to backup
	}
	if err != nil {
		return fmt.Errorf("failed to stat rules directory: %w", err)
	}

	// Verify it's a directory (not a file)
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", rulesPath)
	}

	// Check if directory has files
	entries, err := os.ReadDir(rulesPath)
	if err != nil {
		return fmt.Errorf("failed to read rules directory: %w", err)
	}

	// If directory is empty, no need to backup - clean it up
	if len(entries) == 0 {
		if err := os.Remove(rulesPath); err != nil {
			return fmt.Errorf("failed to remove empty rules directory: %w", err)
		}
		return nil
	}

	// Generate unique backup name (handle collisions)
	baseBackup := fmt.Sprintf("%s_backup_%d", rulesPath, time.Now().Unix())
	backupPath := baseBackup
	counter := 0
	for {
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			break // Path is available
		}
		counter++
		backupPath = fmt.Sprintf("%s_%d", baseBackup, counter)
		if counter > constants.MaxBackupAttempts {
			return fmt.Errorf("unable to find available backup path after %d attempts", constants.MaxBackupAttempts)
		}
	}

	// Perform atomic rename with error checking
	if err := os.Rename(rulesPath, backupPath); err != nil {
		return fmt.Errorf("failed to rename rules directory to backup: %w", err)
	}

	// Verify backup was successful
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	fmt.Printf("✅ Existing rules backed up to %s\n", backupPath)
	return nil
}
