// Phase 14E: Task Verification Engine - Integration Verification
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/smacker/go-tree-sitter" // Reserved for tree-sitter integration
)

// verifyIntegration verifies if integration requirements are met
func verifyIntegration(ctx context.Context, task *Task, codebasePath string) (TaskVerification, error) {
	verification := TaskVerification{
		TaskID:           task.ID,
		VerificationType: "integration",
		Status:           "pending",
		Confidence:       0.0,
		Evidence:         make(map[string]interface{}),
	}

	// Check for integration keywords
	integrationKeywords := []string{"api", "integration", "service", "external", "third-party", "sdk"}
	taskText := strings.ToLower(task.Title + " " + task.Description)

	hasIntegrationKeyword := false
	for _, keyword := range integrationKeywords {
		if strings.Contains(taskText, keyword) {
			hasIntegrationKeyword = true
			break
		}
	}

	if !hasIntegrationKeyword {
		// No integration requirement, skip
		verification.Confidence = 1.0
		verification.Status = "verified"
		verification.Evidence["skipped"] = true
		return verification, nil
	}

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)

	// Check for actual integration code
	integrationFiles, err := findIntegrationCode(ctx, codebasePath, keywords)
	if err != nil {
		LogError(ctx, "Failed to find integration code: %v", err)
		// Fallback to moderate confidence
		verification.Confidence = 0.5
		verification.Status = "pending"
		verification.Evidence["integration_required"] = true
		verification.Evidence["integration_files"] = []string{}
		return verification, nil
	}

	// Calculate confidence based on found integration patterns
	if len(integrationFiles) > 0 {
		verification.Confidence = 0.8
		verification.Status = "verified"
	} else {
		verification.Confidence = 0.3
		verification.Status = "pending"
	}

	verification.Evidence["integration_required"] = true
	verification.Evidence["integration_files"] = integrationFiles

	return verification, nil
}

// IntegrationEvidence contains detailed evidence of integration code found
type IntegrationEvidence struct {
	Files           []string `json:"files"`
	Functions       []string `json:"functions"`
	IntegrationType string   `json:"integration_type"` // "REST", "GraphQL", "gRPC", "WebSocket", "Middleware", "Event", "Unknown"
	ImportPaths     []string `json:"import_paths"`
	ConfigFiles     []string `json:"config_files"`
	ASTMatched      bool     `json:"ast_matched"`   // Whether AST found matches
	RegexMatched    bool     `json:"regex_matched"` // Whether regex found matches
}

// findIntegrationCode searches for actual integration code patterns using hybrid AST + regex approach
func findIntegrationCode(ctx context.Context, codebasePath string, keywords []string) ([]string, error) {
	var integrationFiles []string

	// Enhanced patterns for integration code
	integrationPatterns := []*regexp.Regexp{
		// HTTP clients
		regexp.MustCompile(`(?i)(http\.Client|resty|axios|fetch|requests\.|urllib|httpx)`),
		// API endpoints - fixed regex patterns with proper escaping
		regexp.MustCompile(`(?i)(api\.|endpoint|/api/|/v\d+/|\.(post|get|put|delete|patch)\()`),
		// Service clients
		regexp.MustCompile(`(?i)(client\.|service\.|sdk\.|(Client|Service)\(`),
		// External libraries
		regexp.MustCompile(`(?i)(import.*http|from.*requests|require.*axios|import.*fetch)`),
		// GraphQL patterns
		regexp.MustCompile(`(?i)(graphql|gql|apollo|relay|query|mutation|subscription|gql\()`),
		// gRPC patterns
		regexp.MustCompile(`(?i)(grpc|protobuf|\.proto|rpc|service.*pb|pb\.)`),
		// WebSocket patterns
		regexp.MustCompile(`(?i)(websocket|ws://|wss://|socket\.io|ws\.|WebSocket)`),
		// Middleware patterns
		regexp.MustCompile(`(?i)(middleware|use\(|app\.use|router\.use|express\.|gin\.|mux\.)`),
		// Event handler patterns
		regexp.MustCompile(`(?i)(on\(|addEventListener|emit\(|publish\(|subscribe|event\.|EventEmitter)`),
	}

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log error but continue processing
			LogWarn(ctx, "Error accessing path %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Skip test files
		if strings.Contains(path, "_test.") || strings.Contains(path, ".test.") {
			return nil
		}

		// Check file extension
		ext := filepath.Ext(path)
		supportedExts := map[string]bool{
			".go": true, ".js": true, ".ts": true, ".py": true,
			".java": true, ".jsx": true, ".tsx": true,
		}
		if !supportedExts[ext] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			LogWarn(ctx, "Failed to read file %s: %v", path, err)
			return nil // Continue with next file
		}

		contentStr := string(content)

		// Detect language from file extension
		language := detectLanguageFromExtension(ext)

		// Initialize evidence for this file
		evidence := &IntegrationEvidence{
			Files:           []string{},
			Functions:       []string{},
			IntegrationType: "Unknown",
			ImportPaths:     []string{},
			ConfigFiles:     []string{},
			ASTMatched:      false,
			RegexMatched:    false,
		}

		// Try AST-based analysis first for supported languages
		astMatched := false
		if language != "unknown" {
			funcs, imports, integrationTypes, err := findSymbolsWithAST(contentStr, language, keywords, path)
			if err != nil {
				// Log AST failure but continue with regex fallback
				LogDebug(ctx, "AST analysis failed for %s (language: %s): %v, falling back to regex", path, language, err)
			} else {
				// AST found matches
				if len(funcs) > 0 || len(imports) > 0 {
					astMatched = true
					evidence.ASTMatched = true
					evidence.Functions = funcs
					evidence.ImportPaths = imports
					if len(integrationTypes) > 0 {
						evidence.IntegrationType = strings.Join(integrationTypes, ", ")
					}
				}
			}
		}

		// Fallback to regex if AST didn't find anything or language not supported
		regexMatched := false
		detectedIntegrationType := "Unknown"

		if !astMatched {
			for _, pattern := range integrationPatterns {
				if pattern.MatchString(contentStr) {
					regexMatched = true
					evidence.RegexMatched = true

					// Detect integration type from pattern
					patternStr := pattern.String()
					if strings.Contains(patternStr, "graphql") || strings.Contains(patternStr, "gql") {
						detectedIntegrationType = "GraphQL"
					} else if strings.Contains(patternStr, "grpc") {
						detectedIntegrationType = "gRPC"
					} else if strings.Contains(patternStr, "websocket") || strings.Contains(patternStr, "socket") {
						detectedIntegrationType = "WebSocket"
					} else if strings.Contains(patternStr, "middleware") {
						detectedIntegrationType = "Middleware"
					} else if strings.Contains(patternStr, "event") || strings.Contains(patternStr, "emit") {
						detectedIntegrationType = "Event"
					} else if strings.Contains(patternStr, "http") || strings.Contains(patternStr, "api") {
						detectedIntegrationType = "REST"
					}
					break
				}
			}

			if detectedIntegrationType != "Unknown" && evidence.IntegrationType == "Unknown" {
				evidence.IntegrationType = detectedIntegrationType
			}
		}

		// Also check if keywords appear in file
		hasKeywords := false
		contentLower := strings.ToLower(contentStr)
		for _, keyword := range keywords {
			if strings.Contains(contentLower, strings.ToLower(keyword)) {
				hasKeywords = true
				break
			}
		}

		// File matches if (AST found matches OR regex found matches) AND keywords present
		if (astMatched || regexMatched) && hasKeywords {
			relativePath, err := filepath.Rel(codebasePath, path)
			if err != nil {
				relativePath = path // Fallback to full path
			}
			integrationFiles = append(integrationFiles, relativePath)
			evidence.Files = []string{relativePath}

			// Log successful match with details
			LogDebug(ctx, "Integration code found in %s: AST=%v, Regex=%v, Type=%s",
				relativePath, astMatched, regexMatched, evidence.IntegrationType)
		}

		return nil
	})

	if err != nil {
		LogError(ctx, "Error walking codebase: %v", err)
		return integrationFiles, fmt.Errorf("failed to scan codebase: %w", err)
	}

	return integrationFiles, nil
}
