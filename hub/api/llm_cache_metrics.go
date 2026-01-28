// Package main - LLM cache metrics tracking
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
// This file provides wrapper functions that delegate to the shared pkg package
package main

import (
	"sentinel-hub-api/pkg"
)

// recordCacheHit records a cache hit for metrics
// Delegates to shared pkg package
func recordCacheHit(projectID string) {
	pkg.RecordCacheHit(projectID)
}

// recordCacheMiss records a cache miss for metrics
// Delegates to shared pkg package
func recordCacheMiss(projectID string) {
	pkg.RecordCacheMiss(projectID)
}

// getCacheHitRate calculates cache hit rate for a project
// Delegates to shared pkg package
func getCacheHitRate(projectID string) float64 {
	return pkg.GetCacheHitRate(projectID)
}

// trackModelSelectionSavings tracks savings from model selection decisions
// Delegates to shared pkg package
func trackModelSelectionSavings(projectID string, savings float64, isCheaperModel bool) {
	pkg.TrackModelSelectionSavings(projectID, savings, isCheaperModel)
}

// getModelSelectionSavings returns the total savings from model selection for a project
// Delegates to shared pkg package
func getModelSelectionSavings(projectID string) float64 {
	return pkg.GetModelSelectionSavings(projectID)
}
