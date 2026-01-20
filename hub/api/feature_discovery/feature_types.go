// Package feature_discovery provides discovered feature type definitions
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package feature_discovery

// DiscoveredFeature represents a feature discovered across all layers
type DiscoveredFeature struct {
	UILayer          *UILayerComponents   `json:"ui_layer,omitempty"`
	APILayer         *APILayerEndpoints   `json:"api_layer,omitempty"`
	DatabaseLayer    *DatabaseLayerTables `json:"database_layer,omitempty"`
	LogicLayer       *LogicLayer          `json:"logic_layer,omitempty"`
	IntegrationLayer *IntegrationLayer    `json:"integration_layer,omitempty"`
	TestLayer        *TestLayer           `json:"test_layer,omitempty"`
}

// LogicLayer represents business logic functions
type LogicLayer struct {
	Functions []BusinessLogicFunctionInfo `json:"functions"`
}

// BusinessLogicFunctionInfo contains information about a business logic function
type BusinessLogicFunctionInfo struct {
	Name       string `json:"name"`
	File       string `json:"file"`
	LineNumber int    `json:"line_number"`
}

// IntegrationLayer represents external API integrations
type IntegrationLayer struct {
	Integrations []IntegrationInfo `json:"integrations"`
}

// IntegrationInfo contains information about an external API integration
type IntegrationInfo struct {
	Name       string `json:"name"`
	Method     string `json:"method"`
	Endpoint   string `json:"endpoint"`
	File       string `json:"file"`
	LineNumber int    `json:"line_number"`
}

// TestLayer represents test files and test cases
type TestLayer struct {
	TestFiles []TestFileInfo `json:"test_files"`
}

// TestFileInfo contains information about a test file
type TestFileInfo struct {
	Path   string   `json:"path"`
	Suites []string `json:"suites,omitempty"`
}
