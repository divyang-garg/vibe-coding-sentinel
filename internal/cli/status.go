// Package cli provides status command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runStatus executes the status command
func runStatus(args []string) error {
	fmt.Println("📊 Sentinel Project Status")
	fmt.Println("==========================")

	// Check if .sentinelrc exists
	if _, err := os.Stat(".sentinelrc"); err == nil {
		fmt.Println("✅ Configuration: .sentinelrc found")
	} else {
		fmt.Println("⚠️  Configuration: .sentinelrc not found (run 'sentinel init')")
	}

	// Check for common project files
	projectIndicators := []string{
		"package.json", "requirements.txt", "go.mod", "Cargo.toml", "pom.xml",
		"build.gradle", "Makefile", "Dockerfile", "README.md",
	}

	foundIndicators := 0
	for _, indicator := range projectIndicators {
		if _, err := os.Stat(indicator); err == nil {
			foundIndicators++
		}
	}

	if foundIndicators > 0 {
		fmt.Printf("✅ Project Files: %d/%d indicators found\n", foundIndicators, len(projectIndicators))
	} else {
		fmt.Println("⚠️  Project Files: No common project files detected")
	}

	// Check for git repository
	if _, err := os.Stat(".git"); err == nil {
		fmt.Println("✅ Version Control: Git repository detected")

		// Try to get basic git info
		if output, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
			changes := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(changes) > 0 && changes[0] != "" {
				fmt.Printf("⚠️  Working Directory: %d uncommitted changes\n", len(changes))
			} else {
				fmt.Println("✅ Working Directory: Clean")
			}
		}
	} else {
		fmt.Println("⚠️  Version Control: Not a git repository")
	}

	// Check for test files
	testFiles := 0
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.Contains(path, "/node_modules/") || strings.Contains(path, "/.git/") {
			return nil
		}
		if strings.Contains(info.Name(), "test") || strings.Contains(info.Name(), "spec") ||
			strings.HasSuffix(info.Name(), "_test.go") || strings.HasSuffix(info.Name(), ".test.js") {
			testFiles++
		}
		return nil
	})

	if testFiles > 0 {
		fmt.Printf("✅ Testing: %d test files detected\n", testFiles)
	} else {
		fmt.Println("⚠️  Testing: No test files detected")
	}

	// Check for Sentinel patterns
	if _, err := os.Stat(".sentinel/patterns.json"); err == nil {
		fmt.Println("✅ Patterns: Learned patterns found")
	} else {
		fmt.Println("⚠️  Patterns: No learned patterns (run 'sentinel learn')")
	}

	// Check for Cursor rules
	if _, err := os.Stat(".cursor/rules"); err == nil {
		fmt.Println("✅ Cursor Rules: Rules directory found")
	} else {
		fmt.Println("⚠️  Cursor Rules: Rules directory not found (run 'sentinel init')")
	}

	fmt.Println("")
	fmt.Println("💡 Recommendations:")
	if foundIndicators == 0 {
		fmt.Println("  - Run 'sentinel init' to bootstrap the project")
	}
	if testFiles == 0 {
		fmt.Println("  - Consider adding tests to improve code quality")
	}
	fmt.Println("  - Run 'sentinel audit' to check for security issues")
	if _, err := os.Stat(".sentinel/patterns.json"); os.IsNotExist(err) {
		fmt.Println("  - Run 'sentinel learn' to learn codebase patterns")
	}

	return nil
}
