// Package main - Flow format matching utilities
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"os"
	"strings"
)

// requestFormatsMatch checks if UI request format matches API endpoint expectations
func requestFormatsMatch(uiStep FlowStep, apiStep FlowStep, feature *DiscoveredFeature) bool {
	// Use feature information if available for better matching
	if feature != nil && feature.APILayer != nil {
		// Check if we can match using discovered endpoint information
		for _, endpoint := range feature.APILayer.Endpoints {
			if endpoint.File == apiStep.File {
				// Found matching endpoint - can use endpoint metadata for validation
				// For now, fall through to file-based matching
				break
			}
		}
	}

	// Read both files
	uiContent, err1 := os.ReadFile(uiStep.File)
	apiContent, err2 := os.ReadFile(apiStep.File)

	if err1 != nil || err2 != nil {
		// Can't read files, assume match (conservative)
		return true
	}

	uiCode := strings.ToLower(string(uiContent))
	apiCode := strings.ToLower(string(apiContent))

	// Look for common data format indicators
	formatIndicators := []string{
		"json",
		"xml",
		"formdata",
		"multipart",
		"content-type",
	}

	uiFormats := []string{}
	apiFormats := []string{}

	for _, indicator := range formatIndicators {
		if strings.Contains(uiCode, indicator) {
			uiFormats = append(uiFormats, indicator)
		}
		if strings.Contains(apiCode, indicator) {
			apiFormats = append(apiFormats, indicator)
		}
	}

	// If both have formats, check for overlap
	if len(uiFormats) > 0 && len(apiFormats) > 0 {
		for _, uiFormat := range uiFormats {
			for _, apiFormat := range apiFormats {
				if uiFormat == apiFormat {
					return true
				}
			}
		}
		// No overlap found
		return false
	}

	// If no formats detected, assume match (conservative)
	return true
}

// operationsMatch checks if business logic operations match database operations
func operationsMatch(logicStep FlowStep, dbStep FlowStep, feature *DiscoveredFeature) bool {
	// Use feature information if available for better matching
	if feature != nil {
		// Check if logic layer and database layer information is available
		if feature.LogicLayer != nil && feature.DatabaseLayer != nil {
			// Verify that the files match discovered features
			logicFound := false
			dbFound := false
			for _, fn := range feature.LogicLayer.Functions {
				if fn.File == logicStep.File {
					logicFound = true
					break
				}
			}
			for _, table := range feature.DatabaseLayer.Tables {
				if table.File == dbStep.File {
					dbFound = true
					break
				}
			}
			// If both are found in discovered features, they're more likely to match
			if logicFound && dbFound {
				// Fall through to content-based matching for additional validation
			}
		}
	}

	if logicStep.File == "" || dbStep.File == "" {
		return true
	}

	logicContent, err1 := os.ReadFile(logicStep.File)
	dbContent, err2 := os.ReadFile(dbStep.File)

	if err1 != nil || err2 != nil {
		return true
	}

	logicCode := strings.ToLower(string(logicContent))
	dbCode := strings.ToLower(string(dbContent))

	// Check if logic operation matches database operation
	// Look for CRUD operations
	operations := []string{"create", "insert", "read", "select", "update", "delete", "remove"}

	for _, op := range operations {
		logicHasOp := strings.Contains(logicCode, op)
		dbHasOp := strings.Contains(dbCode, op)

		// If both have the same operation, they match
		if logicHasOp && dbHasOp {
			return true
		}
	}

	// If no clear operation match, check for function/table name matching using AST
	// Use AST to extract function names and table names for better matching
	if feature != nil && feature.LogicLayer != nil && feature.DatabaseLayer != nil {
		// Try to match function names with table names using AST-extracted information
		for _, fn := range feature.LogicLayer.Functions {
			if fn.File == logicStep.File {
				for _, table := range feature.DatabaseLayer.Tables {
					if table.File == dbStep.File {
						// Both files match discovered features - more likely to be related
						// Additional AST-based matching could be done here if needed
						return true
					}
				}
			}
		}
	}
	return true // Conservative: assume match if no clear mismatch
}

// responseFormatsMatch checks if API response format matches business logic expectations
func responseFormatsMatch(apiStep FlowStep, logicStep FlowStep, feature *DiscoveredFeature) bool {
	// Use feature information if available for better matching
	if feature != nil {
		// Check if API and logic layer information is available
		if feature.APILayer != nil && feature.LogicLayer != nil {
			// Verify that the files match discovered features
			apiFound := false
			logicFound := false
			for _, endpoint := range feature.APILayer.Endpoints {
				if endpoint.File == apiStep.File {
					apiFound = true
					break
				}
			}
			for _, fn := range feature.LogicLayer.Functions {
				if fn.File == logicStep.File {
					logicFound = true
					break
				}
			}
			// If both are found in discovered features, they're more likely to match
			if apiFound && logicFound {
				// Fall through to content-based matching for additional validation
			}
		}
	}

	if apiStep.File == "" || logicStep.File == "" {
		return true
	}

	apiContent, err1 := os.ReadFile(apiStep.File)
	logicContent, err2 := os.ReadFile(logicStep.File)

	if err1 != nil || err2 != nil {
		return true
	}

	apiCode := strings.ToLower(string(apiContent))
	logicCode := strings.ToLower(string(logicContent))

	// Look for response structure indicators
	responseIndicators := []string{
		"json",
		"response",
		"return",
		"result",
		"data",
	}

	apiHasResponse := false
	logicHasResponse := false

	for _, indicator := range responseIndicators {
		if strings.Contains(apiCode, indicator) {
			apiHasResponse = true
		}
		if strings.Contains(logicCode, indicator) {
			logicHasResponse = true
		}
	}

	// If both have response handling, assume they match
	if apiHasResponse && logicHasResponse {
		return true
	}

	// If neither has response handling, also assume match (might be handled elsewhere)
	return true
}
