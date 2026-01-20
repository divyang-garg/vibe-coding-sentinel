// Package cli provides hooks command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// runInstallHooks installs git hooks
func runInstallHooks(args []string) error {
	fmt.Println("⚙️  Installing git hooks...")

	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository")
	}

	// Create hooks directory if it doesn't exist
	hooksDir := ".git/hooks"
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Install pre-commit hook
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	preCommitScript := `#!/bin/sh
# Sentinel pre-commit hook

echo "🔍 Running Sentinel pre-commit checks..."
sentinel audit --ci

exit $?
`

	if err := os.WriteFile(preCommitPath, []byte(preCommitScript), 0755); err != nil {
		return fmt.Errorf("failed to write pre-commit hook: %w", err)
	}

	fmt.Println("✅ Installed pre-commit hook")

	// Install pre-push hook
	prePushPath := filepath.Join(hooksDir, "pre-push")
	prePushScript := `#!/bin/sh
# Sentinel pre-push hook

echo "🔍 Running Sentinel pre-push checks..."
sentinel audit --ci

exit $?
`

	if err := os.WriteFile(prePushPath, []byte(prePushScript), 0755); err != nil {
		return fmt.Errorf("failed to write pre-push hook: %w", err)
	}

	fmt.Println("✅ Installed pre-push hook")
	fmt.Println("\n💡 Hooks are now active. They will run automatically on commit and push.")

	return nil
}

// runHook executes a specific hook (called by git)
func runHook(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sentinel hook <type>")
	}

	hookType := args[0]

	fmt.Printf("🔍 Running %s hook...\n", hookType)

	// For now, just run a basic audit
	// In a full implementation, this would be configurable
	return runAudit([]string{"--ci"})
}
