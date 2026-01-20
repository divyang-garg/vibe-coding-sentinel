// Package cli provides fix command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"

	"github.com/divyang-garg/sentinel-hub-api/internal/fix"
)

// runFix executes the fix command
func runFix(args []string) error {
	// Check for rollback command
	if len(args) > 0 && args[0] == "rollback" {
		rollbackOpts := fix.RollbackOptions{
			ListOnly: false,
		}

		// Parse rollback arguments
		for i, arg := range args[1:] {
			if arg == "--list" {
				rollbackOpts.ListOnly = true
			} else if arg == "--session" && i+1 < len(args) {
				rollbackOpts.SessionID = args[i+2]
			} else if len(arg) > 0 && arg[0] != '-' {
				rollbackOpts.SessionID = arg
			}
		}

		return fix.Rollback(rollbackOpts)
	}

	opts := fix.FixOptions{
		TargetPath: ".",
		DryRun:     false,
		Force:      false,
		Pattern:    "",
	}

	// Parse flags
	for i, arg := range args {
		switch arg {
		case "--help", "-h":
			return printFixHelp()
		case "--dry-run", "--safe":
			opts.DryRun = true
		case "--yes", "-y":
			opts.Force = true
			opts.DryRun = false
		case "--pattern":
			if i+1 < len(args) {
				opts.Pattern = args[i+1]
			}
		default:
			// Treat as target path if it doesn't start with --
			if len(arg) > 0 && arg[0] != '-' && opts.Pattern == "" {
				opts.TargetPath = arg
			}
		}
	}

	fmt.Println("🔧 Sentinel Auto-Fix")
	fmt.Println("====================")

	if opts.DryRun {
		fmt.Println("🔍 Dry-run mode: No files will be modified")
	}

	if opts.Pattern != "" {
		fmt.Printf("🎯 Applying pattern: %s\n", opts.Pattern)
	}

	// Perform fixes
	result, err := fix.Fix(opts)
	if err != nil {
		return fmt.Errorf("fix failed: %w", err)
	}

	// Display results
	fmt.Printf("\n✅ Auto-fix complete! Applied %d fixes.\n", result.FixesApplied)
	if result.FilesModified > 0 {
		fmt.Printf("📝 Modified %d files\n", result.FilesModified)
	}
	if result.BackupsCreated > 0 {
		fmt.Printf("💾 Created %d backups in .sentinel/backups/\n", result.BackupsCreated)
		fmt.Println("📝 Fix history saved to .sentinel/fix-history.json")
	}

	if opts.DryRun && result.FixesApplied > 0 {
		fmt.Println("\n💡 Run without --dry-run to apply these fixes")
	}

	return nil
}

// printFixHelp displays help for the fix command
func printFixHelp() error {
	fmt.Println("Usage: sentinel fix [options] [path]")
	fmt.Println("")
	fmt.Println("Apply safe automatic fixes to your codebase.")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --dry-run, --safe   Preview fixes without applying them")
	fmt.Println("  --yes, -y           Apply all fixes without prompting")
	fmt.Println("  --pattern <name>    Apply specific fix pattern only (console, debugger, imports, whitespace)")
	fmt.Println("  rollback            Restore files from last fix session")
	fmt.Println("  rollback --list     List available rollback sessions")
	fmt.Println("  --help, -h          Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  sentinel fix --dry-run              # Preview fixes")
	fmt.Println("  sentinel fix --yes                  # Apply all fixes")
	fmt.Println("  sentinel fix --pattern console      # Only fix console.log")
	fmt.Println("  sentinel fix src/                   # Fix only src/ directory")
	fmt.Println("  sentinel fix rollback               # Undo last fix")
	fmt.Println("  sentinel fix rollback --list        # List rollback sessions")
	return nil
}
