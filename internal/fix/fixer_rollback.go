// Package fix provides rollback functionality for fix operations
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package fix

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/divyang-garg/sentinel-hub-api/internal/config"
)

// recordFixHistory records fix history to a JSON file
func recordFixHistory(result *Result, targetPath string) error {
	// Generate session ID
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	timestamp := time.Now().Format(time.RFC3339)

	history := map[string]interface{}{
		"session_id":      sessionID,
		"timestamp":       timestamp,
		"fixes_applied":   result.FixesApplied,
		"files_modified":  result.FilesModified,
		"backups_created": result.BackupsCreated,
		"target_path":     targetPath,
	}

	historyFile := ".sentinel/fix-history.json"
	if err := os.MkdirAll(".sentinel", 0755); err != nil {
		return err
	}

	// Read existing history (try array first, then single object)
	var historyList []map[string]interface{}
	if existingData, err := os.ReadFile(historyFile); err == nil {
		// Try to unmarshal as array
		if err := json.Unmarshal(existingData, &historyList); err != nil {
			// Try as single object
			var singleEntry map[string]interface{}
			if err2 := json.Unmarshal(existingData, &singleEntry); err2 == nil {
				historyList = []map[string]interface{}{singleEntry}
			}
		}
	}

	// Append to history list
	historyList = append(historyList, history)

	// Write back
	historyJSON, err := json.MarshalIndent(historyList, "", "  ")
	if err != nil {
		return err
	}

	return config.WriteFile(historyFile, string(historyJSON))
}

// RollbackOptions configures rollback behavior
type RollbackOptions struct {
	SessionID string
	ListOnly  bool
}

// Rollback restores files from a previous fix session
func Rollback(opts RollbackOptions) error {
	backupDir := ".sentinel/backups"
	historyFile := ".sentinel/fix-history.json"

	if opts.ListOnly {
		// List available rollback sessions
		if _, err := os.Stat(historyFile); os.IsNotExist(err) {
			fmt.Println("No fix history found")
			return nil
		}

		data, err := os.ReadFile(historyFile)
		if err != nil {
			return fmt.Errorf("failed to read history: %w", err)
		}

		historyList, err := parseHistoryFile(data)
		if err != nil {
			return fmt.Errorf("failed to parse history: %w", err)
		}

		fmt.Println("Available rollback sessions:")
		for i, entry := range historyList {
			fmt.Printf("  %d. Session: %s | Date: %v | Files: %v\n",
				i+1, entry["session_id"], entry["timestamp"], entry["files_modified"])
		}
		return nil
	}

	// Read history to find session
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		return fmt.Errorf("no fix history found")
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return fmt.Errorf("failed to read history: %w", err)
	}

	historyList, err := parseHistoryFile(data)
	if err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}

	// Find session
	var session map[string]interface{}
	if opts.SessionID != "" {
		for _, entry := range historyList {
			if entry["session_id"] == opts.SessionID {
				session = entry
				break
			}
		}
	} else {
		// Use most recent session
		if len(historyList) > 0 {
			session = historyList[len(historyList)-1]
		}
	}

	if session == nil {
		return fmt.Errorf("session not found")
	}

	sessionID, _ := session["session_id"].(string)
	if sessionID == "" {
		sessionID = "unknown"
	}

	// Find backup files for this session
	var restoredCount int
	err = filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Check if file matches session
		if !strings.Contains(path, sessionID) && sessionID != "unknown" {
			return nil
		}

		// Read backup file
		backupContent, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip if can't read
		}

		// Determine original file path
		originalPath := strings.TrimPrefix(path, backupDir+"/")
		originalPath = strings.TrimSuffix(originalPath, ".backup")

		// Restore file
		if err := os.WriteFile(originalPath, backupContent, 0644); err != nil {
			fmt.Printf("Failed to restore %s: %v\n", originalPath, err)
			return nil
		}

		restoredCount++
		fmt.Printf("Restored: %s\n", originalPath)
		return nil
	})

	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if restoredCount > 0 {
		fmt.Printf("✅ Rollback complete: restored %d files from session %s\n", restoredCount, sessionID)
	} else {
		return fmt.Errorf("no backup files found for session %s", sessionID)
	}

	return nil
}

// parseHistoryFile parses history file supporting both array and single object formats
// for backward compatibility with older history files
func parseHistoryFile(data []byte) ([]map[string]interface{}, error) {
	// Try array format first (new format)
	var historyList []map[string]interface{}
	if err := json.Unmarshal(data, &historyList); err == nil {
		return historyList, nil
	}

	// Try single object format (old format - backward compatibility)
	var singleEntry map[string]interface{}
	if err := json.Unmarshal(data, &singleEntry); err == nil {
		return []map[string]interface{}{singleEntry}, nil
	}

	return nil, fmt.Errorf("invalid history file format")
}
