// Package cli provides MCP server command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"
	"os"

	"github.com/divyang-garg/sentinel-hub-api/internal/mcp"
)

// runMCPServer starts the MCP server
func runMCPServer() error {
	server := mcp.NewServer()
	fmt.Fprintf(os.Stderr, "MCP server starting...\n")
	return server.Start()
}
