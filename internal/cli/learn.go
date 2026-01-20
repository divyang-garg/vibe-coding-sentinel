// Package cli provides learn command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"strings"

	"github.com/divyang-garg/sentinel-hub-api/internal/config"
	"github.com/divyang-garg/sentinel-hub-api/internal/patterns"
)

// runLearn executes the learn command
func runLearn(args []string) error {
	// Get Hub configuration from environment
	hubURL, hubAPIKey := config.GetHubConfig()

	opts := patterns.LearnOptions{
		NamingOnly:           false,
		ImportsOnly:          false,
		StructureOnly:        false,
		CodebasePath:         ".",
		OutputJSON:           false,
		IncludeBusinessRules: false,
		HubURL:               hubURL,
		HubAPIKey:            hubAPIKey,
		ProjectID:            "",
	}

	// Parse flags
	for i, arg := range args {
		switch arg {
		case "--naming":
			opts.NamingOnly = true
		case "--imports":
			opts.ImportsOnly = true
		case "--structure":
			opts.StructureOnly = true
		case "--include-business-rules":
			opts.IncludeBusinessRules = true
		case "--output":
			if i+1 < len(args) && args[i+1] == "json" {
				opts.OutputJSON = true
			} else if strings.Contains(arg, "json") {
				opts.OutputJSON = true
			}
		case "--output=json":
			opts.OutputJSON = true
		case "--project-id":
			if i+1 < len(args) {
				opts.ProjectID = args[i+1]
			}
		case "--hub-url":
			if i+1 < len(args) {
				opts.HubURL = args[i+1]
			}
		case "--hub-api-key":
			if i+1 < len(args) {
				opts.HubAPIKey = args[i+1]
			}
		default:
			// Treat as codebase path if it doesn't start with --
			if len(arg) > 0 && arg[0] != '-' {
				opts.CodebasePath = arg
			}
		}
	}

	_, err := patterns.Learn(opts)
	return err
}
