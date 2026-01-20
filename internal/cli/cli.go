// Package cli provides the command-line interface for Sentinel
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import "fmt"

// Execute processes command-line arguments and routes to appropriate handlers
func Execute(args []string) error {
	if len(args) == 0 {
		return printHelp()
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "audit":
		return runAudit(args[1:])
	case "learn":
		return runLearn(args[1:])
	case "fix":
		return runFix(args[1:])
	case "status":
		return runStatus(args[1:])
	case "baseline":
		return runBaseline(args[1:])
	case "history":
		return runHistory(args[1:])
	case "docs":
		return runDocs(args[1:])
	case "install-hooks":
		return runInstallHooks(args[1:])
	case "hook":
		return runHook(args[1:])
	case "validate-rules":
		return runValidateRules(args[1:])
	case "update-rules":
		return runUpdateRules(args[1:])
	case "knowledge":
		return runKnowledge(args[1:])
	case "review":
		return runReview(args[1:])
	case "doc-sync":
		return runDocSync(args[1:])
	case "mcp-server":
		return runMCPServer()
	case "version", "--version", "-v":
		return runVersion()
	case "help", "--help", "-h":
		return printHelp()
	default:
		return fmt.Errorf("unknown command: %s\n\nRun 'sentinel help' for usage", args[0])
	}
}

// printHelp displays usage information
func printHelp() error {
	fmt.Println("Sentinel - Vibe Coding Detection Tool")
	fmt.Println("Usage: sentinel <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init         Initialize Sentinel in current project")
	fmt.Println("  audit        Run security and quality audit")
	fmt.Println("  learn        Learn patterns from codebase")
	fmt.Println("  fix          Apply safe automatic fixes")
	fmt.Println("  status       Show project health status")
	fmt.Println("  baseline     Manage accepted findings")
	fmt.Println("  history      View audit history and trends")
	fmt.Println("  docs         Generate file structure documentation")
	fmt.Println("  install-hooks Install git hooks")
	fmt.Println("  validate-rules Validate Cursor rules syntax")
	fmt.Println("  update-rules Update rules from Hub")
	fmt.Println("  knowledge   Manage knowledge base")
	fmt.Println("  review      Review pending knowledge entries")
	fmt.Println("  doc-sync    Check documentation-code sync")
	fmt.Println("  mcp-server   Start MCP server for Cursor integration")
	fmt.Println("  version      Show version information")
	fmt.Println("  help         Show this help message")
	fmt.Println("")
	fmt.Println("For more information, see: https://github.com/divyang-garg/sentinel-hub-api")
	return nil
}

// runVersion displays version information
func runVersion() error {
	fmt.Println("sentinel v24")
	return nil
}

// runInit is implemented in init.go

// runAudit is implemented in audit.go

// runLearn is implemented in learn.go

// runFix is implemented in fix.go

// runStatus is implemented in status.go

// runMCPServer is implemented in mcp.go
