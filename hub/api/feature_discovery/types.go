// Package feature_discovery provides feature and framework detection
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package feature_discovery

// FeatureDiscovery represents discovered framework information
type FeatureDiscovery struct {
	UIFramework    string            `json:"ui_framework"`
	UIFrameworkVer string            `json:"ui_framework_version,omitempty"`
	APIFramework   string            `json:"api_framework"`
	DatabaseORM    string            `json:"database_orm"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// AnalysisContext stores framework and discovery context
type AnalysisContext struct {
	FrameworkInfo *FeatureDiscovery `json:"framework_info"`
}
