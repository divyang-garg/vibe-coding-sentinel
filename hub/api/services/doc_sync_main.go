// Package doc_sync_main - Main analysis functions for documentation synchronization
// Complies with CODING_STANDARDS.md: Handlers max 300 lines

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
)

// analyzeDocSync performs complete doc-sync analysis
func analyzeDocSync(ctx context.Context, req DocSyncRequest, codebasePath string) (DocSyncResponse, error) {
	roadmapPath := filepath.Join(codebasePath, "docs", "external", "IMPLEMENTATION_ROADMAP.md")

	// Parse status markers
	markers, err := parseStatusMarkers(roadmapPath)
	if err != nil {
		return DocSyncResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse status markers: %v", err),
		}, err
	}

	// Detect implementations for each phase
	var allDiscrepancies []Discrepancy
	for _, marker := range markers {
		evidence := detectImplementation(marker.Phase, codebasePath)
		discrepancies := compareStatus(marker, evidence)
		allDiscrepancies = append(allDiscrepancies, discrepancies...)
	}

	// Validate test coverage
	testCoverageDiscrepancies := validateTestCoverage(markers, filepath.Join(codebasePath, "tests"))
	allDiscrepancies = append(allDiscrepancies, testCoverageDiscrepancies...)

	// Validate feature flags, API endpoints, and commands if requested
	if req.ReportType == "all" || req.ReportType == "status_tracking" {
		featuresDocPath := filepath.Join(codebasePath, "docs", "external", "FEATURES.md")
		flagDiscrepancies := validateFeatureFlags(featuresDocPath, codebasePath)
		allDiscrepancies = append(allDiscrepancies, flagDiscrepancies...)

		mainGoPath := filepath.Join(codebasePath, "hub", "api", "main.go")
		endpointDiscrepancies := validateAPIEndpoints(roadmapPath, mainGoPath)
		allDiscrepancies = append(allDiscrepancies, endpointDiscrepancies...)

		agentPath := filepath.Join(codebasePath, "synapsevibsentinel.sh")
		commandDiscrepancies := validateCommands(featuresDocPath, agentPath)
		allDiscrepancies = append(allDiscrepancies, commandDiscrepancies...)
	}

	// Business rules comparison if requested
	if req.ReportType == "all" || req.ReportType == "business_rules" {
		businessRuleDiscrepancies, err := compareBusinessRules(ctx, req.ProjectID, codebasePath)
		if err != nil {
			log.Printf("Business rules comparison failed: %v", err)
		} else {
			allDiscrepancies = append(allDiscrepancies, businessRuleDiscrepancies...)
		}
	}

	// Generate report
	report := generateReport(markers, allDiscrepancies, req.ProjectID)

	// Store report in database
	reportID, err := storeDocSyncReport(ctx, report)
	if err != nil {
		log.Printf("Failed to store report: %v", err)
		reportID = report.ID
	}

	// Store suggested updates if fix mode is enabled
	var updateCount int
	if fixMode, ok := req.Options["fix"].(bool); ok && fixMode {
		updates := generateUpdateSuggestions(allDiscrepancies)
		if len(updates) > 0 {
			updateIDs, err := storeDocSyncUpdates(ctx, reportID, req.ProjectID, updates)
			if err != nil {
				log.Printf("Failed to store some updates: %v", err)
			}
			updateCount = len(updateIDs)
			log.Printf("Stored %d suggested updates for review", updateCount)
		}
	}

	return DocSyncResponse{
		Success:       true,
		ReportID:      reportID,
		InSync:        report.InSync,
		Discrepancies: report.Discrepancies,
		Summary:       report.Summary,
		Message:       fmt.Sprintf("Analyzed %d phases, found %d discrepancies, stored %d updates", len(markers), len(allDiscrepancies), updateCount),
	}, nil
}

// storeDocSyncReport stores report in database
func storeDocSyncReport(ctx context.Context, report DocSyncReport) (string, error) {
	discrepanciesJSON, err := json.Marshal(report.Discrepancies)
	if err != nil {
		return "", err
	}

	summaryJSON, err := json.Marshal(report.Summary)
	if err != nil {
		return "", err
	}

	query := `
		INSERT INTO doc_sync_reports (id, project_id, report_type, discrepancies, summary, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var reportID string
	err = db.QueryRowContext(ctx, query,
		report.ID,
		report.ProjectID,
		report.ReportType,
		string(discrepanciesJSON),
		string(summaryJSON),
		report.CreatedAt,
	).Scan(&reportID)

	if err != nil {
		return "", fmt.Errorf("failed to store report: %w", err)
	}

	return reportID, nil
}

// storeDocSyncUpdate stores a suggested update in the database
func storeDocSyncUpdate(ctx context.Context, reportID string, projectID string, update DocUpdate) (string, error) {
	query := `
		INSERT INTO doc_sync_updates (id, report_id, project_id, file_path, change_type, old_value, new_value, line_number, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id
	`

	var updateID string
	err := db.QueryRowContext(ctx, query,
		reportID,
		projectID,
		update.FilePath,
		update.ChangeType,
		update.OldValue,
		update.NewValue,
		update.LineNumber,
	).Scan(&updateID)

	if err != nil {
		return "", fmt.Errorf("failed to store update: %w", err)
	}

	return updateID, nil
}

// storeDocSyncUpdates stores multiple updates in the database
func storeDocSyncUpdates(ctx context.Context, reportID string, projectID string, updates []DocUpdate) ([]string, error) {
	var updateIDs []string
	for _, update := range updates {
		updateID, err := storeDocSyncUpdate(ctx, reportID, projectID, update)
		if err != nil {
			log.Printf("Failed to store update for %s: %v", update.FilePath, err)
			continue
		}
		updateIDs = append(updateIDs, updateID)
	}
	return updateIDs, nil
}
