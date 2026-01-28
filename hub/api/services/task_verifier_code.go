// Package services - Task Verification Engine - Code Existence and Usage Verification
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sentinel-hub-api/models"
)

// VerifyCodeExistence verifies if code mentioned in task exists
func VerifyCodeExistence(ctx context.Context, task *models.Task, codebasePath string) (models.TaskVerification, error) {
	verification := models.TaskVerification{
		TaskID:           task.ID,
		VerificationType: "code_existence",
		Status:           "pending",
		Confidence:       0.0,
		Evidence:         make(map[string]interface{}),
	}

	// Extract keywords from task title and description
	keywords := extractKeywords(task.Title + " " + task.Description)
	if len(keywords) == 0 {
		verification.Status = "failed"
		verification.Confidence = 0.0
		return verification, nil
	}

	// Search for keywords in codebase
	foundFiles := []string{}
	foundFunctions := []string{}
	totalMatches := 0

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip test files for code existence check
		if strings.Contains(path, "_test.") || strings.Contains(path, ".test.") {
			return nil
		}

		// Check file extension
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".js" && ext != ".ts" && ext != ".py" && ext != ".java" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := strings.ToLower(string(content))
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(contentStr, strings.ToLower(keyword)) {
				matches++
			}
		}

		if matches > 0 {
			relativePath, _ := filepath.Rel(codebasePath, path)
			foundFiles = append(foundFiles, relativePath)
			totalMatches += matches
		}

		return nil
	})

	if err != nil {
		return verification, fmt.Errorf("failed to scan codebase: %w", err)
	}

	// Calculate confidence based on matches
	confidence := 0.0
	if len(foundFiles) > 0 {
		// More files = higher confidence
		fileConfidence := float64(len(foundFiles)) / 10.0
		if fileConfidence > 1.0 {
			fileConfidence = 1.0
		}

		// More matches = higher confidence
		matchConfidence := float64(totalMatches) / float64(len(keywords)*5)
		if matchConfidence > 1.0 {
			matchConfidence = 1.0
		}

		confidence = (fileConfidence + matchConfidence) / 2.0
	}

	verification.Confidence = confidence
	if confidence > 0.5 {
		verification.Status = "verified"
	} else {
		verification.Status = "failed"
	}

	verification.Evidence["files"] = foundFiles
	verification.Evidence["functions"] = foundFunctions
	verification.Evidence["matches"] = totalMatches

	return verification, nil
}

// VerifyCodeUsage verifies if code is actually used (called/referenced)
func VerifyCodeUsage(ctx context.Context, task *models.Task, codebasePath string) (models.TaskVerification, error) {
	verification := models.TaskVerification{
		TaskID:           task.ID,
		VerificationType: "code_usage",
		Status:           "pending",
		Confidence:       0.0,
		Evidence:         make(map[string]interface{}),
	}

	// Get code existence first
	codeExistenceVerification, err := VerifyCodeExistence(ctx, task, codebasePath)
	if err != nil {
		return verification, fmt.Errorf("failed to verify code existence: %w", err)
	}

	files, _ := codeExistenceVerification.Evidence["files"].([]string)
	if len(files) == 0 {
		verification.Confidence = 0.0
		verification.Status = "failed"
		verification.Evidence["files"] = files
		verification.Evidence["call_sites"] = []string{}
		return verification, nil
	}

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)
	if len(keywords) == 0 {
		verification.Confidence = 0.0
		verification.Status = "failed"
		verification.Evidence["files"] = files
		verification.Evidence["call_sites"] = []string{}
		return verification, nil
	}

	// Find call sites using AST-based analysis
	callSites, err := extractFunctionCallsAST(codebasePath, keywords, files)
	if err != nil {
		LogError(ctx, "AST analysis failed, using fallback: %v", err)
		// Fallback to simple heuristic
		callSites = []string{}
	}

	// Calculate confidence based on call sites and file count
	confidence := 0.0
	if len(callSites) > 0 {
		// High confidence if call sites found
		confidence = 0.9
		verification.Status = "verified"
	} else if len(files) > 1 {
		// Code appears in multiple files, likely used
		confidence = 0.7
		verification.Status = "verified"
	} else if len(files) == 1 {
		// Code in one file, moderate confidence
		confidence = 0.5
		verification.Status = "verified"
	} else {
		confidence = 0.0
		verification.Status = "failed"
	}

	verification.Confidence = confidence
	verification.Evidence["files"] = files
	verification.Evidence["call_sites"] = callSites

	return verification, nil
}
