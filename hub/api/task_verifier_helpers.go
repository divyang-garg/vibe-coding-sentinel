// Phase 14E: Task Verification Engine - Helper Functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// calculateOverallConfidence calculates overall confidence from individual verifications
func calculateOverallConfidence(verifications []TaskVerification, factors VerificationFactors) float64 {
	var totalConfidence float64
	var totalWeight float64

	for _, verification := range verifications {
		var weight float64
		switch verification.VerificationType {
		case "code_existence":
			weight = factors.CodeExistence
		case "code_usage":
			weight = factors.CodeUsage
		case "test_coverage":
			weight = factors.TestCoverage
		case "integration":
			weight = factors.Integration
		default:
			weight = 0.0
		}

		totalConfidence += verification.Confidence * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalConfidence / totalWeight
}

// determineTaskStatus determines task status based on confidence and verifications
func determineTaskStatus(confidence float64, verifications []TaskVerification) string {
	if confidence >= 0.8 {
		return "completed"
	} else if confidence >= 0.5 {
		return "in_progress"
	} else {
		return "pending"
	}
}

// buildEvidenceMap builds evidence map from verifications
func buildEvidenceMap(verifications []TaskVerification) map[string]interface{} {
	evidence := make(map[string]interface{})
	for _, verification := range verifications {
		evidence[verification.VerificationType] = map[string]interface{}{
			"status":     verification.Status,
			"confidence": verification.Confidence,
			"evidence":   verification.Evidence,
		}
	}
	return evidence
}

// storeVerification stores a verification result in the database
func storeVerification(ctx context.Context, verification TaskVerification) error {
	verificationID := uuid.New().String()
	now := time.Now()

	// Check if verification exists
	checkQuery := `SELECT id FROM task_verifications WHERE task_id = $1 AND verification_type = $2`
	var existingID string
	row := queryRowWithTimeout(ctx, checkQuery, verification.TaskID, verification.VerificationType)
	err := row.Scan(&existingID)

	// Initialize verifiedAt before using it in queries
	var verifiedAt *time.Time
	if verification.VerifiedAt != nil {
		verifiedAt = verification.VerifiedAt
	} else if verification.Status == "verified" {
		now := time.Now()
		verifiedAt = &now
	}

	var query string
	var args []interface{}

	if err == nil && existingID != "" {
		// Update existing
		query = `
			UPDATE task_verifications 
			SET status = $1, confidence = $2, evidence = $3, 
			    retry_count = retry_count + 1, verified_at = $4
			WHERE id = $5
		`
		evidenceJSON, _ := json.Marshal(verification.Evidence)
		args = []interface{}{
			verification.Status, verification.Confidence, string(evidenceJSON),
			verifiedAt, existingID,
		}
	} else {
		// Insert new
		query = `
			INSERT INTO task_verifications (
				id, task_id, verification_type, status, confidence, evidence, 
				retry_count, verified_at, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		evidenceJSON, _ := json.Marshal(verification.Evidence)
		args = []interface{}{
			verificationID, verification.TaskID, verification.VerificationType,
			verification.Status, verification.Confidence, string(evidenceJSON),
			verification.RetryCount, verifiedAt, now,
		}
	}

	_, err = execWithTimeout(ctx, query, args...)

	return err
}

// getCachedVerification retrieves cached verification results
func getCachedVerification(ctx context.Context, taskID string) (*VerifyTaskResponse, error) {
	query := `
		SELECT verification_type, status, confidence, evidence, verified_at
		FROM task_verifications
		WHERE task_id = $1
		ORDER BY created_at DESC
	`

	rows, err := queryWithTimeout(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifications: %w", err)
	}
	defer rows.Close()

	verifications := []TaskVerification{}
	var overallConfidence float64

	for rows.Next() {
		var verification TaskVerification
		var evidenceJSON string
		var verifiedAt sql.NullTime

		err := rows.Scan(
			&verification.VerificationType,
			&verification.Status,
			&verification.Confidence,
			&evidenceJSON,
			&verifiedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(evidenceJSON), &verification.Evidence)
		if verifiedAt.Valid {
			verification.VerifiedAt = &verifiedAt.Time
		}

		verifications = append(verifications, verification)
		overallConfidence += verification.Confidence
	}

	if len(verifications) > 0 {
		overallConfidence /= float64(len(verifications))
	}

	status := determineTaskStatus(overallConfidence, verifications)
	evidence := buildEvidenceMap(verifications)

	return &VerifyTaskResponse{
		TaskID:            taskID,
		OverallConfidence: overallConfidence,
		Verifications:     verifications,
		Status:            status,
		Evidence:          evidence,
	}, nil
}

// detectLanguageFromExtension detects language from file extension
func detectLanguageFromExtension(ext string) string {
	langMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".py":   "python",
		".java": "java",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "unknown"
}

// findSymbolsWithAST uses AST to find function/class definitions and imports matching keywords
// NOTE: AST parsing disabled - uses simple pattern matching fallback
func findSymbolsWithAST(code string, language string, keywords []string, filePath string) ([]string, []string, []string, error) {
	// AST parsing disabled - tree-sitter integration required
	// Use simple pattern matching fallback
	return findSymbolsWithPatterns(code, language, keywords, filePath)
}

// findSymbolsWithPatterns uses pattern matching to find symbols
func findSymbolsWithPatterns(code string, language string, keywords []string, filePath string) ([]string, []string, []string, error) {
	var matchedFunctions []string
	var matchedImports []string
	var integrationTypes []string

	// Build keyword map for fast lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	lines := strings.Split(code, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for function definitions
		var funcName string
		switch language {
		case "go":
			if strings.HasPrefix(trimmed, "func ") {
				parts := strings.Fields(trimmed)
				if len(parts) >= 2 {
					funcName = parts[1]
					if idx := strings.Index(funcName, "("); idx > 0 {
						funcName = funcName[:idx]
					}
				}
			}
		case "javascript", "typescript":
			if strings.HasPrefix(trimmed, "function ") {
				parts := strings.Fields(trimmed)
				if len(parts) >= 2 {
					funcName = parts[1]
					if idx := strings.Index(funcName, "("); idx > 0 {
						funcName = funcName[:idx]
					}
				}
			}
		case "python":
			if strings.HasPrefix(trimmed, "def ") {
				parts := strings.Fields(trimmed)
				if len(parts) >= 2 {
					funcName = parts[1]
					if idx := strings.Index(funcName, "("); idx > 0 {
						funcName = funcName[:idx]
					}
				}
			}
		}

		if funcName != "" {
			for kw := range keywordMap {
				if strings.Contains(strings.ToLower(funcName), kw) {
					matchedFunctions = appendIfNotExists(matchedFunctions, funcName)
					break
				}
			}
		}

		// Check for imports
		if strings.HasPrefix(trimmed, "import ") || strings.Contains(trimmed, "require(") {
			matchedImports = append(matchedImports, trimmed)
			// Detect integration type
			lineLower := strings.ToLower(trimmed)
			if strings.Contains(lineLower, "graphql") || strings.Contains(lineLower, "apollo") {
				integrationTypes = appendIfNotExists(integrationTypes, "GraphQL")
			}
			if strings.Contains(lineLower, "grpc") {
				integrationTypes = appendIfNotExists(integrationTypes, "gRPC")
			}
			if strings.Contains(lineLower, "websocket") || strings.Contains(lineLower, "socket") {
				integrationTypes = appendIfNotExists(integrationTypes, "WebSocket")
			}
			if strings.Contains(lineLower, "http") || strings.Contains(lineLower, "axios") || strings.Contains(lineLower, "fetch") || strings.Contains(lineLower, "requests") {
				integrationTypes = appendIfNotExists(integrationTypes, "REST")
			}
		}
	}

	return matchedFunctions, matchedImports, integrationTypes, nil
}
